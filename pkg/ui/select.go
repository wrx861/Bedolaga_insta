package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// INTERACTIVE SELECT (arrow keys)
// ════════════════════════════════════════════════════════════════

type SelectItem struct {
	Title       string
	Description string
}

func (i SelectItem) FilterValue() string { return i.Title }

type selectDelegate struct {
	list.DefaultDelegate
}

type selectModel struct {
	list     list.Model
	selected int
	done     bool
}

// Implement list.Item interface via wrapper
type selectListItem struct {
	title       string
	description string
}

func (i selectListItem) Title() string       { return i.title }
func (i selectListItem) Description() string { return i.description }
func (i selectListItem) FilterValue() string { return i.title }

func newSelectModel(title string, items []SelectItem) selectModel {
	var listItems []list.Item
	for _, item := range items {
		listItems = append(listItems, selectListItem{title: item.Title, description: item.Description})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		PaddingLeft(2)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(ColorDim).
		PaddingLeft(2)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(ColorWhite).
		PaddingLeft(2)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(ColorDim).
		PaddingLeft(2)

	l := list.New(listItems, delegate, 60, len(items)*3+4)
	l.Title = title
	l.Styles.Title = SubtitleStyle
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)

	return selectModel{list: l, selected: -1}
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.selected = m.list.Index()
			m.done = true
			return m, tea.Quit
		case "ctrl+c":
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string {
	if m.done {
		return ""
	}
	return "\n" + m.list.View() + "\n" + DimStyle.Render("  ↑/↓ Навигация  Enter Выбор") + "\n"
}

func SelectOption(title string, items []SelectItem) int {
	if !IsInteractive() {
		fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s (авто)", title, items[0].Title)))
		return 0
	}
	m := newSelectModel(title, items)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s (по умолчанию)", title, items[0].Title)))
		return 0
	}
	final := result.(selectModel)
	if final.selected >= 0 && final.selected < len(items) {
		fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s", title, items[final.selected].Title)))
		return final.selected
	}
	fmt.Println(SuccessStyle.Render(fmt.Sprintf("  ✓ %s: %s (по умолчанию)", title, items[0].Title)))
	return 0
}
