package main

import (
	"fmt"
	"os"
	"strings"

	"bedolaga-installer/pkg/ui"

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
	sep := lipgloss.NewStyle().Foreground(ui.ColorAccent).Render(strings.Repeat("═", 56))

	var b strings.Builder
	b.WriteString(ui.SuccessStyle.Render("  УСТАНОВКА ЗАВЕРШЕНА") + "\n\n")
	b.WriteString(ui.HighlightStyle.Render("  Каталог: ") + ui.InfoStyle.Render(cfg.InstallDir) + "\n")
	b.WriteString(ui.HighlightStyle.Render("  Конфиг:  ") + ui.InfoStyle.Render(cfg.InstallDir+"/.env") + "\n\n")

	b.WriteString(ui.HighlightStyle.Render("  Управление:") + "\n")
	b.WriteString(ui.DimStyle.Render("    bot          ") + "Интерактивное меню\n")
	b.WriteString(ui.DimStyle.Render("    bot logs     ") + "Просмотр логов\n")
	b.WriteString(ui.DimStyle.Render("    bot status   ") + "Статус контейнеров\n")
	b.WriteString(ui.DimStyle.Render("    bot update   ") + "Обновить бота\n\n")

	if cfg.WebhookDomain != "" {
		b.WriteString(ui.HighlightStyle.Render("  Webhook: ") + ui.InfoStyle.Render("https://"+cfg.WebhookDomain) + "\n")
	}
	if cfg.MiniappDomain != "" {
		b.WriteString(ui.HighlightStyle.Render("  MiniApp: ") + ui.InfoStyle.Render("https://"+cfg.MiniappDomain) + "\n")
	}

	fmt.Println("\n  " + sep)
	fmt.Println(ui.SuccessBoxStyle.Render(b.String()))
	fmt.Println("  " + sep)
}
