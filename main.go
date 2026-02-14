package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// VERSION & CONSTANTS
// ════════════════════════════════════════════════════════════════

var appVersion = "2.0.0"

const repoURL = "https://github.com/BEDOLAGA-DEV/remnawave-bedolaga-telegram-bot.git"

// ════════════════════════════════════════════════════════════════
// COLOR PALETTE (Premium Dark)
// ════════════════════════════════════════════════════════════════

var (
	colorPrimary   = lipgloss.Color("#A78BFA") // Violet
	colorSecondary = lipgloss.Color("#818CF8") // Indigo
	colorAccent    = lipgloss.Color("#F59E0B") // Amber/Gold
	colorSuccess   = lipgloss.Color("#34D399") // Emerald
	colorError     = lipgloss.Color("#F87171") // Red
	colorWarning   = lipgloss.Color("#FBBF24") // Yellow
	colorInfo      = lipgloss.Color("#60A5FA") // Blue
	colorDim       = lipgloss.Color("#6B7280") // Gray
	colorWhite     = lipgloss.Color("#F9FAFB") // Near-white
	colorBg        = lipgloss.Color("#1F2937") // Dark bg
	colorBgAlt     = lipgloss.Color("#111827") // Darker bg
)

// ════════════════════════════════════════════════════════════════
// STYLES
// ════════════════════════════════════════════════════════════════

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	stepStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(colorInfo)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	warnStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	highlightStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1).
			Width(70)

	successBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSuccess).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorError).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	accentBar = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)
)

// ════════════════════════════════════════════════════════════════
// BANNER
// ════════════════════════════════════════════════════════════════

func printBanner() {
	banner := `
    ____  __________  ____  __    ___   _________ 
   / __ )/ ____/ __ \/ __ \/ /   /   | / ____/   |
  / __  / __/ / / / / / / / /   / /| |/ / __/ /| |
 / /_/ / /___/ /_/ / /_/ / /___/ ___ / /_/ / ___ |
/_____/_____/_____/\____/_____/_/  |_\____/_/  |_|
                                                   `

	bannerStyled := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render(banner)

	tagline := lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Render("  УСТАНОВЩИК BEDOLAGA BOT")

	version := dimStyle.Render(fmt.Sprintf("  v%s", appVersion))

	separator := lipgloss.NewStyle().
		Foreground(colorDim).
		Render("  ─────────────────────────────────────────────")

	fmt.Println(bannerStyled)
	fmt.Println(tagline + "  " + version)
	fmt.Println(separator)
	fmt.Println()
}

// ════════════════════════════════════════════════════════════════
// PROGRESS TRACKER
// ════════════════════════════════════════════════════════════════

type installProgress struct {
	current int
	total   int
	steps   []string
}

