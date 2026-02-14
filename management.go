package main

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// MANAGEMENT SCRIPT
// ════════════════════════════════════════════════════════════════

func createManagementScript(cfg *Config) {
	script := `#!/bin/bash
# REMNAWAVE BEDOLAGA BOT — Управление
# Этот скрипт запускает TUI-панель управления
exec bedolaga_installer manage "$@"
`
	os.WriteFile("/usr/local/bin/bot", []byte(script), 0755)
	globalProgress.done("Команда 'bot' установлена")
}

// ════════════════════════════════════════════════════════════════
// FINAL INFO
// ════════════════════════════════════════════════════════════════

func printFinalInfo(cfg *Config) {
	sep := lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Repeat("═", 56))

	var b strings.Builder
	b.WriteString(successStyle.Render("  УСТАНОВКА ЗАВЕРШЕНА") + "\n\n")
	b.WriteString(highlightStyle.Render("  Каталог: ") + infoStyle.Render(cfg.InstallDir) + "\n")
	b.WriteString(highlightStyle.Render("  Конфиг:  ") + infoStyle.Render(cfg.InstallDir+"/.env") + "\n\n")

	b.WriteString(highlightStyle.Render("  Управление:") + "\n")
	b.WriteString(dimStyle.Render("    bot          ") + "Интерактивное меню\n")
	b.WriteString(dimStyle.Render("    bot logs     ") + "Просмотр логов\n")
	b.WriteString(dimStyle.Render("    bot status   ") + "Статус контейнеров\n")
	b.WriteString(dimStyle.Render("    bot update   ") + "Обновить бота\n\n")

	if cfg.WebhookDomain != "" {
		b.WriteString(highlightStyle.Render("  Webhook: ") + infoStyle.Render("https://"+cfg.WebhookDomain) + "\n")
	}
	if cfg.MiniappDomain != "" {
		b.WriteString(highlightStyle.Render("  MiniApp: ") + infoStyle.Render("https://"+cfg.MiniappDomain) + "\n")
	}

	fmt.Println("\n  " + sep)
	fmt.Println(successBoxStyle.Render(b.String()))
	fmt.Println("  " + sep)
}
