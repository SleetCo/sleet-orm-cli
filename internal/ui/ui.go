// Package ui provides colored, formatted terminal output for sleet-cli.
// Colors are automatically disabled when NO_COLOR is set or output is redirected.
package ui

import (
	"fmt"
	"os"
	"strings"
)

// ── ANSI codes ────────────────────────────────────────────────────────────────

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	red    = "\033[31m"
	white  = "\033[97m"
)

// colorEnabled returns true if the terminal supports ANSI colors.
var colorEnabled = func() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// Always-on for Windows Terminal, ConEmu, VS Code terminal, etc.
	for _, env := range []string{"WT_SESSION", "ConEmuANSI", "TERM_PROGRAM", "COLORTERM", "TERM"} {
		if os.Getenv(env) != "" {
			return true
		}
	}
	// Check if stdout is a terminal (works on Unix; on Windows falls back gracefully)
	fi, err := os.Stdout.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
		return true
	}
	return false
}()

func c(color, s string) string {
	if colorEnabled {
		return color + s + reset
	}
	return s
}

// ── Public helpers ─────────────────────────────────────────────────────────────

// Success prints a green ✓ line.
func Success(msg string) {
	fmt.Println(c(green+bold, "  ✓ ") + msg)
}

// Info prints a cyan ℹ line.
func Info(msg string) {
	fmt.Println(c(cyan, "  ℹ ") + msg)
}

// Hint prints a dim yellow hint line.
func Hint(msg string) {
	fmt.Println(c(dim+yellow, "  · ") + c(dim, msg))
}

// Error prints a red ✗ line to stderr.
func Error(msg string) {
	fmt.Fprintln(os.Stderr, c(red+bold, "  ✗ ")+c(red, msg))
}

// Banner prints the Sleet logo / version banner.
func Banner(version string) {
	line := strings.Repeat("─", 44)
	fmt.Println()
	fmt.Println(c(cyan+bold, "  ❄  Sleet CLI") + c(dim, "  "+version))
	fmt.Println(c(dim, "  "+line))
}

// Step prints a dim ▶ action label (used before a long operation).
func Step(msg string) {
	fmt.Println(c(dim, "  ▶ ")+c(white, msg))
}

// Arrow renders a pretty path arrow.
func Arrow(path string) string {
	return c(dim, " → ") + c(cyan, path)
}
