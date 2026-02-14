package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// VERSION & CONSTANTS
// ════════════════════════════════════════════════════════════════

var appVersion = "2.2.0"

const repoURL = "https://github.com/BEDOLAGA-DEV/remnawave-bedolaga-telegram-bot.git"

// ════════════════════════════════════════════════════════════════
// MAIN
// ════════════════════════════════════════════════════════════════

func main() {
	// Переходим в существующую директорию чтобы избежать ошибок getcwd
	os.Chdir("/root")

	setupSignalHandler()

	if len(os.Args) < 2 {
		installWizard()
		return
	}

	switch os.Args[1] {
	case "install":
		installWizard()
	case "manage":
		manageBot()
	case "update", "upgrade":
		updateBot()
	case "uninstall", "remove":
		uninstallBot()
	case "version", "--version", "-v":
		fmt.Println(lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("bedolaga_installer") + " " + dimStyle.Render("v"+appVersion))
	case "help", "--help", "-h":
		printBanner()
		fmt.Println(highlightStyle.Render("  Команды:"))
		fmt.Println(dimStyle.Render("    install    ") + "Запустить мастер установки")
		fmt.Println(dimStyle.Render("    manage     ") + "Панель управления ботом (TUI)")
		fmt.Println(dimStyle.Render("    update     ") + "Обновить бота (git pull + пересборка)")
		fmt.Println(dimStyle.Render("    uninstall  ") + "Удалить бота")
		fmt.Println(dimStyle.Render("    version    ") + "Показать версию")
		fmt.Println(dimStyle.Render("    help       ") + "Показать эту справку")
		fmt.Println()
	default:
		printError("Неизвестная команда: " + os.Args[1])
		fmt.Println(dimStyle.Render("  Используйте: bedolaga_installer help"))
		os.Exit(1)
	}
}
