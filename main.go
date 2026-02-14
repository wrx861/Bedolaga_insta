package main

import (
	"fmt"
	"os"

	"bedolaga-installer/pkg/ui"

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
		fmt.Println(lipgloss.NewStyle().Foreground(ui.ColorAccent).Bold(true).Render("bedolaga_installer") + " " + ui.DimStyle.Render("v"+appVersion))
	case "help", "--help", "-h":
		ui.PrintBanner(appVersion)
		fmt.Println(ui.HighlightStyle.Render("  Команды:"))
		fmt.Println(ui.DimStyle.Render("    install    ") + "Запустить мастер установки")
		fmt.Println(ui.DimStyle.Render("    manage     ") + "Панель управления ботом (TUI)")
		fmt.Println(ui.DimStyle.Render("    update     ") + "Обновить бота (git pull + пересборка)")
		fmt.Println(ui.DimStyle.Render("    uninstall  ") + "Удалить бота")
		fmt.Println(ui.DimStyle.Render("    version    ") + "Показать версию")
		fmt.Println(ui.DimStyle.Render("    help       ") + "Показать эту справку")
		fmt.Println()
	default:
		ui.PrintError("Неизвестная команда: " + os.Args[1])
		fmt.Println(ui.DimStyle.Render("  Используйте: bedolaga_installer help"))
		os.Exit(1)
	}
}
