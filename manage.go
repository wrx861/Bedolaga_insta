package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bedolaga-installer/pkg/ui"
)

// ════════════════════════════════════════════════════════════════
// MANAGE BOT (TUI Management Panel)
// ════════════════════════════════════════════════════════════════

func manageBot() {
	// Help не требует установленного бота
	if len(os.Args) > 2 {
		subcmd := os.Args[2]
		if subcmd == "help" || subcmd == "--help" || subcmd == "-h" {
			printManageHelp()
			return
		}
	}

	installDir := findInstallDir()
	if installDir == "" {
		ui.PrintErrorBox(ui.ErrorStyle.Render("Установка бота не найдена!\n") +
			ui.DimStyle.Render("Ожидаемые пути:\n  /opt/remnawave-bedolaga-telegram-bot\n  /root/remnawave-bedolaga-telegram-bot"))
		os.Exit(1)
	}
	composeFile := detectComposeFile(installDir)

	// Прямые субкоманды: bot logs, bot status, etc.
	if len(os.Args) > 2 {
		manageSubcommand(os.Args[2], installDir, composeFile)
		return
	}

	// Интерактивное TUI-меню
	for {
		printManageHeader(installDir)

		idx := ui.SelectOption("Управление", []ui.SelectItem{
			{Title: "Логи", Description: "Просмотр логов в реальном времени (Ctrl+C — выход)"},
			{Title: "Статус", Description: "Контейнеры, порты и потребление ресурсов"},
			{Title: "Перезапуск", Description: "Перезапустить все контейнеры"},
			{Title: "Запуск", Description: "Запустить контейнеры"},
			{Title: "Остановка", Description: "Остановить все контейнеры"},
			{Title: "Обновление", Description: "git pull + пересборка Docker-образов"},
			{Title: "Бэкап", Description: "Резервная копия БД и конфигурации"},
			{Title: "Диагностика", Description: "Проверка работоспособности всех компонентов"},
			{Title: "Конфигурация", Description: "Открыть .env в редакторе"},
			{Title: "Удаление", Description: "Полное удаление бота и контейнеров"},
			{Title: "Выход", Description: "Закрыть панель управления"},
		})

		switch idx {
		case 0:
			manageLogs(installDir, composeFile)
		case 1:
			manageStatus(installDir, composeFile)
			waitForEnter()
		case 2:
			manageRestart(installDir, composeFile)
			waitForEnter()
		case 3:
			manageStart(installDir, composeFile)
			waitForEnter()
		case 4:
			manageStop(installDir, composeFile)
			waitForEnter()
		case 5:
			manageUpdate(installDir, composeFile)
			waitForEnter()
		case 6:
			manageBackup(installDir, composeFile)
			waitForEnter()
		case 7:
			manageHealth(installDir, composeFile)
			waitForEnter()
		case 8:
			manageConfig(installDir)
		case 9:
			manageUninstall(installDir, composeFile)
			return
		default:
			fmt.Println()
			fmt.Println(ui.DimStyle.Render("  Пока!"))
			return
		}
	}
}

// ════════════════════════════════════════════════════════════════
// HEADER & HELPERS
// ════════════════════════════════════════════════════════════════

func printManageHeader(installDir string) {
	fmt.Print("\033[2J\033[H") // Clear screen
	ui.PrintBanner(appVersion)

	fmt.Println(ui.DimStyle.Render("  Каталог:  ") + ui.InfoStyle.Render(installDir))

	out, _ := runShellSilent(`docker ps --format '{{.Names}}' | grep "^remnawave_bot$"`)
	if out != "" {
		fmt.Println(ui.DimStyle.Render("  Статус:   ") + ui.SuccessStyle.Render("● Работает"))
	} else {
		fmt.Println(ui.DimStyle.Render("  Статус:   ") + ui.ErrorStyle.Render("○ Остановлен"))
	}
	fmt.Println()
}

