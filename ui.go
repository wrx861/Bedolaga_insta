package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	s.Style = lipgloss.NewStyle().Foreground(colorAccent)
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
		// Ignore all keys during spinner
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	if m.done {
		return successStyle.Render("  ✓ "+m.message) + "\n"
	}
	return fmt.Sprintf("  %s %s\n", m.spinner.View(), infoStyle.Render(m.message))
}

func runWithSpinner(message string, fn func() error) error {
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

// ════════════════════════════════════════════════════════════════
// INTERACTIVE SELECT (arrow keys)
// ════════════════════════════════════════════════════════════════

type selectItem struct {
	title       string
	description string
}

func (i selectItem) Title() string       { return i.title }
func (i selectItem) Description() string { return i.description }
func (i selectItem) FilterValue() string { return i.title }

type selectModel struct {
	list     list.Model
	selected int
	done     bool
}

func newSelectModel(title string, items []selectItem) selectModel {
	var listItems []list.Item
	for _, item := range items {
		listItems = append(listItems, item)
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		PaddingLeft(2)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(colorDim).
		PaddingLeft(2)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(colorWhite).
		PaddingLeft(2)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(colorDim).
		PaddingLeft(2)

	l := list.New(listItems, delegate, 60, len(items)*3+4)
	l.Title = title
	l.Styles.Title = subtitleStyle
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
			// Don't quit on Ctrl+C in select
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
	return "\n" + m.list.View() + "\n" + dimStyle.Render("  ↑/↓ Навигация  Enter Выбор") + "\n"
}

func selectOption(title string, items []selectItem) int {
	if !isInteractive() {
		// В неинтерактивном режиме выбираем первый вариант
		fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s: %s (авто)", title, items[0].title)))
		return 0
	}
	m := newSelectModel(title, items)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s: %s (по умолчанию)", title, items[0].title)))
		return 0
	}
	final := result.(selectModel)
	if final.selected >= 0 && final.selected < len(items) {
		fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s: %s", title, items[final.selected].title)))
		return final.selected
	}
	// Если ничего не выбрано — первый вариант
	fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s: %s (по умолчанию)", title, items[0].title)))
	return 0
}

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
	ti.PromptStyle = promptStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(colorWhite)
	ti.PlaceholderStyle = dimStyle
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)

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
	b.WriteString("  " + subtitleStyle.Render(m.label) + "\n")
	if m.hint != "" {
		b.WriteString("  " + dimStyle.Render(m.hint) + "\n")
	}
	b.WriteString("\n  " + m.input.View() + "\n")
	if m.errMsg != "" {
		b.WriteString("  " + errorStyle.Render("✗ "+m.errMsg) + "\n")
	}
	if m.required {
		b.WriteString("\n  " + dimStyle.Render("Enter Подтвердить") + "\n")
	} else {
		b.WriteString("\n  " + dimStyle.Render("Enter Подтвердить  Esc Пропустить") + "\n")
	}
	return b.String()
}

func inputText(label, placeholder, hint string, required bool) string {
	maxRetries := 3
	retries := 0
	for {
		m := newInputModel(label, placeholder, hint, required)
		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			if required {
				printError("Ошибка ввода: " + err.Error())
				os.Exit(1)
			}
			return ""
		}
		final := result.(inputModel)
		val := strings.TrimSpace(final.input.Value())
		if required && val == "" {
			retries++
			if retries >= maxRetries || !isInteractive() {
				printError("Это поле обязательно! Запустите установщик в интерактивном режиме.")
				os.Exit(1)
			}
			fmt.Println(errorStyle.Render("  ✗ Это поле обязательно, попробуйте снова"))
			continue
		}
		if val != "" {
			// Скрываем токены и ключи
			displayVal := val
			labelLower := strings.ToLower(label)
			if strings.Contains(labelLower, "token") || strings.Contains(labelLower, "key") || strings.Contains(labelLower, "password") || strings.Contains(labelLower, "secret") {
				if len(val) > 8 {
					displayVal = val[:4] + "..." + val[len(val)-4:]
				} else {
					displayVal = "***"
				}
			}
			fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s: %s", label, displayVal)))
		} else {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  - %s: пропущено", label)))
		}
		return val
	}
}

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
	yes := dimStyle.Render("  Да  ")
	no := dimStyle.Render("  Нет  ")
	if m.selected {
		yes = lipgloss.NewStyle().Foreground(colorBgAlt).Background(colorAccent).Bold(true).Render("  Да  ")
	} else {
		no = lipgloss.NewStyle().Foreground(colorBgAlt).Background(colorError).Bold(true).Render("  Нет  ")
	}
	return fmt.Sprintf("\n  %s\n\n  %s %s\n\n  %s\n",
		subtitleStyle.Render(m.prompt),
		yes, no,
		dimStyle.Render("←/→ Переключить  Enter Подтвердить  y/n Быстрый выбор"))
}

func confirmPrompt(prompt string, defaultYes bool) bool {
	if !isInteractive() {
		// В неинтерактивном режиме используем значение по умолчанию
		if defaultYes {
			fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s: Да (авто)", prompt)))
		} else {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  - %s: Нет (авто)", prompt)))
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
		fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s: Да", prompt)))
	} else {
		fmt.Println(dimStyle.Render(fmt.Sprintf("  - %s: Нет", prompt)))
	}
	return final.selected
}

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

func newProgressModel(msg string) progressModel {
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
		return successStyle.Render("  ✓ " + m.message + " — Завершено") + "\n"
	}
	return fmt.Sprintf("\n  %s\n  %s\n", infoStyle.Render(m.message), m.progress.View())
}
