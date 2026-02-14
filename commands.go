package main

import (
	"fmt"
	"os"
	"path/filepath"

	"bedolaga-installer/pkg/ui"
)

// ════════════════════════════════════════════════════════════════
// INSTALL WIZARD
// ════════════════════════════════════════════════════════════════

func installWizard() {
	ui.PrintBanner(appVersion)
	checkRoot()

	ui.PrintBox("📋 Перед началом",
		ui.InfoStyle.Render("Убедитесь, что у вас есть:")+"\n\n"+
			ui.HighlightStyle.Render("  1. ")+"BOT_TOKEN от @BotFather\n"+
			ui.HighlightStyle.Render("  2. ")+"Ваш Telegram ID (от @userinfobot)\n"+
			ui.HighlightStyle.Render("  3. ")+"REMNAWAVE_API_KEY из настроек панели\n"+
			ui.HighlightStyle.Render("  4. ")+"DNS-записи для доменов (опционально)")

	if !ui.ConfirmPrompt("Начать установку?", true) {
		os.Exit(0)
	}

	cfg := &Config{}

	// 1. System
	globalProgress.advance("Проверка системы")
	detectOS()

	// 2. Packages
	globalProgress.advance("Установка пакетов")
	updateSystem()
	installBasePackages()

	// 3. Docker
	globalProgress.advance("Настройка Docker")
	installDocker()

	// 4. Install dir
	globalProgress.advance("Каталог установки")
	selectInstallDir(cfg)

	// 5. Panel config
	globalProgress.advance("Конфигурация панели")
	checkRemnawavePanel(cfg)

	// 6. Check existing data
	globalProgress.advance("Проверка данных")
	checkPostgresVolume(cfg)

	// 7. Clone
	globalProgress.advance("Клонирование репозитория")
	cloneRepository(cfg)
	createDirectories(cfg)

	// 8. Interactive setup
	globalProgress.advance("Интерактивная настройка")
	interactiveSetup(cfg)

	// 9. Env file
	globalProgress.advance("Файл окружения")
	createEnvFile(cfg)

	// 10. Reverse proxy
	globalProgress.advance("Обратный прокси")
	switch cfg.ReverseProxyType {
	case "nginx_system":
		setupNginxSystem(cfg)
	case "nginx_panel":
		setupNginxPanel(cfg)
	case "caddy":
		setupCaddy(cfg)
	}
	setupSSL(cfg)

	// 11. Docker start
	globalProgress.advance("Docker-контейнеры")
	startDocker(cfg)
	setupFirewall()

	// 12. Finish
	globalProgress.advance("Завершение")
	createManagementScript(cfg)
	printFinalInfo(cfg)

	if ui.IsInteractive() {
		if ui.ConfirmPrompt("Показать логи бота?", false) {
			composeFile := "docker-compose.yml"
			if cfg.PanelInstalledLocally {
				composeFile = "docker-compose.local.yml"
			}
			allowExit = true
			// Используем exec для замены текущего процесса
			cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && docker compose -f %s logs --tail=150 -f bot", cfg.InstallDir, composeFile))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Run()
		}
	}
}

// ════════════════════════════════════════════════════════════════
// UPDATE / UNINSTALL (standalone commands)
// ════════════════════════════════════════════════════════════════

func findInstallDir() string {
	paths := []string{"/opt/remnawave-bedolaga-telegram-bot", "/root/remnawave-bedolaga-telegram-bot"}
	for _, p := range paths {
		if dirExists(p) {
			return p
		}
	}
	cwd, _ := os.Getwd()
	if fileExists(filepath.Join(cwd, "docker-compose.yml")) && fileExists(filepath.Join(cwd, ".env")) {
		return cwd
	}
	return ""
}

func detectComposeFile(installDir string) string {
	if fileExists(filepath.Join(installDir, "docker-compose.local.yml")) {
		return "docker-compose.local.yml"
	}
	return "docker-compose.yml"
}

func updateBot() {
	ui.PrintBanner(appVersion)
	installDir := findInstallDir()
	if installDir == "" {
		ui.PrintErrorBox(ui.ErrorStyle.Render("Установка бота не найдена!"))
		os.Exit(1)
	}
	composeFile := detectComposeFile(installDir)
	ui.PrintInfo("Каталог: " + installDir)

	if !ui.ConfirmPrompt("Начать обновление?", true) {
		os.Exit(0)
	}

	runShellSilent(fmt.Sprintf(`cd %s && cp .env ".env.backup_$(date +%%Y%%m%%d_%%H%%M%%S)" 2>/dev/null || true`, installDir))

	ui.RunWithSpinner("Загрузка последнего кода...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && git pull origin main", installDir))
		return err
	})

	ui.PrintInfo("Пересборка и перезапуск...")
	runShell(fmt.Sprintf("cd %s && docker compose -f %s down && docker compose -f %s up -d --build && docker compose -f %s logs -f -t", installDir, composeFile, composeFile, composeFile))
}

func uninstallBot() {
	ui.PrintBanner(appVersion)
	installDir := findInstallDir()
	if installDir == "" {
		ui.PrintErrorBox(ui.ErrorStyle.Render("Бот не установлен!"))
		os.Exit(1)
	}
	composeFile := detectComposeFile(installDir)
	ui.PrintInfo("Каталог: " + installDir)

	val := ui.InputText("Введите 'yes' для подтверждения удаления", "", "Это остановит и удалит контейнеры бота", true)
	if val != "yes" {
		ui.PrintSuccess("Отменено")
		return
	}

	if ui.ConfirmPrompt("Создать резервную копию сначала?", true) {
		runShellSilent(fmt.Sprintf(`cd %s && tar -czf "/root/bedolaga_backup_$(date +%%Y%%m%%d_%%H%%M%%S).tar.gz" .env data/ 2>/dev/null || true`, installDir))
		ui.PrintSuccess("Резервная копия сохранена в /root/")
	}

	ui.RunWithSpinner("Остановка контейнеров...", func() error {
		runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down -v 2>/dev/null || docker compose down -v 2>/dev/null || true", installDir, composeFile))
		return nil
	})

	runShellSilent("rm -f /etc/nginx/sites-enabled/bedolaga-webhook /etc/nginx/sites-enabled/bedolaga-miniapp")
	runShellSilent("rm -f /etc/nginx/sites-available/bedolaga-webhook /etc/nginx/sites-available/bedolaga-miniapp")
	runShellSilent("nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true")
	if fileExists("/etc/caddy/Caddyfile") {
		runShellSilent(`sed -i '/# === BEGIN Bedolaga Bot ===/,/# === END Bedolaga Bot ===/d' /etc/caddy/Caddyfile`)
		runShellSilent("systemctl reload caddy 2>/dev/null || true")
	}
	os.Remove("/usr/local/bin/bot")

	if ui.ConfirmPrompt("Удалить каталог "+installDir+"?", false) {
		os.RemoveAll(installDir)
	}

	ui.PrintSuccessBox(ui.SuccessStyle.Render("Удаление завершено!"))
}