var globalProgress = installProgress{
	current: 0,
	total:   12,
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

	// Очищаем строку и печатаем прогресс на той же позиции (ANSI escape: \033[K)
	fmt.Printf("\r\033[K  %s %s%s %s  %s %s", counter, bar, empty, pctStr, accentBar.Render("▸"), stepLabel)
	
	// Переход на новую строку только в конце
	if p.current == p.total {
		fmt.Println()
	}
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
		fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s (авто)", items[0].title)))
		return 0
	}
	m := newSelectModel(title, items)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s (по умолчанию)", items[0].title)))
		return 0
	}
	final := result.(selectModel)
	if final.selected >= 0 && final.selected < len(items) {
		fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s", items[final.selected].title)))
		return final.selected
	}
	// Если ничего не выбрано — первый вариант
	fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s (по умолчанию)", items[0].title)))
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
			fmt.Println(successStyle.Render("  ✓ Установлено"))
		} else {
			fmt.Println(dimStyle.Render("  - Пропущено"))
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
			fmt.Println(successStyle.Render("  ✓ Да (по умолчанию)"))
		} else {
			fmt.Println(dimStyle.Render("  - Нет (по умолчанию)"))
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
		fmt.Println(successStyle.Render("  ✓ Да"))
	} else {
		fmt.Println(dimStyle.Render("  - Нет"))
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

// ════════════════════════════════════════════════════════════════
// PRINT HELPERS
// ════════════════════════════════════════════════════════════════

func printStep(msg string)    { fmt.Println("\n" + stepStyle.Render("  ▸ "+msg)) }
func printInfo(msg string)    { fmt.Println(infoStyle.Render("  ℹ " + msg)) }
func printSuccess(msg string) { fmt.Println(successStyle.Render("  ✓ " + msg)) }
func printError(msg string)   { fmt.Println(errorStyle.Render("  ✗ " + msg)) }
func printWarning(msg string) { fmt.Println(warnStyle.Render("  ⚠ " + msg)) }
func printDim(msg string)     { fmt.Println(dimStyle.Render("    " + msg)) }

// Версии для вывода внутри прогресса (без переноса строки, с очисткой)
func printLiveInfo(msg string)    { fmt.Printf("\r\033[K%s\n", infoStyle.Render("  ℹ "+msg)) }
func printLiveSuccess(msg string) { fmt.Printf("\r\033[K%s\n", successStyle.Render("  ✓ "+msg)) }
func printLiveWarning(msg string) { fmt.Printf("\r\033[K%s\n", warnStyle.Render("  ⚠ "+msg)) }
func printLiveError(msg string)   { fmt.Printf("\r\033[K%s\n", errorStyle.Render("  ✗ "+msg)) }

func printBox(title, content string) {
	inner := subtitleStyle.Render(title) + "\n" + content
	fmt.Println(boxStyle.Render(inner))
}

func printSuccessBox(content string) {
	fmt.Println(successBoxStyle.Render(content))
}

func printErrorBox(content string) {
	fmt.Println(errorBoxStyle.Render(content))
}

// ════════════════════════════════════════════════════════════════
// SYSTEM UTILS
// ════════════════════════════════════════════════════════════════

var reader = bufio.NewReader(os.Stdin)

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = "/root"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runCmdSilent(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = "/root"
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func runShell(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = "/root"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runShellSilent(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = "/root"
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("команда превысила таймаут: %s", command)
	}
	return strings.TrimSpace(string(out)), err
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func isInteractive() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSafePassword(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func validateDomain(domain string) bool {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	if !strings.Contains(domain, ".") {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.\-]+[a-zA-Z0-9]$`)
	return re.MatchString(domain)
}

func cleanDomain(input string) string {
	d := strings.TrimPrefix(input, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimSuffix(d, "/")
	return d
}

func checkDomainDNS(domain string) bool {
	serverIP, err := runShellSilent("curl -4 -s --connect-timeout 5 ifconfig.me 2>/dev/null || curl -4 -s --connect-timeout 5 icanhazip.com 2>/dev/null")
	if err != nil || serverIP == "" {
		printWarning("Не удалось определить IP сервера")
		return false
	}
	ips, err := net.LookupHost(domain)
	if err != nil || len(ips) == 0 {
		printWarning(fmt.Sprintf("DNS-запись для %s не найдена", domain))
		return false
	}
	for _, ip := range ips {
		if ip == strings.TrimSpace(serverIP) {
			printSuccess(fmt.Sprintf("DNS %s → %s", domain, ip))
			return true
		}
	}
	printWarning(fmt.Sprintf("Домен %s → %s, IP сервера %s", domain, ips[0], serverIP))
	return false
}

// ════════════════════════════════════════════════════════════════
// CONFIG
// ════════════════════════════════════════════════════════════════

type Config struct {
	InstallDir            string
	PanelInstalledLocally bool
	PanelDir              string
	DockerNetwork         string

	BotToken       string
	AdminIDs       string
	SupportUsername string

	RemnawaveAPIURL    string
	RemnawaveAPIKey    string
	RemnawaveAuthType  string
	RemnawaveUsername   string
	RemnawavePassword  string
	RemnawaveSecretKey string

	WebhookDomain string
	MiniappDomain string

	AdminNotificationsChatID string

	PostgresPassword    string
	KeepExistingVolumes bool
	OldPostgresPassword string

	WebhookSecretToken string
	WebAPIDefaultToken string
	BotRunMode         string
	WebhookURL         string
	WebAPIEnabled      string

	ReverseProxyType string
	SSLEmail         string
}

// ════════════════════════════════════════════════════════════════
// SYSTEM CHECKS
// ════════════════════════════════════════════════════════════════

func checkRoot() {
	if os.Getuid() != 0 {
		printErrorBox(errorStyle.Render("Этот скрипт должен быть запущен от root!"))
		os.Exit(1)
	}
}

func detectOS() string {
	out, _ := runShellSilent("cat /etc/os-release 2>/dev/null | grep ^ID= | cut -d= -f2 | tr -d '\"'")
	prettyName, _ := runShellSilent("cat /etc/os-release 2>/dev/null | grep ^PRETTY_NAME= | cut -d= -f2 | tr -d '\"'")
	if prettyName != "" {
		printInfo("ОС: " + prettyName)
	}
	switch out {
	case "ubuntu", "debian":
		return out
	default:
		if out != "" {
			printWarning("Оптимизировано для Ubuntu/Debian. Обнаружено: " + out)
			if !confirmPrompt("Продолжить на неподдерживаемой ОС?", false) {
				os.Exit(0)
			}
		}
		return out
	}
}

// ════════════════════════════════════════════════════════════════
// PACKAGE INSTALLATION
// ════════════════════════════════════════════════════════════════

func updateSystem() {
	runWithSpinner("Обновление списка пакетов...", func() error {
		_, err := runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get update -y -qq 2>/dev/null")
		return err
	})
}

func installBasePackages() {
	runWithSpinner("Установка базовых пакетов...", func() error {
		// Устанавливаем по частям для надёжности
		packages := []string{
			"curl wget git",
			"nano htop",
			"make openssl ca-certificates gnupg",
			"lsb-release dnsutils",
		}
		for _, pkg := range packages {
			runShellSilent(fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq %s 2>/dev/null || true", pkg))
		}
		// certbot опционален - может не быть в некоторых системах
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq certbot python3-certbot-nginx 2>/dev/null || true")
		return nil
	})
}

func installDocker() {
	if commandExists("docker") {
		ver, _ := runShellSilent("docker --version")
		printSuccess("Docker: " + ver)
	} else {
		runWithSpinner("Установка Docker...", func() error {
			_, err := runShellSilent("DEBIAN_FRONTEND=noninteractive curl -fsSL https://get.docker.com | sh")
			if err != nil {
				return err
			}
			runShellSilent("systemctl enable docker 2>/dev/null || true")
			runShellSilent("systemctl start docker 2>/dev/null || true")
			return nil
		})
		// Проверяем, что Docker реально установился
		if !commandExists("docker") {
			printErrorBox("Не удалось установить Docker!")
			printInfo("Попробуйте установить Docker вручную: curl -fsSL https://get.docker.com | sh")
			os.Exit(1)
		}
		ver, _ := runShellSilent("docker --version")
		printSuccess("Docker установлен: " + ver)
	}
	
	// Проверяем Docker Compose
	if out, err := runShellSilent("docker compose version 2>/dev/null"); err == nil && out != "" {
		printSuccess("Docker Compose: " + out)
	} else if out, err := runShellSilent("docker-compose --version 2>/dev/null"); err == nil && out != "" {
		printSuccess("Docker Compose (standalone): " + out)
	} else {
		printErrorBox("Docker Compose не найден!")
		printInfo("Установите Docker Compose: apt install docker-compose-plugin")
		os.Exit(1)
	}
}

func installNginx() {
	if commandExists("nginx") {
		return
	}
	runWithSpinner("Установка Nginx...", func() error {
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq nginx")
		runShellSilent("systemctl enable nginx")
		runShellSilent("systemctl start nginx")
		return nil
	})
}

func installCaddy() {
	if commandExists("caddy") {
		return
	}
	runWithSpinner("Установка Caddy...", func() error {
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl")
		runShellSilent("curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg 2>/dev/null || true")
		runShellSilent("curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list > /dev/null")
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get update -y -qq")
		runShellSilent("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq caddy")
		return nil
	})
}

// ════════════════════════════════════════════════════════════════
// INTERACTIVE SETUP STEPS
// ════════════════════════════════════════════════════════════════

func selectInstallDir(cfg *Config) {
	idx := selectOption("Каталог установки", []selectItem{
		{title: "/opt/remnawave-bedolaga-telegram-bot", description: "Рекомендуемое расположение"},
		{title: "/root/remnawave-bedolaga-telegram-bot", description: "Домашний каталог"},
		{title: "Свой путь", description: "Указать свой путь"},
	})
	switch idx {
	case 0:
		cfg.InstallDir = "/opt/remnawave-bedolaga-telegram-bot"
	case 1:
		cfg.InstallDir = "/root/remnawave-bedolaga-telegram-bot"
	case 2:
		cfg.InstallDir = inputText("Путь установки", "/opt/my-bot", "Введите полный путь", true)
	}
	printInfo("Каталог: " + highlightStyle.Render(cfg.InstallDir))
}

func checkRemnawavePanel(cfg *Config) {
	idx := selectOption("Расположение панели", []selectItem{
		{title: "Панель на этом сервере", description: "Бот подключается через Docker-сеть"},
		{title: "Панель на другом сервере", description: "Бот подключается через внешний URL"},
	})
	switch idx {
	case 0:
		cfg.PanelInstalledLocally = true
		setupLocalPanel(cfg)
	case 1:
		cfg.PanelInstalledLocally = false
		printInfo("Автономный режим — укажите внешний URL при настройке")
	}
}

func setupLocalPanel(cfg *Config) {
	idx := selectOption("Каталог панели", []selectItem{
		{title: "/opt/remnawave", description: "Стандартный путь установки"},
		{title: "/root/remnawave", description: "Домашний каталог"},
		{title: "Свой путь", description: "Указать свой путь"},
	})
	switch idx {
	case 0:
		cfg.PanelDir = "/opt/remnawave"
	case 1:
		cfg.PanelDir = "/root/remnawave"
	case 2:
		cfg.PanelDir = inputText("Путь к каталогу панели", "/opt/remnawave", "", true)
	}

	if !dirExists(cfg.PanelDir) {
		printWarning("Каталог " + cfg.PanelDir + " не найден")
		printInfo("Переключаемся на режим внешней панели — укажите URL позже")
		cfg.PanelInstalledLocally = false
		cfg.PanelDir = ""
		cfg.DockerNetwork = ""
		return
	}
	
	printSuccess("Панель найдена: " + cfg.PanelDir)
	detectPanelNetwork(cfg)
}

func detectPanelNetwork(cfg *Config) {
	printStep("Поиск Docker-сети")

	found := false

	// Method 1: by running container
	if out, err := runShellSilent(`docker inspect remnawave --format '{{range $net, $config := .NetworkSettings.Networks}}{{$net}}{{"\n"}}{{end}}' 2>/dev/null | grep -v "^$" | grep -v "host" | grep -v "none" | head -1`); err == nil && out != "" {
		cfg.DockerNetwork = out
		found = true
	}

	// Method 2: known names
	if !found {
		known := []string{"remnawave-network", "remnawave_default", "remnawave_network", "remnawave", "remnawave-panel_default"}
		for _, n := range known {
			if _, err := runShellSilent(fmt.Sprintf("docker network inspect %s 2>/dev/null", n)); err == nil {
				cfg.DockerNetwork = n
				found = true
				break
			}
		}
	}

	// Method 3: grep
	if !found {
		if out, err := runShellSilent(`docker network ls --format '{{.Name}}' | grep -i "remnawave" | grep -v "bedolaga" | grep -v "bot" | head -1`); err == nil && out != "" {
			cfg.DockerNetwork = out
			found = true
		}
	}

	if found {
		printSuccess("Docker-сеть: " + highlightStyle.Render(cfg.DockerNetwork))
	} else {
		printWarning("Автоопределение не удалось")
		nets, _ := runShellSilent(`docker network ls --format '{{.Name}}' | grep -v "bridge\|host\|none"`)
		if nets != "" {
			printInfo("Доступные сети:")
			for _, n := range strings.Split(nets, "\n") {
				n = strings.TrimSpace(n)
				if n != "" {
					printDim("• " + n)
				}
			}
		}
		manual := inputText("Имя Docker-сети", "remnawave-network", "Оставьте пустым для пропуска", false)
		if manual != "" {
			cfg.DockerNetwork = manual
		}
	}
}

func inputDomainSafe(label, hint string) string {
	for {
		val := inputText(label, "bot.example.com", hint, false)
		if val == "" {
			return ""
		}
		val = cleanDomain(val)
		if !validateDomain(val) {
			printError("Неверный формат домена: " + val)
			printDim("Ожидаемый формат: bot.example.com")
			idx := selectOption("Что делать?", []selectItem{
				{title: "Попробовать снова", description: "Ввести другой домен"},
				{title: "Использовать всё равно", description: "Продолжить с этим значением"},
				{title: "Пропустить", description: "Не настраивать этот домен"},
			})
			switch idx {
			case 0:
				continue
			case 1:
				return val
			case 2:
				return ""
			}
		}
		printInfo("Проверка DNS...")
		if !checkDomainDNS(val) {
			printWarning("DNS не указывает на этот сервер")
			idx := selectOption("Что делать?", []selectItem{
				{title: "Попробовать другой домен", description: "Ввести другой домен"},
				{title: "Продолжить с этим доменом", description: "DNS можно настроить позже"},
				{title: "Пропустить", description: "Не настраивать этот домен"},
			})
			switch idx {
			case 0:
				continue
			case 1:
				return val
			case 2:
				return ""
			}
		}
		return val
	}
}

func checkPostgresVolume(cfg *Config) {
	cfg.KeepExistingVolumes = false
	cfg.OldPostgresPassword = ""

	if fileExists(filepath.Join(cfg.InstallDir, ".env")) {
		if out, err := runShellSilent(fmt.Sprintf(`grep "^POSTGRES_PASSWORD=" "%s/.env" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'"`, cfg.InstallDir)); err == nil && out != "" {
			cfg.OldPostgresPassword = out
		}
	}

	foundVolumes, _ := runShellSilent(`docker volume ls -q 2>/dev/null | grep -E "(postgres|bot)" | grep -v "remnawave_postgres" || true`)
	if strings.TrimSpace(foundVolumes) == "" {
		printInfo("Существующих томов PostgreSQL нет — чистая установка")
		return
	}

	printWarning("Найдены существующие Docker-тома:")
	for _, v := range strings.Split(foundVolumes, "\n") {
		v = strings.TrimSpace(v)
		if v != "" {
			printDim("• " + v)
		}
	}

	if cfg.OldPostgresPassword != "" {
		printSuccess("Найден старый пароль PostgreSQL в .env")
		idx := selectOption("Существующие данные", []selectItem{
			{title: "Сохранить данные", description: "Сохранить базу, восстановить пароль (рекомендуется)"},
			{title: "Чистая установка", description: "Удалить тома, начать с нуля"},
		})
		if idx == 0 {
			cfg.KeepExistingVolumes = true
		} else {
			runShellSilent(fmt.Sprintf("cd %s 2>/dev/null && docker compose down -v 2>/dev/null || true", cfg.InstallDir))
		}
	} else {
		printWarning("Пароль PostgreSQL не найден — требуется чистая установка")
		if confirmPrompt("Удалить тома и начать с нуля?", true) {
			runShellSilent(fmt.Sprintf("cd %s 2>/dev/null && docker compose down -v 2>/dev/null || true", cfg.InstallDir))
		}
	}
}

func interactiveSetup(cfg *Config) {
	printBox("⚙️  Интерактивная настройка",
		"Введите необходимые данные для настройки бота.\n"+
			dimStyle.Render("Необязательные поля можно пропустить клавишей Esc."))

	// 1. BOT_TOKEN
	cfg.BotToken = inputText("BOT_TOKEN", "123456:ABC-DEF...", "Получить у @BotFather в Telegram", true)

	// 2. ADMIN_IDS
	cfg.AdminIDs = inputText("ADMIN_IDS", "123456789", "Ваш Telegram ID (несколько: 123,456). Узнать у @userinfobot", true)

	// 3. REMNAWAVE_API_URL
	if cfg.PanelInstalledLocally && cfg.DockerNetwork != "" {
		printInfo("Локальная панель — используется внутренний Docker-адрес")
		val := inputText("REMNAWAVE_API_URL", "http://remnawave:3000", "Внутренний адрес для локальной панели", false)
		if val == "" {
			val = "http://remnawave:3000"
		}
		cfg.RemnawaveAPIURL = val
	} else {
		cfg.RemnawaveAPIURL = inputText("REMNAWAVE_API_URL", "https://panel.yourdomain.com", "Внешний URL панели Remnawave", true)
	}

	// 4. REMNAWAVE_API_KEY
	cfg.RemnawaveAPIKey = inputText("REMNAWAVE_API_KEY", "", "Получить в настройках панели Remnawave", true)

	// 5. Auth type
	idx := selectOption("Тип авторизации", []selectItem{
		{title: "API Key", description: "По умолчанию — только API-ключ"},
		{title: "Basic Auth", description: "Авторизация по логину и паролю"},
	})
	cfg.RemnawaveAuthType = "api_key"
	if idx == 1 {
		cfg.RemnawaveAuthType = "basic_auth"
		cfg.RemnawaveUsername = inputText("REMNAWAVE_USERNAME", "", "", false)
		cfg.RemnawavePassword = inputText("REMNAWAVE_PASSWORD", "", "", false)
	}

	// 6. eGames SECRET_KEY
	isLocal := strings.HasPrefix(cfg.RemnawaveAPIURL, "http://remnawave:") ||
		strings.HasPrefix(cfg.RemnawaveAPIURL, "http://localhost:") ||
		strings.HasPrefix(cfg.RemnawaveAPIURL, "http://127.0.0.1:")

	if isLocal {
		printInfo("Локальное подключение к панели — SECRET_KEY не требуется")
	} else {
		if confirmPrompt("Панель установлена через eGames скрипт? (добавляет защиту URL секретом)", false) {
			printDim("Формат: КЛЮЧ:ЗНАЧЕНИЕ (с двоеточием!)")
			printDim("Пример: MHPsUKCz:VfHqrBwp")
			sk := inputText("REMNAWAVE_SECRET_KEY", "KEY:VALUE", "Используйте ДВОЕТОЧИЕ (:), не равно (=)!", false)
			if strings.Contains(sk, "=") && !strings.Contains(sk, ":") {
				sk = strings.Replace(sk, "=", ":", 1)
				printWarning("Автоисправление: заменено = на : → " + sk)
			}
			cfg.RemnawaveSecretKey = sk
		}
	}

	// 7. Webhook domain
	cfg.WebhookDomain = inputDomainSafe("Домен вебхука (необязательно)", "Для режима webhook. Оставьте пустым для polling.")

	// 8. Miniapp domain
	cfg.MiniappDomain = inputDomainSafe("Домен Mini App (необязательно)", "Домен для Telegram Mini App")

	// 9. Notifications
	cfg.AdminNotificationsChatID = inputText("Chat ID уведомлений (необязательно)", "-1001234567890", "ID чата/группы Telegram для уведомлений администратора", false)

	// 10. PostgreSQL password
	if cfg.KeepExistingVolumes && cfg.OldPostgresPassword != "" {
		cfg.PostgresPassword = cfg.OldPostgresPassword
		printSuccess("PostgreSQL: используется сохранённый пароль")
	} else {
		pw := inputText("Пароль PostgreSQL (необязательно)", "", "Оставьте пустым для автогенерации безопасного пароля", false)
		if pw == "" {
			cfg.PostgresPassword = generateSafePassword(24)
			printSuccess("Сгенерирован безопасный пароль PostgreSQL")
		} else {
			cfg.PostgresPassword = pw
		}
	}

	// 11. Reverse proxy
	if cfg.WebhookDomain != "" || cfg.MiniappDomain != "" {
		proxyItems := []selectItem{
			{title: "Nginx (системный)", description: "Автономный nginx на сервере"},
			{title: "Caddy", description: "Автоматический HTTPS, простая настройка"},
			{title: "Пропустить", description: "Настроить вручную позже"},
		}
		if cfg.PanelInstalledLocally {
			nginxNet, _ := runShellSilent("docker inspect remnawave-nginx --format '{{.HostConfig.NetworkMode}}' 2>/dev/null")
			if strings.TrimSpace(nginxNet) == "host" {
				proxyItems = []selectItem{
					{title: "Nginx (панели)", description: "Добавить в nginx панели (host mode)"},
					{title: "Nginx (системный)", description: "Автономный nginx на сервере"},
					{title: "Caddy", description: "Автоматический HTTPS, простая настройка"},
					{title: "Пропустить", description: "Настроить вручную позже"},
				}
			}
		}

		idx := selectOption("Обратный прокси", proxyItems)
		title := proxyItems[idx].title
		switch {
		case strings.Contains(title, "панели"):
			cfg.ReverseProxyType = "nginx_panel"
		case strings.Contains(title, "системный"):
			cfg.ReverseProxyType = "nginx_system"
		case strings.Contains(title, "Caddy"):
			cfg.ReverseProxyType = "caddy"
		default:
			cfg.ReverseProxyType = "skip"
		}
	} else {
		cfg.ReverseProxyType = "skip"
	}

	// Generate tokens
	cfg.WebhookSecretToken = generateToken()
	cfg.WebAPIDefaultToken = generateToken()
	cfg.SupportUsername = "@support"

	if cfg.WebhookDomain != "" {
		cfg.BotRunMode = "webhook"
		cfg.WebhookURL = "https://" + cfg.WebhookDomain
		cfg.WebAPIEnabled = "true"
	} else {
		cfg.BotRunMode = "polling"
		cfg.WebhookURL = ""
		cfg.WebAPIEnabled = "false"
	}

	printSuccessBox(successStyle.Render("Настройка завершена!"))
}

// ════════════════════════════════════════════════════════════════
// CLONE & DIRECTORIES
// ════════════════════════════════════════════════════════════════

func cloneRepository(cfg *Config) {
	if dirExists(cfg.InstallDir) {
		printWarning("Каталог " + cfg.InstallDir + " уже существует")
		if confirmPrompt("Удалить и клонировать заново?", false) {
			os.RemoveAll(cfg.InstallDir)
		} else {
			printInfo("Используем существующий каталог, обновляем...")
			runShellSilent(fmt.Sprintf("cd %s && git pull origin main || true", cfg.InstallDir))
			return
		}
	}

	runWithSpinner("Клонирование репозитория...", func() error {
		_, err := runCmdSilent("git", "clone", repoURL, cfg.InstallDir)
		return err
	})
	printSuccess("Клонировано в " + cfg.InstallDir)
}

func createDirectories(cfg *Config) {
	dirs := []string{"logs", "data", "data/backups", "data/referral_qr", "locales"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(cfg.InstallDir, d), 0755)
	}
	runShellSilent(fmt.Sprintf("chmod -R 755 %s/logs %s/data %s/locales 2>/dev/null || true", cfg.InstallDir, cfg.InstallDir, cfg.InstallDir))
}

// ════════════════════════════════════════════════════════════════
// ENV FILE GENERATION (FULL)
// ════════════════════════════════════════════════════════════════

func createEnvFile(cfg *Config) {
	adminNotifEnabled := "false"
	if cfg.AdminNotificationsChatID != "" {
		adminNotifEnabled = "true"
	}
	adminNotifChatID := cfg.AdminNotificationsChatID
	if adminNotifChatID == "" {
		adminNotifChatID = "-1001234567890"
	}

	if cfg.KeepExistingVolumes && cfg.OldPostgresPassword != "" {
		cfg.PostgresPassword = cfg.OldPostgresPassword
	}

	basicAuthLines := ""
	if cfg.RemnawaveAuthType == "basic_auth" {
		basicAuthLines = fmt.Sprintf("REMNAWAVE_USERNAME=%s\nREMNAWAVE_PASSWORD=%s", cfg.RemnawaveUsername, cfg.RemnawavePassword)
	}

	cabinetJWTSecret := generateToken()

	env := fmt.Sprintf(`# ===============================================
# REMNAWAVE BEDOLAGA BOT CONFIGURATION
# ===============================================
# Generated by bedolaga_installer v%s at %s
# ===============================================

# ===== TELEGRAM BOT =====
BOT_TOKEN=%s
ADMIN_IDS=%s
SUPPORT_USERNAME=%s

# ===== SUPPORT SYSTEM =====
SUPPORT_MENU_ENABLED=true
SUPPORT_SYSTEM_MODE=both
SUPPORT_TICKET_SLA_ENABLED=false
SUPPORT_TICKET_SLA_MINUTES=60
SUPPORT_TICKET_SLA_CHECK_INTERVAL_SECONDS=300
SUPPORT_TICKET_SLA_REMINDER_COOLDOWN_MINUTES=30

# ===== CABINET =====
CABINET_ENABLED=false
CABINET_URL=
CABINET_JWT_SECRET=%s
CABINET_ACCESS_TOKEN_EXPIRE_MINUTES=15
CABINET_REFRESH_TOKEN_EXPIRE_DAYS=7
CABINET_ALLOWED_ORIGINS=
CABINET_EMAIL_VERIFICATION_ENABLED=false
CABINET_EMAIL_AUTH_ENABLED=true

# ===== TEST EMAIL =====
TEST_EMAIL=
TEST_EMAIL_PASSWORD=
CABINET_EMAIL_VERIFICATION_EXPIRE_HOURS=24
CABINET_PASSWORD_RESET_EXPIRE_HOURS=1
CABINET_EMAIL_CHANGE_CODE_EXPIRE_MINUTES=15

# ===== SMTP =====
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM_EMAIL=
SMTP_FROM_NAME=VPN Service
SMTP_USE_TLS=true

# ===== NOTIFICATIONS =====
ADMIN_NOTIFICATIONS_ENABLED=%s
ADMIN_NOTIFICATIONS_CHAT_ID=%s
ADMIN_NOTIFICATIONS_TOPIC_ID=
ADMIN_NOTIFICATIONS_TICKET_TOPIC_ID=
ADMIN_NOTIFICATIONS_NALOG_TOPIC_ID=
ADMIN_REPORTS_ENABLED=false
ADMIN_REPORTS_CHAT_ID=
ADMIN_REPORTS_TOPIC_ID=
ADMIN_REPORTS_SEND_TIME=10:00

# ===== TRAFFIC MONITORING =====
TRAFFIC_FAST_CHECK_ENABLED=false
TRAFFIC_FAST_CHECK_INTERVAL_MINUTES=10
TRAFFIC_FAST_CHECK_THRESHOLD_GB=5.0
TRAFFIC_DAILY_CHECK_ENABLED=false
TRAFFIC_DAILY_CHECK_TIME=00:00
TRAFFIC_DAILY_THRESHOLD_GB=50.0
SUSPICIOUS_NOTIFICATIONS_TOPIC_ID=14
TRAFFIC_MONITORED_NODES=
TRAFFIC_IGNORED_NODES=
TRAFFIC_EXCLUDED_USER_UUIDS=
TRAFFIC_CHECK_BATCH_SIZE=1000
TRAFFIC_CHECK_CONCURRENCY=10
TRAFFIC_NOTIFICATION_COOLDOWN_MINUTES=60
TRAFFIC_SNAPSHOT_TTL_HOURS=24

# ===== BLACKLIST =====
BLACKLIST_CHECK_ENABLED=false
BLACKLIST_GITHUB_URL=https://raw.githubusercontent.com/BEDOLAGA-DEV/remnawave-bedolaga-telegram-bot/refs/heads/main/blacklist.txt
BLACKLIST_UPDATE_INTERVAL_HOURS=24
BLACKLIST_IGNORE_ADMINS=true
SUBSCRIPTION_RENEWAL_BALANCE_THRESHOLD_KOPEKS=20000

# ===== CHANNEL SUBSCRIPTION =====
CHANNEL_SUB_ID=
CHANNEL_IS_REQUIRED_SUB=false
CHANNEL_LINK=
CHANNEL_DISABLE_TRIAL_ON_UNSUBSCRIBE=true
CHANNEL_REQUIRED_FOR_ALL=false

# ===== DATABASE =====
DATABASE_MODE=auto
DATABASE_URL=
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=remnawave_bot
POSTGRES_USER=remnawave_user
POSTGRES_PASSWORD=%s
SQLITE_PATH=./data/bot.db
LOCALES_PATH=./locales

# ===== REDIS =====
REDIS_URL=redis://redis:6379/0
CART_TTL_SECONDS=3600

# ===== REMNAWAVE API =====
REMNAWAVE_API_URL=%s
REMNAWAVE_API_KEY=%s
REMNAWAVE_AUTH_TYPE=%s
REMNAWAVE_CADDY_TOKEN=
%s
REMNAWAVE_SECRET_KEY=%s
REMNAWAVE_USER_DESCRIPTION_TEMPLATE="Bot user: {full_name} {username}"
REMNAWAVE_USER_USERNAME_TEMPLATE="user_{telegram_id}"
REMNAWAVE_USER_DELETE_MODE=delete
REMNAWAVE_AUTO_SYNC_ENABLED=false
REMNAWAVE_AUTO_SYNC_TIMES=03:00

# ===== REMNAWAVE WEBHOOKS =====
REMNAWAVE_WEBHOOK_ENABLED=false
REMNAWAVE_WEBHOOK_PATH=/remnawave-webhook
REMNAWAVE_WEBHOOK_SECRET=

# ===== WEBHOOK NOTIFICATIONS =====
WEBHOOK_NOTIFY_USER_ENABLED=true
WEBHOOK_NOTIFY_SUB_STATUS=true
WEBHOOK_NOTIFY_SUB_EXPIRED=true
WEBHOOK_NOTIFY_SUB_EXPIRING=true
WEBHOOK_NOTIFY_SUB_LIMITED=true
WEBHOOK_NOTIFY_TRAFFIC_RESET=true
WEBHOOK_NOTIFY_SUB_DELETED=true
WEBHOOK_NOTIFY_SUB_REVOKED=true
WEBHOOK_NOTIFY_FIRST_CONNECTED=true
WEBHOOK_NOTIFY_NOT_CONNECTED=true
WEBHOOK_NOTIFY_BANDWIDTH_THRESHOLD=true
WEBHOOK_NOTIFY_DEVICES=true

# ===== SUBSCRIPTIONS =====
SALES_MODE=tariffs

# ===== TRIAL =====
TRIAL_DURATION_DAYS=3
TRIAL_TRAFFIC_LIMIT_GB=10
TRIAL_DEVICE_LIMIT=1
TRIAL_TARIFF_ID=0
TRIAL_PAYMENT_ENABLED=false
TRIAL_ACTIVATION_PRICE=0

# ===== PAID SUBSCRIPTION =====
DEFAULT_DEVICE_LIMIT=3
MAX_DEVICES_LIMIT=15
DEFAULT_TRAFFIC_LIMIT_GB=100
TRIAL_ADD_REMAINING_DAYS_TO_PAID=false

# ===== GLOBAL SUBSCRIPTION PARAMS =====
DEFAULT_TRAFFIC_RESET_STRATEGY=MONTH
RESET_TRAFFIC_ON_PAYMENT=false

# ===== TRAFFIC SETTINGS =====
TRAFFIC_SELECTION_MODE=selectable
FIXED_TRAFFIC_LIMIT_GB=100

# ===== TRAFFIC TOPUP =====
TRAFFIC_TOPUP_ENABLED=true
BUY_TRAFFIC_BUTTON_VISIBLE=true
TRAFFIC_TOPUP_PACKAGES_CONFIG=

# ===== TRAFFIC RESET =====
TRAFFIC_RESET_PRICE_MODE=traffic_with_purchased
TRAFFIC_RESET_BASE_PRICE=0

# ===== SUBSCRIPTION PERIODS =====
AVAILABLE_SUBSCRIPTION_PERIODS=30,90,180
AVAILABLE_RENEWAL_PERIODS=30,90,180

# ===== SIMPLE SUBSCRIPTION =====
SIMPLE_SUBSCRIPTION_ENABLED=true
SIMPLE_SUBSCRIPTION_PERIOD_DAYS=30
SIMPLE_SUBSCRIPTION_DEVICE_LIMIT=1
SIMPLE_SUBSCRIPTION_TRAFFIC_GB=0

# ===== PRICES (kopecks) =====
BASE_SUBSCRIPTION_PRICE=0
PRICE_14_DAYS=7000
PRICE_30_DAYS=10000
PRICE_60_DAYS=25900
PRICE_90_DAYS=36900
PRICE_180_DAYS=69900
PRICE_360_DAYS=109900

# ===== PROMO GROUP DISCOUNTS =====
BASE_PROMO_GROUP_PERIOD_DISCOUNTS_ENABLED=false
BASE_PROMO_GROUP_PERIOD_DISCOUNTS=60:10,90:20,180:40,360:70

# ===== TRAFFIC PACKAGES =====
TRAFFIC_PACKAGES_CONFIG="5:2000:false,10:3500:false,25:7000:false,50:11000:true,100:15000:true,250:17000:false,500:19000:false,1000:19500:true,0:0:true"
PRICE_TRAFFIC_UNLIMITED=20000

# ===== DEVICES =====
PRICE_PER_DEVICE=10000
DEVICES_SELECTION_ENABLED=true
DEVICES_SELECTION_DISABLED_AMOUNT=0

# ===== MODEM =====
MODEM_ENABLED=false
MODEM_PRICE_PER_MONTH=10000
MODEM_PERIOD_DISCOUNTS=3:15,6:20,12:25

DISABLE_WEB_PAGE_PREVIEW=false

# ===== REFERRAL SYSTEM =====
REFERRAL_PROGRAM_ENABLED=true
REFERRAL_MINIMUM_TOPUP_KOPEKS=10000
REFERRAL_FIRST_TOPUP_BONUS_KOPEKS=10000
REFERRAL_INVITER_BONUS_KOPEKS=10000
REFERRAL_COMMISSION_PERCENT=25
REFERRAL_NOTIFICATIONS_ENABLED=true
REFERRAL_NOTIFICATION_RETRY_ATTEMPTS=3

# ===== REFERRAL WITHDRAWAL =====
REFERRAL_WITHDRAWAL_ENABLED=false
REFERRAL_WITHDRAWAL_MIN_AMOUNT_KOPEKS=50000
REFERRAL_WITHDRAWAL_COOLDOWN_DAYS=30
REFERRAL_WITHDRAWAL_ONLY_REFERRAL_BALANCE=true
REFERRAL_WITHDRAWAL_NOTIFICATIONS_TOPIC_ID=0
REFERRAL_WITHDRAWAL_TEST_MODE=false
REFERRAL_WITHDRAWAL_SUSPICIOUS_MIN_DEPOSIT_KOPEKS=100000
REFERRAL_WITHDRAWAL_SUSPICIOUS_MAX_DEPOSITS_PER_MONTH=10
REFERRAL_WITHDRAWAL_SUSPICIOUS_NO_PURCHASES_RATIO=3

# ===== AUTOPAY =====
ENABLE_AUTOPAY=false
AUTOPAY_WARNING_DAYS=3,1
DEFAULT_AUTOPAY_ENABLED=true
DEFAULT_AUTOPAY_DAYS_BEFORE=3
MIN_BALANCE_FOR_AUTOPAY_KOPEKS=10000

# ===== PAYMENT SYSTEMS =====

# Telegram Stars
TELEGRAM_STARS_ENABLED=true
TELEGRAM_STARS_RATE_RUB=1.79

# Tribute
TRIBUTE_ENABLED=false
TRIBUTE_API_KEY=
TRIBUTE_DONATE_LINK=
TRIBUTE_WEBHOOK_PATH=/tribute-webhook
TRIBUTE_WEBHOOK_HOST=0.0.0.0
TRIBUTE_WEBHOOK_PORT=8081

# YooKassa
YOOKASSA_ENABLED=false
YOOKASSA_SHOP_ID=
YOOKASSA_SECRET_KEY=
YOOKASSA_RETURN_URL=
YOOKASSA_DEFAULT_RECEIPT_EMAIL=receipts@yourdomain.com
YOOKASSA_SBP_ENABLED=false
YOOKASSA_VAT_CODE=1
YOOKASSA_PAYMENT_MODE=full_payment
YOOKASSA_PAYMENT_SUBJECT=service
YOOKASSA_WEBHOOK_PATH=/yookassa-webhook
YOOKASSA_WEBHOOK_HOST=0.0.0.0
YOOKASSA_WEBHOOK_PORT=8082
YOOKASSA_MIN_AMOUNT_KOPEKS=5000
YOOKASSA_MAX_AMOUNT_KOPEKS=1000000
YOOKASSA_QUICK_AMOUNT_SELECTION_ENABLED=true
DISABLE_TOPUP_BUTTONS=false
SUPPORT_TOPUP_ENABLED=true
PAYMENT_VERIFICATION_AUTO_CHECK_ENABLED=false
PAYMENT_VERIFICATION_AUTO_CHECK_INTERVAL_MINUTES=10

# NaloGO
NALOGO_ENABLED=false
NALOGO_INN=
NALOGO_PASSWORD=
NALOGO_DEVICE_ID=
NALOGO_STORAGE_PATH=./nalogo_tokens.json
NALOGO_QUEUE_CHECK_INTERVAL=300
NALOGO_QUEUE_RECEIPT_DELAY=3
NALOGO_QUEUE_MAX_ATTEMPTS=10

# Payment descriptions
PAYMENT_SERVICE_NAME=Internet-service
PAYMENT_BALANCE_DESCRIPTION=Balance top-up
PAYMENT_SUBSCRIPTION_DESCRIPTION=Subscription payment
PAYMENT_BALANCE_TEMPLATE={service_name} - {description}
PAYMENT_SUBSCRIPTION_TEMPLATE={service_name} - {description}

# CryptoBot
CRYPTOBOT_ENABLED=false
CRYPTOBOT_API_TOKEN=
CRYPTOBOT_WEBHOOK_SECRET=
CRYPTOBOT_BASE_URL=https://pay.crypt.bot
CRYPTOBOT_TESTNET=false
CRYPTOBOT_WEBHOOK_PATH=/cryptobot-webhook
CRYPTOBOT_WEBHOOK_PORT=8081
CRYPTOBOT_DEFAULT_ASSET=USDT
CRYPTOBOT_ASSETS=USDT,TON,BTC,ETH,LTC,BNB,TRX,USDC
CRYPTOBOT_INVOICE_EXPIRES_HOURS=24

# Heleket
HELEKET_ENABLED=false
HELEKET_MERCHANT_ID=
HELEKET_API_KEY=
HELEKET_BASE_URL=https://api.heleket.com/v1
HELEKET_DEFAULT_CURRENCY=USDT
HELEKET_DEFAULT_NETWORK=
HELEKET_INVOICE_LIFETIME=3600
HELEKET_MARKUP_PERCENT=0
HELEKET_WEBHOOK_PATH=/heleket-webhook
HELEKET_WEBHOOK_HOST=0.0.0.0
HELEKET_WEBHOOK_PORT=8086
HELEKET_CALLBACK_URL=
HELEKET_RETURN_URL=
HELEKET_SUCCESS_URL=

# MulenPay
MULENPAY_ENABLED=false
MULENPAY_API_KEY=
MULENPAY_SECRET_KEY=
MULENPAY_SHOP_ID=
MULENPAY_BASE_URL=https://mulenpay.ru/api
MULENPAY_WEBHOOK_PATH=/mulenpay-webhook
MULENPAY_DISPLAY_NAME=Mulen Pay
MULENPAY_DESCRIPTION="Balance top-up"
MULENPAY_LANGUAGE=ru
MULENPAY_VAT_CODE=0
MULENPAY_PAYMENT_SUBJECT=4
MULENPAY_PAYMENT_MODE=4
MULENPAY_MIN_AMOUNT_KOPEKS=10000
MULENPAY_MAX_AMOUNT_KOPEKS=10000000

# PayPalych / PAL24
PAL24_ENABLED=false
PAL24_API_TOKEN=
PAL24_SHOP_ID=
PAL24_SIGNATURE_TOKEN=
PAL24_BASE_URL=https://pal24.pro/api/v1/
PAL24_WEBHOOK_PATH=/pal24-webhook
PAL24_PAYMENT_DESCRIPTION="Balance top-up"
PAL24_MIN_AMOUNT_KOPEKS=10000
PAL24_MAX_AMOUNT_KOPEKS=100000000
PAL24_REQUEST_TIMEOUT=30
PAL24_SBP_BUTTON_VISIBLE=true
PAL24_CARD_BUTTON_VISIBLE=true

# Platega
PLATEGA_ENABLED=false
PLATEGA_MERCHANT_ID=
PLATEGA_SECRET=
PLATEGA_BASE_URL=https://app.platega.io
PLATEGA_DISPLAY_NAME=Platega
PLATEGA_RETURN_URL=
PLATEGA_FAILED_URL=
PLATEGA_CURRENCY=RUB
PLATEGA_ACTIVE_METHODS=2,10,11,12,13
PLATEGA_MIN_AMOUNT_KOPEKS=100
PLATEGA_MAX_AMOUNT_KOPEKS=100000000
PLATEGA_WEBHOOK_PATH=/platega-webhook
PLATEGA_WEBHOOK_HOST=0.0.0.0
PLATEGA_WEBHOOK_PORT=8086

# Freekassa
FREEKASSA_ENABLED=false
FREEKASSA_SHOP_ID=
FREEKASSA_API_KEY=
FREEKASSA_SECRET_WORD_1=
FREEKASSA_SECRET_WORD_2=
FREEKASSA_DISPLAY_NAME=Freekassa
FREEKASSA_CURRENCY=RUB
FREEKASSA_MIN_AMOUNT_KOPEKS=10000
FREEKASSA_MAX_AMOUNT_KOPEKS=100000000
FREEKASSA_PAYMENT_TIMEOUT_SECONDS=3600
FREEKASSA_WEBHOOK_PATH=/freekassa-webhook
FREEKASSA_WEBHOOK_HOST=0.0.0.0
FREEKASSA_WEBHOOK_PORT=8088
FREEKASSA_PAYMENT_SYSTEM_ID=
FREEKASSA_USE_API=false

# Kassa AI
KASSA_AI_ENABLED=false
KASSA_AI_SHOP_ID=
KASSA_AI_API_KEY=
KASSA_AI_SECRET_WORD_2=
KASSA_AI_DISPLAY_NAME=KassaAI
KASSA_AI_CURRENCY=RUB
KASSA_AI_MIN_AMOUNT_KOPEKS=10000
KASSA_AI_MAX_AMOUNT_KOPEKS=100000000
KASSA_AI_WEBHOOK_PATH=/kassa-ai-webhook
KASSA_AI_WEBHOOK_HOST=0.0.0.0
KASSA_AI_WEBHOOK_PORT=8089
KASSA_AI_PAYMENT_SYSTEM_ID=44

# WATA
WATA_ENABLED=false
WATA_BASE_URL=https://api.wata.pro
WATA_ACCESS_TOKEN=
WATA_TERMINAL_PUBLIC_ID=
WATA_PAYMENT_DESCRIPTION=Balance top-up
WATA_PAYMENT_TYPE=all
WATA_SUCCESS_REDIRECT_URL=
WATA_FAIL_REDIRECT_URL=
WATA_LINK_TTL_MINUTES=60
WATA_MIN_AMOUNT_KOPEKS=10000
WATA_MAX_AMOUNT_KOPEKS=10000000
WATA_REQUEST_TIMEOUT=30
WATA_WEBHOOK_PATH=/wata-webhook
WATA_WEBHOOK_HOST=0.0.0.0
WATA_WEBHOOK_PORT=8087
WATA_PUBLIC_KEY_CACHE_SECONDS=3600

# CloudPayments
CLOUDPAYMENTS_ENABLED=false
CLOUDPAYMENTS_PUBLIC_ID=
CLOUDPAYMENTS_API_SECRET=
CLOUDPAYMENTS_API_URL=https://api.cloudpayments.ru
CLOUDPAYMENTS_WIDGET_URL=https://widget.cloudpayments.ru/show
CLOUDPAYMENTS_DESCRIPTION=Balance top-up
CLOUDPAYMENTS_CURRENCY=RUB
CLOUDPAYMENTS_MIN_AMOUNT_KOPEKS=10000
CLOUDPAYMENTS_MAX_AMOUNT_KOPEKS=10000000
CLOUDPAYMENTS_WEBHOOK_PATH=/cloudpayments-webhook
CLOUDPAYMENTS_WEBHOOK_HOST=0.0.0.0
CLOUDPAYMENTS_WEBHOOK_PORT=8089
CLOUDPAYMENTS_SKIN=mini
CLOUDPAYMENTS_REQUIRE_EMAIL=false
CLOUDPAYMENTS_TEST_MODE=false

# ===== INTERFACE & UX =====
ENABLE_LOGO_MODE=true
LOGO_FILE=vpn_logo.png
MAIN_MENU_MODE=default
MENU_LAYOUT_ENABLED=false
HIDE_SUBSCRIPTION_LINK=false
CONNECT_BUTTON_MODE=miniapp_subscription
MINIAPP_CUSTOM_URL=
MINIAPP_STATIC_PATH=miniapp
MINIAPP_SERVICE_NAME_EN=Bedolaga VPN
MINIAPP_SERVICE_NAME_RU=Bedolaga VPN
MINIAPP_SERVICE_DESCRIPTION_EN=Secure & Fast Connection
MINIAPP_SERVICE_DESCRIPTION_RU=Secure & Fast Connection
CONNECT_BUTTON_HAPP_DOWNLOAD_ENABLED=false
HAPP_DOWNLOAD_LINK_IOS=
HAPP_DOWNLOAD_LINK_ANDROID=
HAPP_DOWNLOAD_LINK_MACOS=
HAPP_DOWNLOAD_LINK_WINDOWS=
HAPP_DOWNLOAD_LINK_PC=
HAPP_CRYPTOLINK_REDIRECT_TEMPLATE=
SKIP_RULES_ACCEPT=false
SKIP_REFERRAL_CODE=false

# ===== MONITORING =====
MONITORING_INTERVAL=60
INACTIVE_USER_DELETE_MONTHS=3
TRIAL_WARNING_HOURS=2
ENABLE_NOTIFICATIONS=true
NOTIFICATION_RETRY_ATTEMPTS=3
MONITORING_LOGS_RETENTION_DAYS=30
NOTIFICATION_CACHE_HOURS=24

# ===== SERVER STATUS =====
SERVER_STATUS_MODE=disabled
SERVER_STATUS_EXTERNAL_URL=
SERVER_STATUS_METRICS_URL=
SERVER_STATUS_METRICS_USERNAME=
SERVER_STATUS_METRICS_PASSWORD=
SERVER_STATUS_METRICS_VERIFY_SSL=true
SERVER_STATUS_REQUEST_TIMEOUT=10
SERVER_STATUS_ITEMS_PER_PAGE=10

# ===== MAINTENANCE =====
MAINTENANCE_MODE=false
MAINTENANCE_CHECK_INTERVAL=30
MAINTENANCE_AUTO_ENABLE=true
MAINTENANCE_MONITORING_ENABLED=true
MAINTENANCE_RETRY_ATTEMPTS=1
MAINTENANCE_MESSAGE=Technical maintenance in progress. Service temporarily unavailable.

# ===== LOCALIZATION =====
DEFAULT_LANGUAGE=ru
AVAILABLE_LANGUAGES=ru,en,ua,zh,fa
LANGUAGE_SELECTION_ENABLED=true
PRICE_ROUNDING_ENABLED=true
TZ=Europe/Moscow

# ===== APP CONFIG =====
APP_CONFIG_PATH=app-config.json
ENABLE_DEEP_LINKS=true
APP_CONFIG_CACHE_TTL=3600

# ===== BAN SYSTEM =====
BAN_SYSTEM_ENABLED=false
BAN_SYSTEM_API_URL=
BAN_SYSTEM_API_TOKEN=
BAN_SYSTEM_REQUEST_TIMEOUT=30

# ===== BACKUPS =====
BACKUP_AUTO_ENABLED=true
BACKUP_INTERVAL_HOURS=24
BACKUP_TIME=03:00
BACKUP_MAX_KEEP=7
BACKUP_COMPRESSION=true
BACKUP_INCLUDE_LOGS=false
BACKUP_LOCATION=/app/data/backups
BACKUP_SEND_ENABLED=true
BACKUP_SEND_CHAT_ID=-100123456789
BACKUP_SEND_TOPIC_ID=
BACKUP_ARCHIVE_PASSWORD=

# ===== VERSION CHECK =====
VERSION_CHECK_ENABLED=true
VERSION_CHECK_REPO=BEDOLAGA-DEV/remnawave-bedolaga-telegram-bot
VERSION_CHECK_INTERVAL_HOURS=1

# ===== LOGGING =====
LOG_LEVEL=INFO
LOG_FILE=logs/bot.log
LOG_ROTATION_ENABLED=false
LOG_ROTATION_TIME=00:00
LOG_ROTATION_KEEP_DAYS=7
LOG_ROTATION_COMPRESS=true
LOG_ROTATION_SEND_TO_TELEGRAM=false
LOG_ROTATION_CHAT_ID=
LOG_ROTATION_TOPIC_ID=
LOG_DIR=logs
LOG_INFO_FILE=info.log
LOG_WARNING_FILE=warning.log
LOG_ERROR_FILE=error.log
LOG_PAYMENTS_FILE=payments.log

# ===== DEVELOPMENT =====
DEBUG=false
WEBHOOK_URL=%s
WEBHOOK_PATH=/webhook
WEBHOOK_SECRET_TOKEN=%s
WEBHOOK_DROP_PENDING_UPDATES=true
WEBHOOK_MAX_QUEUE_SIZE=1024
WEBHOOK_WORKERS=4
WEBHOOK_ENQUEUE_TIMEOUT=0.1
WEBHOOK_WORKER_SHUTDOWN_TIMEOUT=30.0
BOT_RUN_MODE=%s

# ===== CONTESTS =====
CONTESTS_ENABLED=false
CONTESTS_BUTTON_VISIBLE=false
REFERRAL_CONTESTS_ENABLED=false

# ===== AUTO PURCHASE =====
AUTO_PURCHASE_AFTER_TOPUP_ENABLED=false

# ===== ACTIVATE BUTTON =====
ACTIVATE_BUTTON_VISIBLE=false

# ===== WEB API =====
WEB_API_ENABLED=%s
WEB_API_HOST=0.0.0.0
WEB_API_PORT=8080
WEB_API_WORKERS=1
WEB_API_ALLOWED_ORIGINS=*
WEB_API_DOCS_ENABLED=false
WEB_API_TITLE=Remnawave Bot Admin API
WEB_API_VERSION=1.0.0
WEB_API_DEFAULT_TOKEN=%s
WEB_API_DEFAULT_TOKEN_NAME=Bootstrap Token
WEB_API_TOKEN_HASH_ALGORITHM=sha256
WEB_API_REQUEST_LOGGING=true

MINIAPP_STATIC_PATH=miniapp
`,
		appVersion, time.Now().Format("2006-01-02 15:04:05"),
		cfg.BotToken, cfg.AdminIDs, cfg.SupportUsername,
		cabinetJWTSecret,
		adminNotifEnabled, adminNotifChatID,
		cfg.PostgresPassword,
		cfg.RemnawaveAPIURL, cfg.RemnawaveAPIKey, cfg.RemnawaveAuthType,
		basicAuthLines, cfg.RemnawaveSecretKey,
		cfg.WebhookURL, cfg.WebhookSecretToken, cfg.BotRunMode,
		cfg.WebAPIEnabled, cfg.WebAPIDefaultToken,
	)

	envPath := filepath.Join(cfg.InstallDir, ".env")
	if err := os.WriteFile(envPath, []byte(env), 0600); err != nil {
		printError("Ошибка записи .env: " + err.Error())
		os.Exit(1)
	}
	printSuccess("Файл .env создан (" + envPath + ")")
}

// ════════════════════════════════════════════════════════════════
// DOCKER COMPOSE GENERATION
// ════════════════════════════════════════════════════════════════

func createStandaloneCompose(cfg *Config) {
	content := `services:
  postgres:
    image: postgres:15-alpine
    container_name: remnawave_bot_db
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-remnawave_bot}
      POSTGRES_USER: ${POSTGRES_USER:-remnawave_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-secure_password_123}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --locale=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - bot_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-remnawave_user} -d ${POSTGRES_DB:-remnawave_bot}"]
      interval: 30s
      timeout: 5s
      retries: 5
      start_period: 30s

  redis:
    image: redis:7-alpine
    container_name: remnawave_bot_redis
    restart: unless-stopped
    command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    networks:
      - bot_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  bot:
    build: .
    container_name: remnawave_bot
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    env_file:
      - .env
    environment:
      DOCKER_ENV: "true"
      DATABASE_MODE: "auto"
      POSTGRES_HOST: "postgres"
      POSTGRES_PORT: "5432"
      POSTGRES_DB: "${POSTGRES_DB:-remnawave_bot}"
      POSTGRES_USER: "${POSTGRES_USER:-remnawave_user}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:-secure_password_123}"
      REDIS_URL: "redis://redis:6379/0"
      TZ: "Europe/Moscow"
      LOCALES_PATH: "${LOCALES_PATH:-/app/locales}"
    volumes:
      - ./logs:/app/logs:rw
      - ./data:/app/data:rw
      - ./locales:/app/locales:rw
      - /etc/timezone:/etc/timezone:ro
      - /etc/localtime:/etc/localtime:ro
      - ./vpn_logo.png:/app/vpn_logo.png:ro
    ports:
      - "${WEB_API_PORT:-8080}:8080"
    networks:
      - bot_network
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://localhost:8080/health || exit 1"]
      interval: 60s
      timeout: 10s
      retries: 3
      start_period: 30s

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  bot_network:
    name: remnawave_bot_network
    driver: bridge
`
	os.WriteFile(filepath.Join(cfg.InstallDir, "docker-compose.yml"), []byte(content), 0644)
}

func createLocalCompose(cfg *Config) {
	networkName := cfg.DockerNetwork
	if networkName == "" {
		networkName = "remnawave-network"
	}
	content := fmt.Sprintf(`services:
  postgres:
    image: postgres:15-alpine
    container_name: remnawave_bot_db
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-remnawave_bot}
      POSTGRES_USER: ${POSTGRES_USER:-remnawave_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-secure_password_123}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --locale=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - bot_network
      - remnawave_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-remnawave_user} -d ${POSTGRES_DB:-remnawave_bot}"]
      interval: 30s
      timeout: 5s
      retries: 5
      start_period: 30s

  redis:
    image: redis:7-alpine
    container_name: remnawave_bot_redis
    restart: unless-stopped
    command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    networks:
      - bot_network
      - remnawave_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  bot:
    build: .
    container_name: remnawave_bot
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    env_file:
      - .env
    environment:
      DOCKER_ENV: "true"
      DATABASE_MODE: "auto"
      POSTGRES_HOST: "postgres"
      POSTGRES_PORT: "5432"
      POSTGRES_DB: "${POSTGRES_DB:-remnawave_bot}"
      POSTGRES_USER: "${POSTGRES_USER:-remnawave_user}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:-secure_password_123}"
      REDIS_URL: "redis://redis:6379/0"
      TZ: "Europe/Moscow"
      LOCALES_PATH: "${LOCALES_PATH:-/app/locales}"
    volumes:
      - ./logs:/app/logs:rw
      - ./data:/app/data:rw
      - ./locales:/app/locales:rw
      - /etc/timezone:/etc/timezone:ro
      - /etc/localtime:/etc/localtime:ro
      - ./vpn_logo.png:/app/vpn_logo.png:ro
    ports:
      - "${WEB_API_PORT:-8080}:8080"
    networks:
      - bot_network
      - remnawave_network
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://localhost:8080/health || exit 1"]
      interval: 60s
      timeout: 10s
      retries: 3
      start_period: 30s

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  bot_network:
    name: remnawave_bot_network
    driver: bridge
  remnawave_network:
    name: %s
    external: true
`, networkName)
	os.WriteFile(filepath.Join(cfg.InstallDir, "docker-compose.local.yml"), []byte(content), 0644)
}

// ════════════════════════════════════════════════════════════════
// NGINX SETUP
// ════════════════════════════════════════════════════════════════

func setupNginxSystem(cfg *Config) {
	installNginx()

	nginxAvail := "/etc/nginx/sites-available"
	nginxEnabled := "/etc/nginx/sites-enabled"
	os.MkdirAll(nginxAvail, 0755)
	os.MkdirAll(nginxEnabled, 0755)

	if cfg.WebhookDomain != "" {
		conf := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    client_max_body_size 32m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_buffering off;
        proxy_request_buffering off;
    }

    location = /app-config.json {
        add_header Access-Control-Allow-Origin "*";
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
`, cfg.WebhookDomain)
		os.WriteFile(filepath.Join(nginxAvail, "bedolaga-webhook"), []byte(conf), 0644)
		os.Remove(filepath.Join(nginxEnabled, "bedolaga-webhook"))
		os.Symlink(filepath.Join(nginxAvail, "bedolaga-webhook"), filepath.Join(nginxEnabled, "bedolaga-webhook"))
	}

	if cfg.MiniappDomain != "" {
		conf := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    client_max_body_size 32m;
    root %s/miniapp;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
        expires 1h;
        add_header Cache-Control "public";
    }

    location /miniapp/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }

    location = /app-config.json {
        add_header Access-Control-Allow-Origin "*";
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
`, cfg.MiniappDomain, cfg.InstallDir)
		os.WriteFile(filepath.Join(nginxAvail, "bedolaga-miniapp"), []byte(conf), 0644)
		os.Remove(filepath.Join(nginxEnabled, "bedolaga-miniapp"))
		os.Symlink(filepath.Join(nginxAvail, "bedolaga-miniapp"), filepath.Join(nginxEnabled, "bedolaga-miniapp"))
	}

	runShellSilent("nginx -t && systemctl reload nginx")
	printSuccess("Nginx настроен")
}

