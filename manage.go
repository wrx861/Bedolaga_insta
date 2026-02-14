package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ════════════════════════════════════════════════════════════════
// MANAGE BOT (TUI Management Panel)
// ════════════════════════════════════════════════════════════════

func manageBot() {
	installDir := findInstallDir()
	if installDir == "" {
		printErrorBox(errorStyle.Render("Установка бота не найдена!\n") +
			dimStyle.Render("Ожидаемые пути:\n  /opt/remnawave-bedolaga-telegram-bot\n  /root/remnawave-bedolaga-telegram-bot"))
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

		idx := selectOption("Управление", []selectItem{
			{title: "Логи", description: "Просмотр логов в реальном времени (Ctrl+C — выход)"},
			{title: "Статус", description: "Контейнеры, порты и потребление ресурсов"},
			{title: "Перезапуск", description: "Перезапустить все контейнеры"},
			{title: "Запуск", description: "Запустить контейнеры"},
			{title: "Остановка", description: "Остановить все контейнеры"},
			{title: "Обновление", description: "git pull + пересборка Docker-образов"},
			{title: "Бэкап", description: "Резервная копия БД и конфигурации"},
			{title: "Диагностика", description: "Проверка работоспособности всех компонентов"},
			{title: "Конфигурация", description: "Открыть .env в редакторе"},
			{title: "Удаление", description: "Полное удаление бота и контейнеров"},
			{title: "Выход", description: "Закрыть панель управления"},
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
			fmt.Println(dimStyle.Render("  Пока!"))
			return
		}
	}
}

// ════════════════════════════════════════════════════════════════
// HEADER & HELPERS
// ════════════════════════════════════════════════════════════════

func printManageHeader(installDir string) {
	fmt.Print("\033[2J\033[H") // Clear screen
	printBanner()

	fmt.Println(dimStyle.Render("  Каталог:  ") + infoStyle.Render(installDir))

	out, _ := runShellSilent(`docker ps --format '{{.Names}}' | grep "^remnawave_bot$"`)
	if out != "" {
		fmt.Println(dimStyle.Render("  Статус:   ") + successStyle.Render("● Работает"))
	} else {
		fmt.Println(dimStyle.Render("  Статус:   ") + errorStyle.Render("○ Остановлен"))
	}
	fmt.Println()
}

func waitForEnter() {
	fmt.Println()
	fmt.Print(dimStyle.Render("  Нажмите Enter для продолжения..."))
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// ════════════════════════════════════════════════════════════════
// MANAGE: LOGS
// ════════════════════════════════════════════════════════════════

func manageLogs(installDir, composeFile string) {
	printInfo("Логи бота (Ctrl+C для выхода)...")
	fmt.Println()
	allowExit = true
	runShell(fmt.Sprintf("cd %s && docker compose -f %s logs -f --tail=150 bot", installDir, composeFile))
	allowExit = false
}

// ════════════════════════════════════════════════════════════════
// MANAGE: STATUS
// ════════════════════════════════════════════════════════════════

func manageStatus(installDir, composeFile string) {
	sep := dimStyle.Render("  ─────────────────────────────────────────────────────")

	fmt.Println()
	fmt.Println(highlightStyle.Render("  Контейнеры"))
	fmt.Println(sep)
	out, _ := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s ps --format 'table {{.Name}}\\t{{.Status}}\\t{{.Ports}}' 2>/dev/null", installDir, composeFile))
	if out != "" {
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(strings.ToLower(line), "up") || strings.Contains(line, "NAME") {
				fmt.Println("  " + line)
			} else {
				fmt.Println(errorStyle.Render("  " + line))
			}
		}
	} else {
		fmt.Println(dimStyle.Render("  Контейнеры не найдены"))
	}

	fmt.Println()
	fmt.Println(highlightStyle.Render("  Ресурсы"))
	fmt.Println(sep)
	out, _ = runShellSilent(`docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | grep -E "remnawave|postgres|redis|NAME"`)
	if out != "" {
		for _, line := range strings.Split(out, "\n") {
			fmt.Println("  " + line)
		}
	} else {
		fmt.Println(dimStyle.Render("  Нет данных"))
	}

	fmt.Println()
	fmt.Println(highlightStyle.Render("  Диск"))
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
	runWithSpinner("Перезапуск контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s restart", installDir, composeFile))
		return err
	})
	printSuccess("Контейнеры перезапущены")
}

