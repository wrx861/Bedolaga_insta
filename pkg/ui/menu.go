package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// LIGHTWEIGHT MENU (fast arrow-key navigation)
// ════════════════════════════════════════════════════════════════

type menuReadyMsg struct{}

type menuModel struct {
	title    string
	items    []SelectItem
	cursor   int
	selected int
	done     bool
	ready    bool
}

func newMenuModel(title string, items []SelectItem) menuModel {
	return menuModel{
		title:    title,
		items:    items,
		cursor:   0,
		selected: -1,
	}
}

func (m menuModel) Init() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return menuReadyMsg{}
	})
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case menuReadyMsg:
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		if !m.ready {
			return m, nil
		}
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil
		case "home":
			m.cursor = 0
			return m, nil
		case "end":
			m.cursor = len(m.items) - 1
			return m, nil
		case "enter":
			m.selected = m.cursor
			m.done = true
			return m, tea.Quit
		case "ctrl+c":
			return m, nil
		}
	}
	return m, nil
}

var (
	menuActiveTitle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	menuActiveDesc = lipgloss.NewStyle().
			Foreground(ColorDim)

	menuNormalTitle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	menuNormalDesc = lipgloss.NewStyle().
			Foreground(ColorDim)

	menuCursor = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Render("▸ ")

	menuBlank = "  "
)

func (m menuModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  " + SubtitleStyle.Render(m.title) + "\n\n")

	for i, item := range m.items {
		if i == m.cursor {
			b.WriteString("  " + menuCursor + menuActiveTitle.Render(item.Title))
			if item.Description != "" {
				b.WriteString(menuActiveDesc.Render("  " + item.Description))
			}
		} else {
			b.WriteString("  " + menuBlank + menuNormalTitle.Render(item.Title))
			if item.Description != "" {
				b.WriteString(menuNormalDesc.Render("  " + item.Description))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n  " + DimStyle.Render("↑/↓ Навигация  Enter Выбор") + "\n")
	return b.String()
}

// MenuOption — lightweight fast menu for management panels
func MenuOption(title string, items []SelectItem) int {
	if !IsInteractive() {
		fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s (авто)", title, items[0].Title)))
		return 0
	}
	m := newMenuModel(title, items)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s (по умолчанию)", title, items[0].Title)))
		return 0
	}
	final := result.(menuModel)
	if final.selected >= 0 && final.selected < len(items) {
		fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s", title, items[final.selected].Title)))
		return final.selected
	}
	fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s (по умолчанию)", title, items[0].Title)))
	return 0
}