func waitForEnter() {
	fmt.Println()
	fmt.Print(ui.DimStyle.Render("  Нажмите Enter для продолжения..."))
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// ════════════════════════════════════════════════════════════════
// MANAGE: LOGS
// ════════════════════════════════════════════════════════════════

func manageLogs(installDir, composeFile string) {
	ui.PrintInfo("Логи бота (Ctrl+C для выхода)...")
	fmt.Println()
	allowExit = true
	runShell(fmt.Sprintf("cd %s && docker compose -f %s logs -f --tail=150 bot", installDir, composeFile))
	allowExit = false
}

// ════════════════════════════════════════════════════════════════
// MANAGE: STATUS
// ════════════════════════════════════════════════════════════════

func manageStatus(installDir, composeFile string) {
	sep := ui.DimStyle.Render("  ─────────────────────────────────────────────────────")

	fmt.Println()
	fmt.Println(ui.HighlightStyle.Render("  Контейнеры"))
	fmt.Println(sep)
	out, _ := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s ps --format 'table {{.Name}}\\t{{.Status}}\\t{{.Ports}}' 2>/dev/null", installDir, composeFile))
	if out != "" {
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(strings.ToLower(line), "up") || strings.Contains(line, "NAME") {
				fmt.Println("  " + line)
			} else {
				fmt.Println(ui.ErrorStyle.Render("  " + line))
			}
		}
	} else {
		fmt.Println(ui.DimStyle.Render("  Контейнеры не найдены"))
	}

	fmt.Println()
	fmt.Println(ui.HighlightStyle.Render("  Ресурсы"))
	fmt.Println(sep)
	out, _ = runShellSilent(`docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | grep -E "remnawave|postgres|redis|NAME"`)
	if out != "" {
		for _, line := range strings.Split(out, "\n") {
			fmt.Println("  " + line)
		}
	} else {
		fmt.Println(ui.DimStyle.Render("  Нет данных"))
	}

	fmt.Println()
	fmt.Println(ui.HighlightStyle.Render("  Диск"))
	fmt.Println(sep)
	out, _ = runShellSilent(`docker system df --format "table {{.Type}}\t{{.TotalCount}}\t{{.Size}}\t{{.Reclaimable}}" 2>/dev/null`)
	if out != "" {
		for _, line := range strings.Split(out, "\n") {
			fmt.Println("  " + line)
		}
	}
}

// ════════════════════════════════════════════════════════════════
// MANAGE: RESTART / START / STOP
// ════════════════════════════════════════════════════════════════

func manageRestart(installDir, composeFile string) {
	ui.RunWithSpinner("Перезапуск контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s restart", installDir, composeFile))
		return err
	})
	ui.PrintSuccess("Контейнеры перезапущены")
}

func manageStart(installDir, composeFile string) {
	ui.RunWithSpinner("Запуск контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s up -d", installDir, composeFile))
		return err
	})
	ui.PrintSuccess("Контейнеры запущены")
}

func manageStop(installDir, composeFile string) {
	if !ui.ConfirmPrompt("Остановить все контейнеры?", true) {
		return
	}
	ui.RunWithSpinner("Остановка контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down", installDir, composeFile))
		return err
	})
	ui.PrintSuccess("Контейнеры остановлены")
}

// ════════════════════════════════════════════════════════════════
// MANAGE: UPDATE
// ════════════════════════════════════════════════════════════════

func manageUpdate(installDir, composeFile string) {
	if !ui.ConfirmPrompt("Начать обновление бота?", true) {
		return
	}

	runShellSilent(fmt.Sprintf(`cd %s && cp .env ".env.backup_$(date +%%Y%%m%%d_%%H%%M%%S)" 2>/dev/null || true`, installDir))
	ui.PrintSuccess("Резервная копия .env создана")

	ui.RunWithSpinner("Загрузка обновлений...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && git pull origin main", installDir))
		return err
	})

	ui.RunWithSpinner("Пересборка контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down && docker compose -f %s up -d --build", installDir, composeFile, composeFile))
		return err
	})

	ui.PrintSuccess("Обновление завершено")
}

// ════════════════════════════════════════════════════════════════
// MANAGE: BACKUP
// ════════════════════════════════════════════════════════════════