func manageStart(installDir, composeFile string) {
	runWithSpinner("Запуск контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s up -d", installDir, composeFile))
		return err
	})
	printSuccess("Контейнеры запущены")
}

func manageStop(installDir, composeFile string) {
	if !confirmPrompt("Остановить все контейнеры?", true) {
		return
	}
	runWithSpinner("Остановка контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down", installDir, composeFile))
		return err
	})
	printSuccess("Контейнеры остановлены")
}

// ════════════════════════════════════════════════════════════════
// MANAGE: UPDATE
// ════════════════════════════════════════════════════════════════

func manageUpdate(installDir, composeFile string) {
	if !confirmPrompt("Начать обновление бота?", true) {
		return
	}

	// Backup .env
	runShellSilent(fmt.Sprintf(`cd %s && cp .env ".env.backup_$(date +%%Y%%m%%d_%%H%%M%%S)" 2>/dev/null || true`, installDir))
	printSuccess("Резервная копия .env создана")

	// Git pull
	runWithSpinner("Загрузка обновлений...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && git pull origin main", installDir))
		return err
	})

	// Rebuild
	runWithSpinner("Пересборка контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down && docker compose -f %s up -d --build", installDir, composeFile, composeFile))
		return err
	})

	printSuccess("Обновление завершено")
}

// ════════════════════════════════════════════════════════════════
// MANAGE: BACKUP
// ════════════════════════════════════════════════════════════════

func manageBackup(installDir, composeFile string) {
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(installDir, "data", "backups", timestamp)
	os.MkdirAll(backupDir, 0755)

	fmt.Println()

	// Database
	runWithSpinner("Бэкап базы данных...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s exec -T postgres pg_dump -U remnawave_user remnawave_bot > %s/database.sql 2>/dev/null", installDir, composeFile, backupDir))
		return err
	})

	// Config files
	runShellSilent(fmt.Sprintf("cp %s/.env %s/.env 2>/dev/null", installDir, backupDir))
	runShellSilent(fmt.Sprintf("cp %s/docker-compose*.yml %s/ 2>/dev/null", installDir, backupDir))

	printSuccess("Бэкап создан: " + backupDir)

	// Cleanup old backups (keep 5)
	runShellSilent(fmt.Sprintf(`ls -dt "%s/data/backups"/*/ 2>/dev/null | tail -n +6 | xargs -r rm -rf`, installDir))

	// Show backup size
	out, _ := runShellSilent(fmt.Sprintf("du -sh %s 2>/dev/null | awk '{print $1}'", backupDir))
	if out != "" {
		printInfo("Размер: " + out)
	}
}

// ════════════════════════════════════════════════════════════════
// MANAGE: HEALTH
// ════════════════════════════════════════════════════════════════

func manageHealth(installDir, composeFile string) {
	sep := dimStyle.Render("  ─────────────────────────────────────────────────────")

	fmt.Println()
	fmt.Println(accentBar.Render("  ДИАГНОСТИКА СИСТЕМЫ"))
	fmt.Println(sep)
	fmt.Println()

	// Bot container
	out, _ := runShellSilent(`docker ps --format '{{.Names}}' | grep "^remnawave_bot$"`)
	if out != "" {
		printSuccess("Бот: работает")
	} else {
		printError("Бот: не запущен")
	}

	// PostgreSQL
	_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s exec -T postgres pg_isready -U remnawave_user 2>/dev/null", installDir, composeFile))
	if err == nil {
		printSuccess("PostgreSQL: работает")
	} else {
		printError("PostgreSQL: не доступен")
	}

	// Redis
	_, err = runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s exec -T redis redis-cli ping 2>/dev/null", installDir, composeFile))
	if err == nil {
		printSuccess("Redis: работает")
	} else {
		printError("Redis: не доступен")
	}

	// Docker
	if commandExists("docker") {
		printSuccess("Docker: установлен")
	} else {
		printError("Docker: не найден")
	}

	// Disk space
	out, _ = runShellSilent("df -h / | tail -1 | awk '{print $4}'")
	if out != "" {
		printInfo("Свободно на диске: " + out)
	}

	// Memory
	out, _ = runShellSilent("free -h | grep Mem | awk '{printf \"%s / %s\", $3, $2}'")
	if out != "" {
		printInfo("Память: " + out)
	}

	fmt.Println()
	fmt.Println(dimStyle.Render("  Последние логи:"))
	fmt.Println(sep)
	runShell(fmt.Sprintf("cd %s && docker compose -f %s logs --tail=10 bot 2>/dev/null", installDir, composeFile))
}