func setupNginxPanel(cfg *Config) {
	panelNginxConf := filepath.Join(cfg.PanelDir, "nginx.conf")
	if !fileExists(panelNginxConf) {
		printWarning("nginx.conf панели не найден, переключаемся на системный nginx")
		setupNginxSystem(cfg)
		return
	}

	runShellSilent(fmt.Sprintf(`cp "%s" "%s.backup.$(date +%%Y%%m%%d_%%H%%M%%S)"`, panelNginxConf, panelNginxConf))
	runShellSilent(fmt.Sprintf(`sed -i '/# === BEGIN Bedolaga Bot ===/,/# === END Bedolaga Bot ===/d' "%s"`, panelNginxConf))

	block := "\n# === BEGIN Bedolaga Bot ===\n"
	if cfg.WebhookDomain != "" {
		block += fmt.Sprintf(`server {
    server_name %s;
    listen 443 ssl;
    http2 on;
    ssl_certificate "/etc/letsencrypt/live/%s/fullchain.pem";
    ssl_certificate_key "/etc/letsencrypt/live/%s/privkey.pem";
    client_max_body_size 32m;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_buffering off;
    }
}
`, cfg.WebhookDomain, cfg.WebhookDomain, cfg.WebhookDomain)
	}
	if cfg.MiniappDomain != "" {
		block += fmt.Sprintf(`server {
    server_name %s;
    listen 443 ssl;
    http2 on;
    ssl_certificate "/etc/letsencrypt/live/%s/fullchain.pem";
    ssl_certificate_key "/etc/letsencrypt/live/%s/privkey.pem";
    client_max_body_size 32m;
    location /miniapp/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }
    location = /app-config.json {
        add_header Access-Control-Allow-Origin "*";
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    location / {
        root /var/www/remnawave-miniapp;
        try_files $uri $uri/ /index.html;
        expires 1h;
        add_header Cache-Control "public, immutable";
    }
}
`, cfg.MiniappDomain, cfg.MiniappDomain, cfg.MiniappDomain)
	}
	block += "# === END Bedolaga Bot ===\n"

	f, err := os.OpenFile(panelNginxConf, os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(block)
		f.Close()
	}

	runShellSilent(fmt.Sprintf("cd %s && docker compose up -d remnawave-nginx 2>/dev/null || docker restart remnawave-nginx 2>/dev/null || true", cfg.PanelDir))
	printSuccess("Nginx панели обновлён")
}

