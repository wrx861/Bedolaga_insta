package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// ════════════════════════════════════════════════════════════════
// PROGRESS BAR (for downloads/builds)
// ════════════════════════════════════════════════════════════════

type progressDoneMsg struct{}

type progressModel struct {
	progress progress.Model
	message  string
	percent  float64
	done     bool
}

func NewProgressModel(msg string) progressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
	)
	return progressModel{progress: p, message: msg}
}

func (m progressModel) Init() tea.Cmd { return nil }

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd
	case float64:
		m.percent = msg
		return m, m.progress.SetPercent(msg)
	case progressDoneMsg:
		m.done = true
		return m, tea.Quit
	case tea.KeyMsg:
		return m, nil
	}
	return m, nil
}

func (m progressModel) View() string {
	if m.done {
		return SuccessStyle.Render("  ✓ " + m.message + " — Завершено") + "\n"
	}
	return fmt.Sprintf("\n  %s\n  %s\n", InfoStyle.Render(m.message), m.progress.View())
}
