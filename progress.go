package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"bedolaga-installer/pkg/ui"

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
	silent   bool
}

var globalProgress = installProgress{
	current: 0,
	total:   12,
	silent:  true,
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
	bar := lipgloss.NewStyle().Foreground(ui.ColorAccent).Render(strings.Repeat("━", filled))
	empty := lipgloss.NewStyle().Foreground(ui.ColorDim).Render(strings.Repeat("━", barWidth-filled))
	pctStr := lipgloss.NewStyle().Foreground(ui.ColorAccent).Bold(true).Render(fmt.Sprintf("%3d%%", int(pct*100)))

	stepLabel := lipgloss.NewStyle().Foreground(ui.ColorWhite).Bold(true).Render(stepName)
	counter := ui.DimStyle.Render(fmt.Sprintf("[%2d/%d]", p.current, p.total))

	p.lastLine = fmt.Sprintf("  %s %s%s %s  %s %s", counter, bar, empty, pctStr, ui.AccentBar.Render("▸"), stepLabel)

	fmt.Printf("\r\033[K%s", p.lastLine)

	if p.current == p.total {
		fmt.Println()
	}
}

func (p *installProgress) log(msg string) {
	if p.silent {
		return
	}
	fmt.Printf("\r\033[K%s\n", msg)
	if p.lastLine != "" {
		fmt.Printf("%s", p.lastLine)
	}
}

func (p *installProgress) done(msg string) {
	fmt.Printf("\r\033[K%s\n", ui.SuccessStyle.Render("  ✓ "+msg))
	if p.lastLine != "" && p.current < p.total {
		fmt.Printf("%s", p.lastLine)
	}
}

func (p *installProgress) fail(msg string) {
	fmt.Printf("\r\033[K%s\n", ui.ErrorStyle.Render("  ✗ "+msg))
}

func (p *installProgress) warn(msg string) {
	fmt.Printf("\r\033[K%s\n", ui.WarnStyle.Render("  ⚠ "+msg))
}

func (p *installProgress) info(msg string) {
	fmt.Printf("\r\033[K%s\n", ui.InfoStyle.Render("  ℹ "+msg))
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
			msg := ui.WarnStyle.Render("  ⚠  Обнаружено Ctrl+C!")
			fmt.Println(msg)
			fmt.Println(ui.DimStyle.Render("  Установка в процессе. Выход сейчас может оставить систему в нестабильном состоянии."))
			fmt.Println()
			fmt.Print(ui.PromptStyle.Render("  Точно выйти? ") + ui.DimStyle.Render("(введите 'yes' для подтверждения): "))
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) == "yes" {
				fmt.Println(ui.DimStyle.Render("  Выход..."))
				os.Exit(1)
			}
			fmt.Println(ui.SuccessStyle.Render("  ✓ Продолжаем установку..."))
			fmt.Println()
		}
	}()
}