// ════════════════════════════════════════════════════════════════
// CADDY SETUP
// ════════════════════════════════════════════════════════════════

func setupCaddy(cfg *Config) {
	installCaddy()

	caddyFile := "/etc/caddy/Caddyfile"
	if fileExists(caddyFile) {
		runShellSilent(fmt.Sprintf(`cp "%s" "%s.backup.$(date +%%Y%%m%%d_%%H%%M%%S)"`, caddyFile, caddyFile))
		runShellSilent(fmt.Sprintf(`sed -i '/# === BEGIN Bedolaga Bot ===/,/# === END Bedolaga Bot ===/d' "%s"`, caddyFile))
	}

	block := "\n# === BEGIN Bedolaga Bot ===\n"
	if cfg.WebhookDomain != "" {
		block += fmt.Sprintf(`%s {
    reverse_proxy localhost:8080
}

`, cfg.WebhookDomain)
	}
	if cfg.MiniappDomain != "" {
		block += fmt.Sprintf(`%s {
    @api path /miniapp/*
    reverse_proxy @api localhost:8080
    @config path /app-config.json
    reverse_proxy @config localhost:8080
    header @config Access-Control-Allow-Origin *
    root * %s/miniapp
    try_files {path} {path}/ /index.html
    file_server
}

`, cfg.MiniappDomain, cfg.InstallDir)
	}
	block += "# === END Bedolaga Bot ===\n"

	f, _ := os.OpenFile(caddyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		f.WriteString(block)
		f.Close()
	}

	runShellSilent("systemctl enable caddy 2>/dev/null || true")
	runShellSilent("systemctl reload caddy 2>/dev/null || systemctl restart caddy 2>/dev/null || true")
	printSuccess("Caddy настроен (автоматический HTTPS)")
}

