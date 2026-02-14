package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// CONFIRM DIALOG
// ════════════════════════════════════════════════════════════════

type confirmModel struct {
	prompt   string
	selected bool // true = yes
	done     bool
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "tab":
			m.selected = !m.selected
			return m, nil
		case "right", "l":
			m.selected = !m.selected
			return m, nil
		case "y":
			m.selected = true
			m.done = true
			return m, tea.Quit
		case "n":
			m.selected = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "ctrl+c":
			return m, nil
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}
	yes := DimStyle.Render("  Да  ")
	no := DimStyle.Render("  Нет  ")
	if m.selected {
		yes = lipgloss.NewStyle().Foreground(ColorBgAlt).Background(ColorAccent).Bold(true).Render("  Да  ")
	} else {
		no = lipgloss.NewStyle().Foreground(ColorBgAlt).Background(ColorError).Bold(true).Render("  Нет  ")
	}
	return fmt.Sprintf("\n  %s\n\n  %s %s\n\n  %s\n",
		SubtitleStyle.Render(m.prompt),
		yes, no,
		DimStyle.Render("←/→ Переключить  Enter Подтвердить  y/n Быстрый выбор"))
}

func ConfirmPrompt(prompt string, defaultYes bool) bool {
	if !IsInteractive() {
		if defaultYes {
			fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: Да (авто)", prompt)))
		} else {
			fmt.Println(DimStyle.Render(fmt.Sprintf("  - %s: Нет (авто)", prompt)))
		}
		return defaultYes
	}
	m := confirmModel{prompt: prompt, selected: defaultYes}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return defaultYes
	}
	final := result.(confirmModel)
	if final.selected {
		fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: Да", prompt)))
	} else {
		fmt.Println(DimStyle.Render(fmt.Sprintf("  - %s: Нет", prompt)))
	}
	return final.selected
}
