package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// SPINNER (for long operations)
// ════════════════════════════════════════════════════════════════

type spinnerDoneMsg struct{}

type spinnerModel struct {
	spinner  spinner.Model
	message  string
	done     bool
	quitting bool
}

func newSpinnerModel(msg string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(ColorAccent)
	return spinnerModel{spinner: s, message: msg}
}

func (m spinnerModel) Init() tea.Cmd { return m.spinner.Tick }

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		m.quitting = true
		return m, tea.Quit
	case tea.KeyMsg:
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	if m.done {
		return SuccessStyle.Render("  ✓ "+m.message) + "\n"
	}
	return fmt.Sprintf("  %s %s\n", m.spinner.View(), InfoStyle.Render(m.message))
}

func RunWithSpinner(message string, fn func() error) error {
	p := tea.NewProgram(newSpinnerModel(message))
	var fnErr error
	go func() {
		fnErr = fn()
		p.Send(spinnerDoneMsg{})
	}()
	if _, err := p.Run(); err != nil {
		return err
	}
	return fnErr
}