// ════════════════════════════════════════════════════════════════
// SSL SETUP
// ════════════════════════════════════════════════════════════════

func setupSSL(cfg *Config) {
	if cfg.ReverseProxyType == "caddy" || cfg.ReverseProxyType == "skip" {
		return
	}
	if cfg.WebhookDomain == "" && cfg.MiniappDomain == "" {
		return
	}

	if !confirmPrompt("Получить SSL-сертификаты сейчас?", true) {
		printInfo("Вы можете получить сертификаты позже: certbot --nginx -d yourdomain.com")
		return
	}

	cfg.SSLEmail = inputText("Email Let's Encrypt", "admin@example.com", "Email для уведомлений о SSL-сертификатах", true)

	isPanelMode := cfg.ReverseProxyType == "nginx_panel"

	for _, domain := range []string{cfg.WebhookDomain, cfg.MiniappDomain} {
		if domain == "" {
			continue
		}
		runWithSpinner("Получение SSL для "+domain+"...", func() error {
			if isPanelMode {
				runShellSilent("docker stop remnawave-nginx 2>/dev/null || true")
				runShellSilent("systemctl stop nginx 2>/dev/null || true")
				time.Sleep(2 * time.Second)
				err := runShell(fmt.Sprintf("certbot certonly --standalone -d %s --email %s --agree-tos --non-interactive", domain, cfg.SSLEmail))
				runShellSilent("docker start remnawave-nginx 2>/dev/null || true")
				runShellSilent("systemctl start nginx 2>/dev/null || true")
				return err
			}
			return runShell(fmt.Sprintf("certbot --nginx -d %s --email %s --agree-tos --non-interactive", domain, cfg.SSLEmail))
		})
	}

	runShellSilent("systemctl enable certbot.timer 2>/dev/null || true")
	runShellSilent("systemctl start certbot.timer 2>/dev/null || true")
}

