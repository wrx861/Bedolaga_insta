package main

import (
        "fmt"
        "os"
        "syscall"

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

        // Fix stdin for curl | bash — reopen from /dev/tty
        reopenStdinIfPipe()

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


// reopenStdinIfPipe fixes interactive mode when run via "curl ... | bash".
// stdin is a pipe from curl, but /dev/tty is the real terminal.
func reopenStdinIfPipe() {
        fi, err := os.Stdin.Stat()
        if err != nil {
                return
        }
        if (fi.Mode() & os.ModeCharDevice) != 0 {
                return // already a terminal
        }
        tty, err := os.Open("/dev/tty")
        if err != nil {
                return // truly non-interactive (cron, CI, etc.)
        }
        // Replace fd 0 with /dev/tty
        syscall.Dup2(int(tty.Fd()), 0)
        tty.Close()
        os.Stdin = os.NewFile(0, "/dev/stdin")
}
