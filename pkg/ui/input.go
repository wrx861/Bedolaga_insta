package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// INTERACTIVE TEXT INPUT
// ════════════════════════════════════════════════════════════════

type inputModel struct {
	input    textinput.Model
	label    string
	hint     string
	done     bool
	required bool
	errMsg   string
}

func newInputModel(label, placeholder, hint string, required bool) inputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 50
	ti.PromptStyle = PromptStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorWhite)
	ti.PlaceholderStyle = DimStyle
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(ColorAccent)

	return inputModel{
		input:    ti,
		label:    label,
		hint:     hint,
		required: required,
	}
}

func (m inputModel) Init() tea.Cmd { return textinput.Blink }

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(m.input.Value())
			if m.required && val == "" {
				m.errMsg = "Это поле обязательно"
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case "ctrl+c":
			return m, nil
		case "esc":
			if !m.required {
				m.done = true
				return m, tea.Quit
			}
			return m, nil
		}
	}
	m.errMsg = ""
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  " + SubtitleStyle.Render(m.label) + "\n")
	if m.hint != "" {
		b.WriteString("  " + DimStyle.Render(m.hint) + "\n")
	}
	b.WriteString("\n  " + m.input.View() + "\n")
	if m.errMsg != "" {
		b.WriteString("  " + ErrorStyle.Render("✗ "+m.errMsg) + "\n")
	}
	if m.required {
		b.WriteString("\n  " + DimStyle.Render("Enter Подтвердить") + "\n")
	} else {
		b.WriteString("\n  " + DimStyle.Render("Enter Подтвердить  Esc Пропустить") + "\n")
	}
	return b.String()
}

func InputText(label, placeholder, hint string, required bool) string {
	maxRetries := 3
	retries := 0
	for {
		m := newInputModel(label, placeholder, hint, required)
		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			if required {
				PrintError("Ошибка ввода: " + err.Error())
				os.Exit(1)
			}
			return ""
		}
		final := result.(inputModel)
		val := strings.TrimSpace(final.input.Value())
		if required && val == "" {
			retries++
			if retries >= maxRetries || !IsInteractive() {
				PrintError("Это поле обязательно! Запустите установщик в интерактивном режиме.")
				os.Exit(1)
			}
			fmt.Println(ErrorStyle.Render("  ✗ Это поле обязательно, попробуйте снова"))
			continue
		}
		if val != "" {
			displayVal := val
			labelLower := strings.ToLower(label)
			if strings.Contains(labelLower, "token") || strings.Contains(labelLower, "key") || strings.Contains(labelLower, "password") || strings.Contains(labelLower, "secret") {
				if len(val) > 8 {
					displayVal = val[:4] + "..." + val[len(val)-4:]
				} else {
					displayVal = "***"
				}
			}
			fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s", label, displayVal)))
		} else {
			fmt.Println(DimStyle.Render(fmt.Sprintf("  - %s: пропущено", label)))
		}
		return val
	}
}
