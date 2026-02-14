package ui

import "fmt"

// ════════════════════════════════════════════════════════════════
// BANNER
// ════════════════════════════════════════════════════════════════

func PrintBanner(version string) {
	banner := `
    ____  __________  ____  __    ___   _________ 
   / __ )/ ____/ __ \/ __ \/ /   /   | / ____/   |
  / __  / __/ / / / / / / / /   / /| |/ / __/ /| |
 / /_/ / /___/ /_/ / /_/ / /___/ ___ / /_/ / ___ |
/_____/_____/_____/\____/_____/_/  |_\____/_/  |_|
                                                   `

	bannerStyled := SubtitleStyle.Copy().
		Foreground(ColorPrimary).
		Render(banner)

	tagline := AccentBar.Render("  УСТАНОВЩИК BEDOLAGA BOT")
	ver := DimStyle.Render(fmt.Sprintf("  v%s", version))
	separator := DimStyle.Render("  ─────────────────────────────────────────────")

	fmt.Println(bannerStyled)
	fmt.Println(tagline + "  " + ver)
	fmt.Println(separator)
	fmt.Println()
}
