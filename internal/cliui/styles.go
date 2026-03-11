package cliui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBorder      = lipgloss.Color("#2b3345")
	ColorTextPrimary = lipgloss.Color("#e2e8f0")
	ColorTextMuted   = lipgloss.Color("#64748b")
	ColorAccent      = lipgloss.Color("#60a5fa")
	ColorAccentLight = lipgloss.Color("#93c5fd")
	ColorSuccess     = lipgloss.Color("#22c55e")
	ColorWarning     = lipgloss.Color("#fbbf24")
	ColorDanger      = lipgloss.Color("#ef4444")
)

var (
	HeadingStyle = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	CommandStyle = lipgloss.NewStyle().Foreground(ColorAccentLight)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorTextMuted)
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
	ValueStyle   = lipgloss.NewStyle().Foreground(ColorTextPrimary)
)

func RenderStdout(style lipgloss.Style, text string) string {
	return StyleStdout(style, text)
}

func RenderStderr(style lipgloss.Style, text string) string {
	return StyleStderr(style, text)
}

func StyleStdout(style lipgloss.Style, text string) string {
	return renderToWriter(os.Stdout, style, text)
}

func StyleStderr(style lipgloss.Style, text string) string {
	return renderToWriter(os.Stderr, style, text)
}

func Fatal(format string, args ...any) {
	fmt.Fprint(os.Stderr, StyleStderr(ErrorStyle, fmt.Sprintf(format, args...))+"\n")
	os.Exit(1)
}

func renderToWriter(w *os.File, style lipgloss.Style, text string) string {
	if !shouldColorizeFile(w) {
		return text
	}
	return style.Render(text)
}

func shouldColorizeFile(file *os.File) bool {
	if file == nil {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if force := os.Getenv("CLICOLOR_FORCE"); force != "" && force != "0" {
		return true
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