func manageBackup(installDir, composeFile string) {
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(installDir, "data", "backups", timestamp)
	os.MkdirAll(backupDir, 0755)

	fmt.Println()

	ui.RunWithSpinner("Бэкап базы данных...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s exec -T postgres pg_dump -U remnawave_user remnawave_bot > %s/database.sql 2>/dev/null", installDir, composeFile, backupDir))
		return err
	})

	runShellSilent(fmt.Sprintf("cp %s/.env %s/.env 2>/dev/null", installDir, backupDir))
	runShellSilent(fmt.Sprintf("cp %s/docker-compose*.yml %s/ 2>/dev/null", installDir, backupDir))

	ui.PrintSuccess("Бэкап создан: " + backupDir)

	runShellSilent(fmt.Sprintf(`ls -dt "%s/data/backups"/*/ 2>/dev/null | tail -n +6 | xargs -r rm -rf`, installDir))

	out, _ := runShellSilent(fmt.Sprintf("du -sh %s 2>/dev/null | awk '{print $1}'", backupDir))
	if out != "" {
		ui.PrintInfo("Размер: " + out)
	}
}

// ════════════════════════════════════════════════════════════════
// MANAGE: HEALTH
// ════════════════════════════════════════════════════════════════

func manageHealth(installDir, composeFile string) {
	sep := ui.DimStyle.Render("  ─────────────────────────────────────────────────────")

	fmt.Println()
	fmt.Println(ui.AccentBar.Render("  ДИАГНОСТИКА СИСТЕМЫ"))
	fmt.Println(sep)
	fmt.Println()

	out, _ := runShellSilent(`docker ps --format '{{.Names}}' | grep "^remnawave_bot$"`)
	if out != "" {
		ui.PrintSuccess("Бот: работает")
	} else {
		ui.PrintError("Бот: не запущен")
	}

	_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s exec -T postgres pg_isready -U remnawave_user 2>/dev/null", installDir, composeFile))
	if err == nil {
		ui.PrintSuccess("PostgreSQL: работает")
	} else {
		ui.PrintError("PostgreSQL: не доступен")
	}

	_, err = runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s exec -T redis redis-cli ping 2>/dev/null", installDir, composeFile))
	if err == nil {
		ui.PrintSuccess("Redis: работает")
	} else {
		ui.PrintError("Redis: не доступен")
	}

	if commandExists("docker") {
		ui.PrintSuccess("Docker: установлен")
	} else {
		ui.PrintError("Docker: не найден")
	}

	out, _ = runShellSilent("df -h / | tail -1 | awk '{print $4}'")
	if out != "" {
		ui.PrintInfo("Свободно на диске: " + out)
	}

	out, _ = runShellSilent("free -h | grep Mem | awk '{printf \"%s / %s\", $3, $2}'")
	if out != "" {
		ui.PrintInfo("Память: " + out)
	}

	fmt.Println()
	fmt.Println(ui.DimStyle.Render("  Последние логи:"))
	fmt.Println(sep)
	runShell(fmt.Sprintf("cd %s && docker compose -f %s logs --tail=10 bot 2>/dev/null", installDir, composeFile))
}

// ════════════════════════════════════════════════════════════════
// MANAGE: CONFIG
// ════════════════════════════════════════════════════════════════

