package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

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
