package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ════════════════════════════════════════════════════════════════
// COLOR PALETTE (Premium Dark)
// ════════════════════════════════════════════════════════════════

var (
	ColorPrimary   = lipgloss.Color("#A78BFA") // Violet
	ColorSecondary = lipgloss.Color("#818CF8") // Indigo
	ColorAccent    = lipgloss.Color("#F59E0B") // Amber/Gold
	ColorSuccess   = lipgloss.Color("#34D399") // Emerald
	ColorError     = lipgloss.Color("#F87171") // Red
	ColorWarning   = lipgloss.Color("#FBBF24") // Yellow
	ColorInfo      = lipgloss.Color("#60A5FA") // Blue
	ColorDim       = lipgloss.Color("#6B7280") // Gray
	ColorWhite     = lipgloss.Color("#F9FAFB") // Near-white
	ColorBg        = lipgloss.Color("#1F2937") // Dark bg
	ColorBgAlt     = lipgloss.Color("#111827") // Darker bg
)

// ════════════════════════════════════════════════════════════════
// STYLES
// ════════════════════════════════════════════════════════════════

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StepStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	WarnStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	DimStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1).
			Width(70)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	AccentBar = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	PromptStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)
)

// ════════════════════════════════════════════════════════════════
// UTILS
// ════════════════════════════════════════════════════════════════

// IsInteractive checks if stdin is a terminal
func IsInteractive() bool {
	fileInfo, _ := fmt.Fprintln, fmt.Fprintln // dummy to keep import
	_ = fileInfo
	return isInteractiveFn()
}