// ════════════════════════════════════════════════════════════════
// DOCKER START
// ════════════════════════════════════════════════════════════════

func startDocker(cfg *Config) {
	runShellSilent(fmt.Sprintf("cd %s && docker compose down 2>/dev/null || true", cfg.InstallDir))
	runShellSilent(fmt.Sprintf("cd %s && docker compose -f docker-compose.local.yml down 2>/dev/null || true", cfg.InstallDir))

	composeFile := "docker-compose.yml"

	if cfg.PanelInstalledLocally {
		if cfg.DockerNetwork != "" {
			runShellSilent(fmt.Sprintf("docker network create %s 2>/dev/null || true", cfg.DockerNetwork))
		}
		createLocalCompose(cfg)
		composeFile = "docker-compose.local.yml"
	} else {
		if !fileExists(filepath.Join(cfg.InstallDir, "docker-compose.yml")) {
			createStandaloneCompose(cfg)
		}
	}

	runWithSpinner("Сборка и запуск контейнеров...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s up -d --build 2>&1", cfg.InstallDir, composeFile))
		return err
	})

	printInfo("Ожидание контейнеров...")
	time.Sleep(8 * time.Second)

	// Show status
	out, _ := runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s ps --format 'table {{.Name}}\\t{{.Status}}' 2>/dev/null", cfg.InstallDir, composeFile))
	if out != "" {
		fmt.Println()
		fmt.Println(dimStyle.Render("  " + strings.ReplaceAll(out, "\n", "\n  ")))
		fmt.Println()
	}

	if cfg.PanelInstalledLocally && cfg.DockerNetwork != "" {
		ensureNetworkConnection(cfg)
		verifyPanelConnection()
	}
}

