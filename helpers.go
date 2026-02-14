package main

import "fmt"

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
