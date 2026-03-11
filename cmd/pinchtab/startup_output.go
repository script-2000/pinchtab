package main

import (
	"fmt"
	"github.com/pinchtab/pinchtab/internal/cliui"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pinchtab/pinchtab/internal/config"
)

type startupBannerOptions struct {
	Mode       string
	ListenAddr string
	PublicURL  string
	Strategy   string
	Allocation string
	ProfileDir string
}

func printStartupBanner(cfg *config.RuntimeConfig, opts startupBannerOptions) {
	writeBannerLine(renderStartupLogo(blankIfEmpty(opts.Mode, "server")))
	writeBannerf("  %s  %s\n", styleLabel("listen"), styleValue(blankIfEmpty(opts.ListenAddr, cfg.ListenAddr())))
	if opts.PublicURL != "" {
		writeBannerf("  %s  %s\n", styleLabel("url"), styleValue(opts.PublicURL))
	}
	strat := blankIfEmpty(opts.Strategy, "manual")
	alloc := blankIfEmpty(opts.Allocation, "none")
	writeBannerf("  %s  %s\n", styleLabel("str,plc"), styleValue(fmt.Sprintf("%s,%s", strat, alloc)))

	daemonStatus := cliui.StyleStdout(cliui.WarningStyle, "not installed")
	if IsDaemonInstalled() {
		daemonStatus = cliui.StyleStdout(cliui.SuccessStyle, "ok")
	}
	writeBannerf("  %s  %s\n", styleLabel("daemon"), daemonStatus)

	if opts.ProfileDir != "" {
		writeBannerf("  %s  %s\n", styleLabel("profile"), styleValue(opts.ProfileDir))
	}
	printSecuritySummary(os.Stdout, cfg, "  ")
	writeBannerLine("")
}

func printSecuritySummary(w io.Writer, cfg *config.RuntimeConfig, prefix string) {
	posture := assessSecurityPosture(cfg)
	writeSummaryf(w, "\n%s%s %s  %s\n", prefix, styleHeading("Security posture"), styleSecurityBar(posture.Level, posture.Bar), styleSecurityLevel(posture.Level))

	for _, check := range posture.Checks {
		writeSummaryf(w, "%s  %s %s %s\n", prefix, styleMarker(check.Passed), styleCheckLabel(check.Label), styleCheckDetail(check.Passed, check.Detail))
	}
}

func blankIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func renderStartupLogo(mode string) string {
	return styleLogo(startupLogo) + "  " + styleMode(mode)
}

func writeBannerLine(line string) {
	_, _ = fmt.Fprintln(os.Stdout, line)
}

func writeBannerf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

func writeSummaryf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func styleHeading(text string) string {
	return applyStyle(text, cliui.HeadingStyle)
}

func styleCommand(text string) string {
	return applyStyle(text, cliui.CommandStyle)
}

func styleMuted(text string) string {
	return applyStyle(text, cliui.MutedStyle)
}

func styleSuccess(text string) string {
	return applyStyle(text, cliui.SuccessStyle)
}

func styleWarning(text string) string {
	return applyStyle(text, cliui.WarningStyle)
}

func styleError(text string) string {
	return applyStyle(text, cliui.ErrorStyle)
}

func styleLogo(text string) string {
	return applyStyle(text, cliui.HeadingStyle)
}

func styleMode(text string) string {
	return applyStyle(text, cliui.CommandStyle)
}

func styleLabel(text string) string {
	return applyStyle(fmt.Sprintf("%-8s", text), cliui.MutedStyle)
}

func styleValue(text string) string {
	return applyStyle(text, cliui.ValueStyle)
}

func styleCheckLabel(text string) string {
	return applyStyle(fmt.Sprintf("%-20s", text), cliui.MutedStyle)
}

func styleCheckDetail(passed bool, text string) string {
	if passed {
		return applyStyle(text, cliui.SuccessStyle)
	}
	return applyStyle(text, cliui.WarningStyle)
}

func styleMarker(passed bool) string {
	if passed {
		return applyStyle("ok", cliui.SuccessStyle)
	}
	return applyStyle("!!", cliui.ErrorStyle)
}

func styleSecurityLevel(level string) string {
	return applyStyle(level, lipgloss.NewStyle().Foreground(lipgloss.Color(securityLevelColor(level))).Bold(true))
}

func styleSecurityBar(level, bar string) string {
	return applyStyle(bar, lipgloss.NewStyle().Foreground(lipgloss.Color(securityLevelColor(level))).Bold(true))
}

func securityLevelColor(level string) string {
	switch level {
	case "LOCKED":
		return string(cliui.ColorSuccess)
	case "GUARDED":
		return string(cliui.ColorWarning)
	case "ELEVATED":
		return string(cliui.ColorDanger)
	default:
		return string(cliui.ColorDanger)
	}
}

func applyStyle(text string, style lipgloss.Style) string {
	return cliui.RenderStdout(style, text)
}

const startupLogo = `   ____  _            _     _____     _
  |  _ \(_)_ __   ___| |__ |_   _|_ _| |__
  | |_) | | '_ \ / __| '_ \  | |/ _  | '_ \
  |  __/| | | | | (__| | | | | | (_| | |_) |
  |_|   |_|_| |_|\___|_| |_| |_|\__,_|_.__/`