func manageConfig(installDir string) {
	envPath := filepath.Join(installDir, ".env")
	if !fileExists(envPath) {
		ui.PrintError("Файл .env не найден: " + envPath)
		return
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, envPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	fmt.Println()
	ui.PrintWarning("Перезапустите бота для применения изменений: bot restart")
	waitForEnter()
}

// ════════════════════════════════════════════════════════════════
// MANAGE: UNINSTALL
// ════════════════════════════════════════════════════════════════

func manageUninstall(installDir, composeFile string) {
	fmt.Println()
	fmt.Println(ui.ErrorStyle.Render("  ╔══════════════════════════════════╗"))
	fmt.Println(ui.ErrorStyle.Render("  ║         УДАЛЕНИЕ БОТА            ║"))
	fmt.Println(ui.ErrorStyle.Render("  ╚══════════════════════════════════╝"))
	fmt.Println()

	if !ui.ConfirmPrompt("Вы уверены, что хотите удалить бота?", false) {
		ui.PrintSuccess("Отменено")
		return
	}

	val := ui.InputText("Введите 'yes' для подтверждения", "", "Это действие нельзя отменить", true)
	if val != "yes" {
		ui.PrintSuccess("Отменено")
		return
	}

	if ui.ConfirmPrompt("Создать резервную копию перед удалением?", true) {
		runShellSilent(fmt.Sprintf(`cd %s && tar -czf "/root/bedolaga_backup_$(date +%%Y%%m%%d_%%H%%M%%S).tar.gz" .env data/ 2>/dev/null || true`, installDir))
		ui.PrintSuccess("Резервная копия сохранена в /root/")
	}

	ui.RunWithSpinner("Остановка контейнеров...", func() error {
		runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down -v 2>/dev/null || true", installDir, composeFile))
		return nil
	})

	runShellSilent("rm -f /etc/nginx/sites-enabled/bedolaga-* /etc/nginx/sites-available/bedolaga-*")
	runShellSilent("nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true")
	if fileExists("/etc/caddy/Caddyfile") {
		runShellSilent(`sed -i '/# === BEGIN Bedolaga Bot ===/,/# === END Bedolaga Bot ===/d' /etc/caddy/Caddyfile`)
		runShellSilent("systemctl reload caddy 2>/dev/null || true")
	}

	os.Remove("/usr/local/bin/bot")

	if ui.ConfirmPrompt("Удалить каталог "+installDir+"?", false) {
		os.RemoveAll(installDir)
		ui.PrintSuccess("Каталог удалён")
	} else {
		ui.PrintInfo("Каталог сохранён: " + installDir)
	}

	ui.PrintSuccessBox(ui.SuccessStyle.Render("Удаление завершено!"))
}

// ════════════════════════════════════════════════════════════════
// SUBCOMMAND ROUTER (for: bot logs, bot status, etc.)
// ════════════════════════════════════════════════════════════════

func manageSubcommand(subcmd, installDir, composeFile string) {
	switch subcmd {
	case "logs":
		manageLogs(installDir, composeFile)
	case "status":
		manageStatus(installDir, composeFile)
	case "restart":
		manageRestart(installDir, composeFile)
	case "start":
		manageStart(installDir, composeFile)
	case "stop":
		manageStop(installDir, composeFile)
	case "update", "upgrade":
		manageUpdate(installDir, composeFile)
	case "backup":
		manageBackup(installDir, composeFile)
	case "health", "check":
		manageHealth(installDir, composeFile)
	case "config", "edit":
		manageConfig(installDir)
	case "uninstall", "remove":
		manageUninstall(installDir, composeFile)
	case "help", "--help", "-h":
		printManageHelp()
	default:
		ui.PrintError("Неизвестная команда: " + subcmd)
		fmt.Println()
		printManageHelp()
		os.Exit(1)
	}
}

func printManageHelp() {
	ui.PrintBanner(appVersion)
	fmt.Println(ui.HighlightStyle.Render("  Использование:") + ui.DimStyle.Render(" bot [команда]"))
	fmt.Println()
	fmt.Println(ui.DimStyle.Render("  (без аргументов)  ") + "Интерактивное меню со стрелками")
	fmt.Println()
	fmt.Println(ui.InfoStyle.Render("  logs            ") + "  Просмотр логов")
	fmt.Println(ui.InfoStyle.Render("  status          ") + "  Статус контейнеров и ресурсы")
	fmt.Println(ui.InfoStyle.Render("  restart         ") + "  Перезапуск контейнеров")
	fmt.Println(ui.InfoStyle.Render("  start           ") + "  Запуск контейнеров")
	fmt.Println(ui.InfoStyle.Render("  stop            ") + "  Остановка контейнеров")
	fmt.Println(ui.InfoStyle.Render("  update          ") + "  Обновление (git pull + rebuild)")
	fmt.Println(ui.InfoStyle.Render("  backup          ") + "  Создать резервную копию")
	fmt.Println(ui.InfoStyle.Render("  health          ") + "  Диагностика системы")
	fmt.Println(ui.InfoStyle.Render("  config          ") + "  Редактировать .env")
	fmt.Println(ui.InfoStyle.Render("  uninstall       ") + "  Удалить бота")
	fmt.Println(ui.InfoStyle.Render("  help            ") + "  Эта справка")
	fmt.Println()
}