func ensureNetworkConnection(cfg *Config) {
	net := cfg.DockerNetwork
	containers := []string{"remnawave_bot", "remnawave_bot_db", "remnawave_bot_redis"}
	for _, c := range containers {
		out, _ := runShellSilent(fmt.Sprintf("docker ps --format '{{.Names}}' | grep '^%s$'", c))
		if out == "" {
			continue
		}
		nets, _ := runShellSilent(fmt.Sprintf(`docker inspect %s --format '{{range $net, $_ := .NetworkSettings.Networks}}{{$net}} {{end}}'`, c))
		if !strings.Contains(nets, net) {
			runShellSilent(fmt.Sprintf("docker network connect %s %s 2>/dev/null", net, c))
		}
	}
}

func verifyPanelConnection() {
	time.Sleep(3 * time.Second)
	if out, err := runShellSilent("docker exec remnawave_bot getent hosts remnawave 2>/dev/null | awk '{print $1}'"); err == nil && out != "" {
		printSuccess("Подключение к панели проверено: remnawave -> " + out + ":3000")
	} else {
		printWarning("Не удаётся разрешить 'remnawave' — проверьте сетевое подключение вручную")
	}
}

// ════════════════════════════════════════════════════════════════
// FIREWALL (optional)
// ════════════════════════════════════════════════════════════════

func setupFirewall() {
	if !confirmPrompt("Настроить Firewall (UFW)?", false) {
		return
	}
	runWithSpinner("Настройка firewall...", func() error {
		if !commandExists("ufw") {
			runShellSilent("apt-get install -y ufw")
		}
		runShellSilent("ufw --force reset")
		runShellSilent("ufw default deny incoming")
		runShellSilent("ufw default allow outgoing")
		runShellSilent("ufw allow 22/tcp")
		runShellSilent("ufw allow 80/tcp")
		runShellSilent("ufw allow 443/tcp")
		runShellSilent("ufw --force enable")
		return nil
	})
}

// ════════════════════════════════════════════════════════════════
// MANAGEMENT SCRIPT
// ════════════════════════════════════════════════════════════════

func createManagementScript(cfg *Config) {
	composeFile := "docker-compose.yml"
	if cfg.PanelInstalledLocally {
		composeFile = "docker-compose.local.yml"
	}

	script := fmt.Sprintf(`#!/bin/bash
# ╔══════════════════════════════════════════════════════════════╗
# ║   REMNAWAVE BEDOLAGA BOT — MANAGEMENT                       ║
# ╚══════════════════════════════════════════════════════════════╝

INSTALL_DIR="%s"
COMPOSE_FILE="%s"

# Colors
P='\033[0;35m'   # Purple
G='\033[0;32m'   # Green
R='\033[0;31m'   # Red
Y='\033[1;33m'   # Yellow
C='\033[0;36m'   # Cyan
W='\033[1;37m'   # White
D='\033[0;90m'   # Dim
A='\033[38;5;214m' # Amber
NC='\033[0m'

check_dir() { [ ! -d "$INSTALL_DIR" ] && echo -e "${R}  x Bot not found: $INSTALL_DIR${NC}" && exit 1; cd "$INSTALL_DIR"; }

do_logs()    { check_dir; echo -e "${C}  Streaming logs (Ctrl+C to stop)...${NC}"; docker compose -f "$COMPOSE_FILE" logs -f --tail=150 bot; }
do_status()  { check_dir; echo; echo -e "${W}  Container Status${NC}"; echo -e "${D}  --------------------------------------------------${NC}"; docker compose -f "$COMPOSE_FILE" ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null; echo; echo -e "${W}  Resource Usage${NC}"; echo -e "${D}  --------------------------------------------------${NC}"; docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | grep -E "remnawave|postgres|redis" || echo "  No data"; }
do_restart() { check_dir; echo -e "${C}  Restarting...${NC}"; docker compose -f "$COMPOSE_FILE" restart && echo -e "${G}  ✓ Restarted${NC}"; }
do_start()   { check_dir; echo -e "${C}  Starting...${NC}"; docker compose -f "$COMPOSE_FILE" up -d && echo -e "${G}  ✓ Started${NC}"; }
do_stop()    { check_dir; echo -e "${C}  Stopping...${NC}"; docker compose -f "$COMPOSE_FILE" down && echo -e "${G}  ✓ Stopped${NC}"; }

do_update() {
    check_dir
    echo -e "${A}  Updating bot...${NC}"
    cp .env ".env.backup_$(date +%%Y%%m%%d_%%H%%M%%S)" 2>/dev/null
    echo -e "${C}  Pulling latest code...${NC}"
    git pull origin main
    echo -e "${C}  Rebuilding containers...${NC}"
    docker compose -f "$COMPOSE_FILE" down && docker compose -f "$COMPOSE_FILE" up -d --build && docker compose -f "$COMPOSE_FILE" logs -f -t
}

do_backup() {
    check_dir
    local BK="$INSTALL_DIR/data/backups/$(date +%%Y%%m%%d_%%H%%M%%S)"
    mkdir -p "$BK"
    echo -e "${C}  Creating backup...${NC}"
    docker compose -f "$COMPOSE_FILE" exec -T postgres pg_dump -U remnawave_user remnawave_bot > "$BK/database.sql" 2>/dev/null && echo -e "${G}  ✓ database.sql${NC}" || echo -e "${R}  x DB backup failed${NC}"
    cp .env "$BK/.env" 2>/dev/null && echo -e "${G}  ✓ .env${NC}"
    cp docker-compose*.yml "$BK/" 2>/dev/null
    echo -e "${G}  ✓ Backup: $BK${NC}"
    ls -dt "$INSTALL_DIR/data/backups"/*/  2>/dev/null | tail -n +6 | xargs -r rm -rf
}

do_health() {
    check_dir
    echo
    echo -e "${A}  ╔══════════════════════════════════╗${NC}"
    echo -e "${A}  ║     SYSTEM DIAGNOSTICS            ║${NC}"
    echo -e "${A}  ╚══════════════════════════════════╝${NC}"
    echo
    docker ps --format '{{.Names}}' | grep -q "remnawave_bot$" && echo -e "${G}  ✓ Bot: running${NC}" || echo -e "${R}  x Bot: not running${NC}"
    docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U remnawave_user >/dev/null 2>&1 && echo -e "${G}  ✓ PostgreSQL: healthy${NC}" || echo -e "${R}  x PostgreSQL: down${NC}"
    docker compose -f "$COMPOSE_FILE" exec -T redis redis-cli ping >/dev/null 2>&1 && echo -e "${G}  ✓ Redis: healthy${NC}" || echo -e "${R}  x Redis: down${NC}"
    echo
    echo -e "${D}  Last 10 log lines:${NC}"
    docker compose -f "$COMPOSE_FILE" logs --tail=10 bot 2>/dev/null
}

do_config()    { check_dir; ${EDITOR:-nano} "$INSTALL_DIR/.env"; echo -e "${Y}  Restart to apply: bot restart${NC}"; }

do_uninstall() {
    check_dir
    echo -e "${R}  ╔══════════════════════════════════╗${NC}"
    echo -e "${R}  ║       BOT REMOVAL                 ║${NC}"
    echo -e "${R}  ╚══════════════════════════════════╝${NC}"
    echo -e "${Y}  This will stop and remove bot containers.${NC}"
    read -p "  Type 'yes' to confirm: " CONFIRM
    [ "$CONFIRM" != "yes" ] && echo "  Cancelled" && return
    docker compose -f "$COMPOSE_FILE" down
    read -p "  Delete data (volumes)? (y/n): " -n 1 -r; echo
    [[ $REPLY =~ ^[Yy]$ ]] && docker compose -f "$COMPOSE_FILE" down -v
    [ -f "/usr/local/bin/bot" ] && rm -f /usr/local/bin/bot && echo -e "${G}  ✓ 'bot' command removed${NC}"
    echo -e "${G}  ✓ Removal complete${NC}"
    echo -e "${Y}  Directory preserved: $INSTALL_DIR${NC}"
}

show_menu() {
    clear
    echo -e "${P}"
    echo "  ╔══════════════════════════════════════════════════════╗"
    echo "  ║                                                      ║"
    echo -e "  ║  ${A} REMNAWAVE BEDOLAGA BOT${P}                          ║"
    echo "  ║                                                      ║"
    echo "  ╚══════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "  ${D}Directory:${NC} ${C}$INSTALL_DIR${NC}"
    docker ps --format '{{.Names}}' | grep -q "remnawave_bot$" && echo -e "  ${D}Status:${NC}    ${G}● Running${NC}" || echo -e "  ${D}Status:${NC}    ${R}○ Stopped${NC}"
    echo
    echo -e "  ${D}─────────────────────────────────────────${NC}"
    echo
    echo -e "  ${A}1${NC}  Logs              ${A}6${NC}  Backup"
    echo -e "  ${A}2${NC}  Status            ${A}7${NC}  Health Check"
    echo -e "  ${A}3${NC}  Restart           ${A}8${NC}  Edit Config"
    echo -e "  ${A}4${NC}  Start             ${A}9${NC}  Update Bot"
    echo -e "  ${A}5${NC}  Stop              ${A}0${NC}  Uninstall"
    echo
    echo -e "  ${D}─────────────────────────────────────────${NC}"
    echo -e "  ${D}q${NC}  Exit"
    echo
}

interactive_menu() {
    while true; do
        show_menu
        read -p "  $(echo -e ${A})>${NC} " choice
        case $choice in
            1) do_logs ;; 2) do_status; read -p "  Press Enter..." ;; 3) do_restart; read -p "  Press Enter..." ;;
            4) do_start; read -p "  Press Enter..." ;; 5) do_stop; read -p "  Press Enter..." ;; 6) do_backup; read -p "  Press Enter..." ;;
            7) do_health; read -p "  Press Enter..." ;; 8) do_config ;; 9) do_update; read -p "  Press Enter..." ;;
            0) do_uninstall; break ;; q|Q) echo -e "  ${D}Bye!${NC}"; exit 0 ;; *) echo -e "${R}  Invalid choice${NC}"; sleep 0.5 ;;
        esac
    done
}

show_help() {
    echo -e "${A}  REMNAWAVE BEDOLAGA BOT${NC}"
    echo -e "  ${D}Usage: bot [command]${NC}"
    echo
    echo -e "  ${W}(no args)${NC}  Interactive menu"
    echo -e "  ${W}logs${NC}       View bot logs"
    echo -e "  ${W}status${NC}     Container status"
    echo -e "  ${W}restart${NC}    Restart bot"
    echo -e "  ${W}start${NC}      Start bot"
    echo -e "  ${W}stop${NC}       Stop bot"
    echo -e "  ${W}update${NC}     Update (git pull + rebuild)"
    echo -e "  ${W}backup${NC}     Create backup"
    echo -e "  ${W}health${NC}     System diagnostics"
    echo -e "  ${W}config${NC}     Edit .env"
    echo -e "  ${W}uninstall${NC}  Remove bot"
}

case "$1" in
    logs) do_logs ;; status) do_status ;; restart) do_restart ;; start) do_start ;; stop) do_stop ;;
    update|upgrade) do_update ;; backup) do_backup ;; health|check) do_health ;; config|edit) do_config ;;
    uninstall|remove) do_uninstall ;; help|--help|-h) show_help ;; "") interactive_menu ;;
    *) echo -e "${R}  Unknown: $1${NC}"; echo "  Use: bot help"; exit 1 ;;
esac
`, cfg.InstallDir, composeFile)

	os.WriteFile("/usr/local/bin/bot", []byte(script), 0755)
	printSuccess("Команда управления 'bot' установлена")
}

