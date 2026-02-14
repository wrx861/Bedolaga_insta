package ui

import "fmt"

// ════════════════════════════════════════════════════════════════
// PRINT HELPERS
// ════════════════════════════════════════════════════════════════

func PrintStep(msg string)    { fmt.Println("\n" + StepStyle.Render("  ▸ "+msg)) }
func PrintInfo(msg string)    { fmt.Println(InfoStyle.Render("  ℹ " + msg)) }
func PrintSuccess(msg string) { fmt.Println(SuccessStyle.Render("  ✓ " + msg)) }
func PrintError(msg string)   { fmt.Println(ErrorStyle.Render("  ✗ " + msg)) }
func PrintWarning(msg string) { fmt.Println(WarnStyle.Render("  ⚠ " + msg)) }
func PrintDim(msg string)     { fmt.Println(DimStyle.Render("    " + msg)) }

func PrintLiveInfo(msg string)    { fmt.Printf("\r\033[K%s\n", InfoStyle.Render("  ℹ "+msg)) }
func PrintLiveSuccess(msg string) { fmt.Printf("\r\033[K%s\n", SuccessStyle.Render("  ✓ "+msg)) }
func PrintLiveWarning(msg string) { fmt.Printf("\r\033[K%s\n", WarnStyle.Render("  ⚠ "+msg)) }
func PrintLiveError(msg string)   { fmt.Printf("\r\033[K%s\n", ErrorStyle.Render("  ✗ "+msg)) }

func PrintBox(title, content string) {
	inner := SubtitleStyle.Render(title) + "\n" + content
	fmt.Println(BoxStyle.Render(inner))
}

func PrintSuccessBox(content string) {
	fmt.Println(SuccessBoxStyle.Render(content))
}

func PrintErrorBox(content string) {
	fmt.Println(ErrorBoxStyle.Render(content))
}