// ════════════════════════════════════════════════════════════════
// MANAGE: CONFIG
// ════════════════════════════════════════════════════════════════

func manageConfig(installDir string) {
	envPath := filepath.Join(installDir, ".env")
	if !fileExists(envPath) {
		printError("Файл .env не найден: " + envPath)
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
	printWarning("Перезапустите бота для применения изменений: bot restart")
	waitForEnter()
}

// ════════════════════════════════════════════════════════════════
// MANAGE: UNINSTALL
// ════════════════════════════════════════════════════════════════

func manageUninstall(installDir, composeFile string) {
	fmt.Println()
	fmt.Println(errorStyle.Render("  ╔══════════════════════════════════╗"))
	fmt.Println(errorStyle.Render("  ║         УДАЛЕНИЕ БОТА            ║"))
	fmt.Println(errorStyle.Render("  ╚══════════════════════════════════╝"))
	fmt.Println()

	if !confirmPrompt("Вы уверены, что хотите удалить бота?", false) {
		printSuccess("Отменено")
		return
	}

	val := inputText("Введите 'yes' для подтверждения", "", "Это действие нельзя отменить", true)
	if val != "yes" {
		printSuccess("Отменено")
		return
	}

	if confirmPrompt("Создать резервную копию перед удалением?", true) {
		runShellSilent(fmt.Sprintf(`cd %s && tar -czf "/root/bedolaga_backup_$(date +%%Y%%m%%d_%%H%%M%%S).tar.gz" .env data/ 2>/dev/null || true`, installDir))
		printSuccess("Резервная копия сохранена в /root/")
	}

	runWithSpinner("Остановка контейнеров...", func() error {
		runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down -v 2>/dev/null || true", installDir, composeFile))
		return nil
	})

	// Cleanup reverse proxy
	runShellSilent("rm -f /etc/nginx/sites-enabled/bedolaga-* /etc/nginx/sites-available/bedolaga-*")
	runShellSilent("nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true")
	if fileExists("/etc/caddy/Caddyfile") {
		runShellSilent(`sed -i '/# === BEGIN Bedolaga Bot ===/,/# === END Bedolaga Bot ===/d' /etc/caddy/Caddyfile`)
		runShellSilent("systemctl reload caddy 2>/dev/null || true")
	}

	os.Remove("/usr/local/bin/bot")

	if confirmPrompt("Удалить каталог "+installDir+"?", false) {
		os.RemoveAll(installDir)
		printSuccess("Каталог удалён")
	} else {
		printInfo("Каталог сохранён: " + installDir)
	}

	printSuccessBox(successStyle.Render("Удаление завершено!"))
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
		printError("Неизвестная команда: " + subcmd)
		fmt.Println()
		printManageHelp()
		os.Exit(1)
	}
}

func printManageHelp() {
	printBanner()
	fmt.Println(highlightStyle.Render("  Использование:") + dimStyle.Render(" bot [команда]"))
	fmt.Println()
	fmt.Println(dimStyle.Render("  (без аргументов)  ") + "Интерактивное меню со стрелками")
	fmt.Println()
	fmt.Println(infoStyle.Render("  logs            ") + "  Просмотр логов")
	fmt.Println(infoStyle.Render("  status          ") + "  Статус контейнеров и ресурсы")
	fmt.Println(infoStyle.Render("  restart         ") + "  Перезапуск контейнеров")
	fmt.Println(infoStyle.Render("  start           ") + "  Запуск контейнеров")
	fmt.Println(infoStyle.Render("  stop            ") + "  Остановка контейнеров")
	fmt.Println(infoStyle.Render("  update          ") + "  Обновление (git pull + rebuild)")
	fmt.Println(infoStyle.Render("  backup          ") + "  Создать резервную копию")
	fmt.Println(infoStyle.Render("  health          ") + "  Диагностика системы")
	fmt.Println(infoStyle.Render("  config          ") + "  Редактировать .env")
	fmt.Println(infoStyle.Render("  uninstall       ") + "  Удалить бота")
	fmt.Println(infoStyle.Render("  help            ") + "  Эта справка")
	fmt.Println()
}
