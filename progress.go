package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// PROGRESS TRACKER
// ════════════════════════════════════════════════════════════════

type installProgress struct {
	current  int
	total    int
	steps    []string
	lastLine string
	silent   bool // Подавлять промежуточный вывод
}

var globalProgress = installProgress{
	current: 0,
	total:   12,
	silent:  true, // Лайв-режим по умолчанию
	steps: []string{
		"Проверка системы",
		"Установка пакетов",
		"Настройка Docker",
		"Каталог установки",
		"Конфигурация панели",
		"Проверка данных",
		"Клонирование репозитория",
		"Интерактивная настройка",
		"Файл окружения",
		"Обратный прокси",
		"Docker-контейнеры",
		"Завершение",
	},
}

func (p *installProgress) advance(stepName string) {
	p.current++
	pct := float64(p.current) / float64(p.total)

	barWidth := 40
	filled := int(pct * float64(barWidth))
	bar := lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Repeat("━", filled))
	empty := lipgloss.NewStyle().Foreground(colorDim).Render(strings.Repeat("━", barWidth-filled))
	pctStr := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(fmt.Sprintf("%3d%%", int(pct*100)))

	stepLabel := lipgloss.NewStyle().Foreground(colorWhite).Bold(true).Render(stepName)
	counter := dimStyle.Render(fmt.Sprintf("[%2d/%d]", p.current, p.total))

	p.lastLine = fmt.Sprintf("  %s %s%s %s  %s %s", counter, bar, empty, pctStr, accentBar.Render("▸"), stepLabel)

	// Очищаем строку и печатаем прогресс на той же позиции (ANSI escape: \033[K)
	fmt.Printf("\r\033[K%s", p.lastLine)

	// Переход на новую строку только в конце
	if p.current == p.total {
		fmt.Println()
	}
}

// Показать сообщение над прогресс-баром
func (p *installProgress) log(msg string) {
	if p.silent {
		return
	}
	// Очистить текущую строку, напечатать сообщение, затем прогресс снова
	fmt.Printf("\r\033[K%s\n", msg)
	if p.lastLine != "" {
		fmt.Printf("%s", p.lastLine)
	}
}

// Показать финальный статус этапа (всегда показывается)
func (p *installProgress) done(msg string) {
	fmt.Printf("\r\033[K%s\n", successStyle.Render("  ✓ "+msg))
	if p.lastLine != "" && p.current < p.total {
		fmt.Printf("%s", p.lastLine)
	}
}

// Показать ошибку (всегда показывается)
func (p *installProgress) fail(msg string) {
	fmt.Printf("\r\033[K%s\n", errorStyle.Render("  ✗ "+msg))
}

// Показать предупреждение (всегда показывается)
func (p *installProgress) warn(msg string) {
	fmt.Printf("\r\033[K%s\n", warnStyle.Render("  ⚠ "+msg))
}

// Показать инфо (всегда показывается)
func (p *installProgress) info(msg string) {
	fmt.Printf("\r\033[K%s\n", infoStyle.Render("  ℹ "+msg))
}

// ════════════════════════════════════════════════════════════════
// SIGNAL HANDLING (Ctrl+C protection)
// ════════════════════════════════════════════════════════════════

var sigChan = make(chan os.Signal, 1)
var allowExit = false

func setupSignalHandler() {
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigChan {
			if allowExit {
				fmt.Println()
				os.Exit(0)
			}
			fmt.Println()
			fmt.Println()
			msg := warnStyle.Render("  ⚠  Обнаружено Ctrl+C!")
			fmt.Println(msg)
			fmt.Println(dimStyle.Render("  Установка в процессе. Выход сейчас может оставить систему в нестабильном состоянии."))
			fmt.Println()
			fmt.Print(promptStyle.Render("  Точно выйти? ") + dimStyle.Render("(введите 'yes' для подтверждения): "))
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) == "yes" {
				fmt.Println(dimStyle.Render("  Выход..."))
				os.Exit(1)
			}
			fmt.Println(successStyle.Render("  ✓ Продолжаем установку..."))
			fmt.Println()
		}
	}()
}