// ════════════════════════════════════════════════════════════════
// FINAL INFO
// ════════════════════════════════════════════════════════════════

func printFinalInfo(cfg *Config) {
	sep := lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Repeat("═", 56))

	var b strings.Builder
	b.WriteString(successStyle.Render("  УСТАНОВКА ЗАВЕРШЕНА") + "\n\n")
	b.WriteString(highlightStyle.Render("  Каталог: ") + infoStyle.Render(cfg.InstallDir) + "\n")
	b.WriteString(highlightStyle.Render("  Конфиг:  ") + infoStyle.Render(cfg.InstallDir+"/.env") + "\n\n")

	b.WriteString(highlightStyle.Render("  Управление:") + "\n")
	b.WriteString(dimStyle.Render("    bot          ") + "Интерактивное меню\n")
	b.WriteString(dimStyle.Render("    bot logs     ") + "Просмотр логов\n")
	b.WriteString(dimStyle.Render("    bot status   ") + "Статус контейнеров\n")
	b.WriteString(dimStyle.Render("    bot update   ") + "Обновить бота\n\n")

	if cfg.WebhookDomain != "" {
		b.WriteString(highlightStyle.Render("  Webhook: ") + infoStyle.Render("https://"+cfg.WebhookDomain) + "\n")
	}
	if cfg.MiniappDomain != "" {
		b.WriteString(highlightStyle.Render("  MiniApp: ") + infoStyle.Render("https://"+cfg.MiniappDomain) + "\n")
	}

	fmt.Println("\n  " + sep)
	fmt.Println(successBoxStyle.Render(b.String()))
	fmt.Println("  " + sep)
}

// ════════════════════════════════════════════════════════════════
// UPDATE / UNINSTALL (standalone commands)
// ════════════════════════════════════════════════════════════════

func findInstallDir() string {
	paths := []string{"/opt/remnawave-bedolaga-telegram-bot", "/root/remnawave-bedolaga-telegram-bot"}
	for _, p := range paths {
		if dirExists(p) {
			return p
		}
	}
	cwd, _ := os.Getwd()
	if fileExists(filepath.Join(cwd, "docker-compose.yml")) && fileExists(filepath.Join(cwd, ".env")) {
		return cwd
	}
	return ""
}

func detectComposeFile(installDir string) string {
	if fileExists(filepath.Join(installDir, "docker-compose.local.yml")) {
		return "docker-compose.local.yml"
	}
	return "docker-compose.yml"
}

func updateBot() {
	printBanner()
	installDir := findInstallDir()
	if installDir == "" {
		printErrorBox(errorStyle.Render("Установка бота не найдена!"))
		os.Exit(1)
	}
	composeFile := detectComposeFile(installDir)
	printInfo("Каталог: " + installDir)

	if !confirmPrompt("Начать обновление?", true) {
		os.Exit(0)
	}

	runShellSilent(fmt.Sprintf(`cd %s && cp .env ".env.backup_$(date +%%Y%%m%%d_%%H%%M%%S)" 2>/dev/null || true`, installDir))

	runWithSpinner("Загрузка последнего кода...", func() error {
		_, err := runShellSilent(fmt.Sprintf("cd %s && git pull origin main", installDir))
		return err
	})

	printInfo("Пересборка и перезапуск...")
	runShell(fmt.Sprintf("cd %s && docker compose -f %s down && docker compose -f %s up -d --build && docker compose -f %s logs -f -t", installDir, composeFile, composeFile, composeFile))
}

func uninstallBot() {
	printBanner()
	installDir := findInstallDir()
	if installDir == "" {
		printErrorBox(errorStyle.Render("Бот не установлен!"))
		os.Exit(1)
	}
	composeFile := detectComposeFile(installDir)
	printInfo("Каталог: " + installDir)

	val := inputText("Введите 'yes' для подтверждения удаления", "", "Это остановит и удалит контейнеры бота", true)
	if val != "yes" {
		printSuccess("Отменено")
		return
	}

	if confirmPrompt("Создать резервную копию сначала?", true) {
		runShellSilent(fmt.Sprintf(`cd %s && tar -czf "/root/bedolaga_backup_$(date +%%Y%%m%%d_%%H%%M%%S).tar.gz" .env data/ 2>/dev/null || true`, installDir))
		printSuccess("Резервная копия сохранена в /root/")
	}

	runWithSpinner("Остановка контейнеров...", func() error {
		runShellSilent(fmt.Sprintf("cd %s && docker compose -f %s down -v 2>/dev/null || docker compose down -v 2>/dev/null || true", installDir, composeFile))
		return nil
	})

	runShellSilent("rm -f /etc/nginx/sites-enabled/bedolaga-webhook /etc/nginx/sites-enabled/bedolaga-miniapp")
	runShellSilent("rm -f /etc/nginx/sites-available/bedolaga-webhook /etc/nginx/sites-available/bedolaga-miniapp")
	runShellSilent("nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true")
	if fileExists("/etc/caddy/Caddyfile") {
		runShellSilent(`sed -i '/# === BEGIN Bedolaga Bot ===/,/# === END Bedolaga Bot ===/d' /etc/caddy/Caddyfile`)
		runShellSilent("systemctl reload caddy 2>/dev/null || true")
	}
	os.Remove("/usr/local/bin/bot")

	if confirmPrompt("Удалить каталог "+installDir+"?", false) {
		os.RemoveAll(installDir)
	}

	printSuccessBox(successStyle.Render("Удаление завершено!"))
}

// ════════════════════════════════════════════════════════════════
// INSTALL WIZARD
// ════════════════════════════════════════════════════════════════

func installWizard() {
	printBanner()
	checkRoot()

	printBox("📋 Перед началом",
		infoStyle.Render("Убедитесь, что у вас есть:")+"\n\n"+
			highlightStyle.Render("  1. ")+"BOT_TOKEN от @BotFather\n"+
			highlightStyle.Render("  2. ")+"Ваш Telegram ID (от @userinfobot)\n"+
			highlightStyle.Render("  3. ")+"REMNAWAVE_API_KEY из настроек панели\n"+
			highlightStyle.Render("  4. ")+"DNS-записи для доменов (опционально)")

	if !confirmPrompt("Начать установку?", true) {
		os.Exit(0)
	}

	cfg := &Config{}

	// 1. System
	globalProgress.advance("Проверка системы")
	detectOS()

	// 2. Packages
	globalProgress.advance("Установка пакетов")
	updateSystem()
	installBasePackages()

	// 3. Docker
	globalProgress.advance("Настройка Docker")
	installDocker()

	// 4. Install dir
	globalProgress.advance("Каталог установки")
	selectInstallDir(cfg)

	// 5. Panel config
	globalProgress.advance("Конфигурация панели")
	checkRemnawavePanel(cfg)

	// 6. Check existing data
	globalProgress.advance("Проверка данных")
	checkPostgresVolume(cfg)

	// 7. Clone
	globalProgress.advance("Клонирование репозитория")
	cloneRepository(cfg)
	createDirectories(cfg)

	// 8. Interactive setup
	globalProgress.advance("Интерактивная настройка")
	interactiveSetup(cfg)

	// 9. Env file
	globalProgress.advance("Файл окружения")
	createEnvFile(cfg)

	// 10. Reverse proxy
	globalProgress.advance("Обратный прокси")
	switch cfg.ReverseProxyType {
	case "nginx_system":
		setupNginxSystem(cfg)
	case "nginx_panel":
		setupNginxPanel(cfg)
	case "caddy":
		setupCaddy(cfg)
	}
	setupSSL(cfg)

	// 11. Docker start
	globalProgress.advance("Docker-контейнеры")
	startDocker(cfg)
	setupFirewall()

	// 12. Finish
	globalProgress.advance("Завершение")
	createManagementScript(cfg)
	printFinalInfo(cfg)

	if confirmPrompt("Показать логи бота?", false) {
		composeFile := "docker-compose.yml"
		if cfg.PanelInstalledLocally {
			composeFile = "docker-compose.local.yml"
		}
		allowExit = true
		runShell(fmt.Sprintf("cd %s && docker compose -f %s logs --tail=150 -f bot", cfg.InstallDir, composeFile))
	}
}

// ════════════════════════════════════════════════════════════════
// MAIN
// ════════════════════════════════════════════════════════════════

func main() {
	// Переходим в существующую директорию чтобы избежать ошибок getcwd
	os.Chdir("/root")
	
	setupSignalHandler()

	if len(os.Args) < 2 {
		installWizard()
		return
	}

	switch os.Args[1] {
	case "install":
		installWizard()
	case "update", "upgrade":
		updateBot()
	case "uninstall", "remove":
		uninstallBot()
	case "version", "--version", "-v":
		fmt.Println(lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("bedolaga_installer") + " " + dimStyle.Render("v"+appVersion))
	case "help", "--help", "-h":
		printBanner()
		fmt.Println(highlightStyle.Render("  Команды:"))
		fmt.Println(dimStyle.Render("    install    ") + "Запустить мастер установки")
		fmt.Println(dimStyle.Render("    update     ") + "Обновить бота (git pull + пересборка)")
		fmt.Println(dimStyle.Render("    uninstall  ") + "Удалить бота")
		fmt.Println(dimStyle.Render("    version    ") + "Показать версию")
		fmt.Println(dimStyle.Render("    help       ") + "Показать эту справку")
		fmt.Println()
	default:
		printError("Неизвестная команда: " + os.Args[1])
		fmt.Println(dimStyle.Render("  Используйте: bedolaga_installer help"))
		os.Exit(1)
	}
}
