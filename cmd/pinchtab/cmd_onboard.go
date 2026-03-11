package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/pinchtab/pinchtab/internal/cliui"
	"github.com/pinchtab/pinchtab/internal/config"
	"github.com/spf13/cobra"
)

const (
	pinchtabDaemonUnitName = "pinchtab.service"
	pinchtabLaunchdLabel   = "com.pinchtab.pinchtab"
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Guided interactive setup (recommended first run)",
	Long:  "Step-by-step setup to configure your profiles, security defaults, and background service.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		force, _ := cmd.Flags().GetBool("force")
		installDaemonOpt, _ := cmd.Flags().GetBool("install-daemon")
		handleOnboardCommand(cfg, force, installDaemonOpt)
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon [action]",
	Short: "Manage the background service",
	Long:  "Start, stop, install, or check the status of the PinchTab background service.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		sub := ""
		if len(args) > 0 {
			sub = args[0]
		}
		handleDaemonCommand(cfg, sub)
	},
}

func init() {
	onboardCmd.Flags().Bool("install-daemon", false, "Install and start a user-level background service")
	onboardCmd.Flags().Bool("force", false, "Replace the existing config with fresh secure defaults")
	rootCmd.AddCommand(onboardCmd)
	rootCmd.AddCommand(daemonCmd)
}

func handleOnboardCommand(_ *config.RuntimeConfig, force bool, installDaemonOpt bool) {
	configPath, effectiveCfg, configStatus, err := ensureOnboardConfig(force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "onboard failed: %v\n", err)
		os.Exit(1)
	}

	if isInteractiveTerminal() {
		if err := runInteractiveOnboard(configPath, effectiveCfg, configStatus, installDaemonOpt); err != nil {
			fmt.Fprintf(os.Stderr, "onboard failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if installDaemonOpt {
		if _, err := installDaemon(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "daemon install failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Print(renderOnboardGuide(configPath, effectiveCfg, configStatus, installDaemonOpt))
}

func handleDaemonCommand(_ *config.RuntimeConfig, subcommand string) {
	if subcommand == "" || subcommand == "help" || subcommand == "--help" || subcommand == "-h" {
		printDaemonStatusSummary()

		if subcommand == "" && isInteractiveTerminal() {
			var picked string
			err := huh.NewSelect[string]().
				Title("Daemon Actions").
				Options(
					huh.NewOption("Start service", "start"),
					huh.NewOption("Stop service", "stop"),
					huh.NewOption("Restart service", "restart"),
					huh.NewOption("Install service", "install"),
					huh.NewOption("Uninstall service", "uninstall"),
					huh.NewOption("Exit", "exit"),
				).
				Value(&picked).
				WithTheme(pinchtabOnboardTheme()).
				Run()

			if err != nil || picked == "exit" || picked == "" {
				os.Exit(0)
			}
			subcommand = picked
		} else {
			daemonUsage()
			if subcommand == "" {
				os.Exit(0)
			}
			return
		}
	}

	manager, err := currentDaemonManager()
	if err != nil {
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, err.Error()))
		os.Exit(1)
	}

	switch subcommand {
	case "install":
		configPath, _, _, err := ensureOnboardConfig(false)
		if err != nil {
			fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, fmt.Sprintf("daemon install failed: %v", err)))
			os.Exit(1)
		}
		message, err := manager.Install(managerEnvironment(manager).execPath, configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, fmt.Sprintf("daemon install failed: %v", err)))
			fmt.Println()
			fmt.Println(manager.ManualInstructions())
			os.Exit(1)
		}
		fmt.Println(cliui.StyleStdout(cliui.SuccessStyle, "  [ok] ") + message)
		printDaemonFollowUp()
	case "start":
		printDaemonManagerResult(manager.Start())
	case "restart":
		printDaemonManagerResult(manager.Restart())
	case "stop":
		printDaemonManagerResult(manager.Stop())
	case "uninstall":
		message, err := manager.Uninstall()
		if err != nil {
			fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, err.Error()))
			fmt.Println()
			fmt.Println(manager.ManualInstructions())
			os.Exit(1)
		}
		fmt.Println(cliui.StyleStdout(cliui.SuccessStyle, "  [ok] ") + message)
	default:
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, fmt.Sprintf("unknown daemon command: %s", subcommand)))
		daemonUsage()
		os.Exit(2)
	}
}

func daemonUsage() {
	fmt.Println(cliui.StyleStdout(cliui.HeadingStyle, "Usage:") + " " + cliui.StyleStdout(cliui.CommandStyle, "pinchtab daemon <install|start|restart|stop|uninstall>"))
	fmt.Println()
	fmt.Println(cliui.StyleStdout(cliui.MutedStyle, "Manage the PinchTab user-level background service."))
	fmt.Println()
}

func printDaemonStatusSummary() {
	manager, err := currentDaemonManager()
	if err != nil {
		fmt.Println(cliui.StyleStdout(cliui.ErrorStyle, "  Error: ") + err.Error())
		return
	}

	installed := IsDaemonInstalled()
	running := IsDaemonRunning()

	fmt.Println(cliui.StyleStdout(cliui.HeadingStyle, "Daemon status:"))

	status := cliui.StyleStdout(cliui.WarningStyle, "not installed")
	if installed {
		status = cliui.StyleStdout(cliui.SuccessStyle, "installed")
	}
	fmt.Printf("  %-12s %s\n", cliui.StyleStdout(cliui.MutedStyle, "Service:"), status)

	state := cliui.StyleStdout(cliui.MutedStyle, "stopped")
	if running {
		state = cliui.StyleStdout(cliui.SuccessStyle, "active (running)")
	}
	fmt.Printf("  %-12s %s\n", cliui.StyleStdout(cliui.MutedStyle, "State:"), state)

	if running {
		pid, _ := manager.Pid()
		if pid != "" {
			fmt.Printf("  %-12s %s\n", cliui.StyleStdout(cliui.MutedStyle, "PID:"), cliui.StyleStdout(cliui.ValueStyle, pid))
		}
	}

	if installed {
		fmt.Printf("  %-12s %s\n", cliui.StyleStdout(cliui.MutedStyle, "Path:"), cliui.StyleStdout(cliui.ValueStyle, manager.ServicePath()))
	}

	if installed {
		logs, err := manager.Logs(5)
		if err == nil && strings.TrimSpace(logs) != "" {
			fmt.Println()
			fmt.Println(cliui.StyleStdout(cliui.HeadingStyle, "Recent logs:"))
			lines := strings.Split(logs, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					fmt.Printf("  %s\n", cliui.StyleStdout(cliui.MutedStyle, line))
				}
			}
		}
	}
	fmt.Println()
}

func printDaemonManagerResult(message string, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, err.Error()))
		os.Exit(1)
	}
	if strings.HasPrefix(message, "Installed") || strings.HasPrefix(message, "Pinchtab daemon") {
		fmt.Println(cliui.StyleStdout(cliui.SuccessStyle, "  [ok] ") + message)
	} else {
		// For status, it might be a block of text
		fmt.Println(message)
	}
}

func ensureOnboardConfig(force bool) (string, *config.RuntimeConfig, onboardConfigStatus, error) {
	_, configPath, err := config.LoadFileConfig()
	if err != nil {
		return "", nil, "", err
	}

	exists := fileExists(configPath)
	status := onboardConfigVerified
	if !exists || force {
		defaults := config.DefaultFileConfig()
		token, err := config.GenerateAuthToken()
		if err != nil {
			return "", nil, "", err
		}
		defaults.Server.Token = token
		if err := config.SaveFileConfig(&defaults, configPath); err != nil {
			return "", nil, "", err
		}
		if !exists {
			status = onboardConfigCreated
		} else {
			status = onboardConfigRecovered
		}
	}

	return configPath, config.Load(), status, nil
}

type onboardConfigStatus string

const (
	onboardConfigCreated   onboardConfigStatus = "created"
	onboardConfigRecovered onboardConfigStatus = "recovered"
	onboardConfigVerified  onboardConfigStatus = "verified"
)

func runInteractiveOnboard(configPath string, cfg *config.RuntimeConfig, configStatus onboardConfigStatus, installDaemonOpt bool) error {
	totalSteps := 6
	if installDaemonOpt {
		totalSteps = 7
	}

	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println(renderOnboardTitle())
	fmt.Println()

	// Step 1: Config
	fmt.Print(renderOnboardConfigStep(1, totalSteps, configPath, configStatus))

	// Step 2: API
	fmt.Print(renderOnboardAPIAccessStep(2, totalSteps, cfg))

	// Step 3: Capabilities
	fmt.Print(renderOnboardSensitiveStep(3, totalSteps, cfg))

	// Step 4: Attach
	fmt.Print(renderOnboardAttachStep(4, totalSteps, cfg))

	// Step 5: IDPI
	fmt.Print(renderOnboardIDPIStep(5, totalSteps, cfg))

	// Step 6: Daemon (optional)
	currentStep := 6
	daemonInstalled := false
	if installDaemonOpt {
		fmt.Print(cliui.StyleStdout(cliui.MutedStyle, "  Installing background service..."))
		if _, err := installDaemon(configPath); err != nil {
			fmt.Printf("\r  %s %v\n", cliui.StyleStdout(cliui.ErrorStyle, "!!"), err)
		} else {
			fmt.Print("\r")
			daemonInstalled = true
			fmt.Print(renderOnboardDaemonStep(6, totalSteps, true))
		}
		currentStep++
	}

	fmt.Print(renderOnboardNextStep(currentStep, totalSteps, daemonInstalled))
	fmt.Println()

	return nil
}

func installDaemon(configPath string) (string, error) {
	manager, err := currentDaemonManager()
	if err != nil {
		return "", err
	}
	return manager.Install(managerEnvironment(manager).execPath, configPath)
}

func renderOnboardGuide(configPath string, cfg *config.RuntimeConfig, configStatus onboardConfigStatus, installDaemonOpt bool) string {
	var out bytes.Buffer
	total := 6
	if installDaemonOpt {
		total = 7
	}

	out.WriteString(renderOnboardTitle() + "\n\n")
	out.WriteString(renderOnboardConfigStep(1, total, configPath, configStatus))
	out.WriteString(renderOnboardAPIAccessStep(2, total, cfg))
	out.WriteString(renderOnboardSensitiveStep(3, total, cfg))
	out.WriteString(renderOnboardAttachStep(4, total, cfg))
	out.WriteString(renderOnboardIDPIStep(5, total, cfg))

	daemonInstalled := false
	if installDaemonOpt {
		daemonInstalled = IsDaemonInstalled()
		out.WriteString(renderOnboardDaemonStep(6, total, daemonInstalled))
		out.WriteString(renderOnboardNextStep(7, total, daemonInstalled))
	} else {
		out.WriteString(renderOnboardNextStep(6, total, false))
	}

	return out.String()
}

func summarizeToken(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func formatStringSlice(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	return strings.Join(ss, ", ")
}

func writeSecurityToggle(out *bytes.Buffer, key string, enabled bool, description string) {
	val := "false"
	if enabled {
		val = "true"
	}
	writeOnboardKV(out, key, val)
	writeOnboardNote(out, description)
}

type onboardOptions struct {
	installDaemon bool
	force         bool
}

type commandRunner interface {
	CombinedOutput(name string, arg ...string) ([]byte, error)
}

type osCommandRunner struct{}

func (r osCommandRunner) CombinedOutput(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).CombinedOutput()
}

type daemonEnvironment struct {
	execPath      string
	homeDir       string
	osName        string
	userID        string
	xdgConfigHome string
}

type daemonManager interface {
	Install(execPath, configPath string) (string, error)
	ServicePath() string
	Start() (string, error)
	Restart() (string, error)
	Status() (string, error)
	Stop() (string, error)
	Uninstall() (string, error)
	ManualInstructions() string
	Pid() (string, error)
	Logs(n int) (string, error)
}

type systemdUserManager struct {
	env    daemonEnvironment
	runner commandRunner
}

type launchdManager struct {
	env    daemonEnvironment
	runner commandRunner
}

func IsDaemonInstalled() bool {
	manager, err := currentDaemonManager()
	if err != nil {
		return false
	}
	_, err = os.Stat(manager.ServicePath())
	return err == nil
}

func IsDaemonRunning() bool {
	manager, err := currentDaemonManager()
	if err != nil {
		return false
	}
	status, err := manager.Status()
	if err != nil {
		return false
	}
	// Simple string checks for common "running" indicators
	return strings.Contains(status, "state = running") || // launchd
		strings.Contains(status, "Active: active (running)") // systemd
}

func pinchtabOnboardTheme() *huh.Theme {
	t := huh.ThemeBase()

	textPrimary := cliui.ColorTextPrimary
	textSecondary := lipgloss.Color("#94a3b8")
	textMuted := cliui.ColorTextMuted
	accent := cliui.ColorAccent
	accentLight := cliui.ColorAccentLight
	success := cliui.ColorSuccess
	warning := cliui.ColorWarning
	destructive := cliui.ColorDanger

	t.Form.Base = lipgloss.NewStyle().
		Foreground(textPrimary).
		Padding(0, 0)

	t.Group.Base = lipgloss.NewStyle().Foreground(textPrimary)
	t.Group.Title = lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)
	t.Group.Description = lipgloss.NewStyle().
		Foreground(textMuted)

	t.FieldSeparator = lipgloss.NewStyle().SetString("\n")

	t.Focused.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(cliui.ColorBorder).
		BorderLeft(true).
		PaddingLeft(1)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = lipgloss.NewStyle().Foreground(accent).Bold(true)
	t.Focused.NoteTitle = lipgloss.NewStyle().
		Foreground(accentLight).
		Bold(true)
	t.Focused.Description = lipgloss.NewStyle().Foreground(textPrimary)
	t.Focused.ErrorIndicator = lipgloss.NewStyle().Foreground(destructive)
	t.Focused.ErrorMessage = lipgloss.NewStyle().Foreground(destructive)
	t.Focused.SelectSelector = lipgloss.NewStyle().Foreground(accentLight).SetString("> ")
	t.Focused.Option = lipgloss.NewStyle().Foreground(textPrimary)
	t.Focused.SelectedOption = lipgloss.NewStyle().Foreground(success).Bold(true)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(success).SetString("✓ ")
	t.Focused.UnselectedOption = lipgloss.NewStyle().Foreground(textSecondary)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(textMuted).SetString("• ")
	t.Focused.NextIndicator = lipgloss.NewStyle().Foreground(accentLight).SetString("→")
	t.Focused.PrevIndicator = lipgloss.NewStyle().Foreground(accentLight).SetString("←")
	t.Focused.FocusedButton = lipgloss.NewStyle().
		Foreground(accentLight).
		Bold(true).
		Underline(true)
	t.Focused.BlurredButton = lipgloss.NewStyle().
		Foreground(textSecondary).
		Padding(0, 0)
	t.Focused.Next = lipgloss.NewStyle().
		Foreground(accentLight).
		Bold(true).
		Underline(true)
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(accentLight)
	t.Focused.TextInput.CursorText = lipgloss.NewStyle().Foreground(textPrimary)
	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().Foreground(textMuted)
	t.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(accentLight)
	t.Focused.TextInput.Text = lipgloss.NewStyle().Foreground(textPrimary)

	t.Blurred = t.Focused
	t.Blurred.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.HiddenBorder()).
		PaddingLeft(1)
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.Title = lipgloss.NewStyle().Foreground(textSecondary)
	t.Blurred.NoteTitle = lipgloss.NewStyle().Foreground(textSecondary).Bold(true)
	t.Blurred.Description = lipgloss.NewStyle().Foreground(textSecondary)
	t.Blurred.Option = lipgloss.NewStyle().Foreground(textSecondary)
	t.Blurred.SelectedOption = lipgloss.NewStyle().Foreground(success)
	t.Blurred.UnselectedOption = lipgloss.NewStyle().Foreground(textMuted)
	t.Blurred.NextIndicator = lipgloss.NewStyle().Foreground(warning)
	t.Blurred.PrevIndicator = lipgloss.NewStyle().Foreground(warning)

	return t
}

func renderOnboardConfigStep(step, total int, configPath string, configStatus onboardConfigStatus) string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "%s\n", styleOnboardStep(step, total, "Config"))
	switch configStatus {
	case onboardConfigCreated:
		writeOnboardMessage(&out, true, "Wrote a default config to %s.", configPath)
	case onboardConfigRecovered:
		writeOnboardWarning(&out, "Recovered the secure baseline in %s.", configPath)
		writeOnboardNote(&out, "Existing browser, profile, port, and timeout settings were preserved.")
	default:
		writeOnboardMessage(&out, true, "Verified the existing config at %s.", configPath)
		writeOnboardNote(&out, "Recovery defaults were already in place.")
	}
	writeOnboardNote(&out, "PinchTab will read this file unless `PINCHTAB_CONFIG` overrides it.")
	out.WriteString("\n")
	return out.String()
}

func renderOnboardAPIAccessStep(step, total int, cfg *config.RuntimeConfig) string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "%s\n", styleOnboardStep(step, total, "API access"))
	writeOnboardKV(&out, "server.bind", cfg.Bind)
	writeOnboardNote(&out, "Keeps the API local to this machine when set to 127.0.0.1.")
	writeOnboardKV(&out, "server.token", summarizeToken(cfg.Token))
	writeOnboardNote(&out, "A bearer token gates every API request and CLI call that reaches the server.")
	out.WriteString("\n")
	return out.String()
}

func renderOnboardSensitiveStep(step, total int, cfg *config.RuntimeConfig) string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "%s\n", styleOnboardStep(step, total, "Sensitive capabilities"))
	writeSecurityToggle(&out, "security.allowEvaluate", cfg.AllowEvaluate, "JavaScript execution in the page context")
	writeSecurityToggle(&out, "security.allowMacro", cfg.AllowMacro, "higher-level macro automation routes")
	writeSecurityToggle(&out, "security.allowScreencast", cfg.AllowScreencast, "live page streaming output")
	writeSecurityToggle(&out, "security.allowDownload", cfg.AllowDownload, "server-side download endpoints")
	writeSecurityToggle(&out, "security.allowUpload", cfg.AllowUpload, "uploading local files through browser flows")
	out.WriteString("\n")
	return out.String()
}

func renderOnboardAttachStep(step, total int, cfg *config.RuntimeConfig) string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "%s\n", styleOnboardStep(step, total, "Attach policy"))
	writeOnboardKV(&out, "security.attach.enabled", fmt.Sprintf("%t", cfg.AttachEnabled))
	writeOnboardNote(&out, "Leave this off unless you intentionally attach to an externally managed Chrome.")
	writeOnboardKV(&out, "security.attach.allowHosts", strings.Join(cfg.AttachAllowHosts, ", "))
	writeOnboardNote(&out, "Keeps attach targets scoped to trusted hosts.")
	writeOnboardKV(&out, "security.attach.allowSchemes", strings.Join(cfg.AttachAllowSchemes, ", "))
	writeOnboardNote(&out, "Restricts attach URLs to websocket transport.")
	out.WriteString("\n")
	return out.String()
}

func renderOnboardIDPIStep(step, total int, cfg *config.RuntimeConfig) string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "%s\n", styleOnboardStep(step, total, "IDPI protections"))
	writeOnboardKV(&out, "security.idpi.enabled", fmt.Sprintf("%t", cfg.IDPI.Enabled))
	writeOnboardNote(&out, "Enables indirect prompt-injection defenses for navigation and extracted content.")
	writeOnboardKV(&out, "security.idpi.allowedDomains", formatStringSlice(cfg.IDPI.AllowedDomains))
	writeOnboardNote(&out, "Whitelists which sites are considered trusted enough for the local-default setup.")
	writeOnboardKV(&out, "security.idpi.strictMode", fmt.Sprintf("%t", cfg.IDPI.StrictMode))
	writeOnboardNote(&out, "Blocks disallowed domains or suspicious content instead of only warning.")
	writeOnboardKV(&out, "security.idpi.scanContent", fmt.Sprintf("%t", cfg.IDPI.ScanContent))
	writeOnboardNote(&out, "Scans `/text` and `/snapshot` output for prompt-injection patterns.")
	writeOnboardKV(&out, "security.idpi.wrapContent", fmt.Sprintf("%t", cfg.IDPI.WrapContent))
	writeOnboardNote(&out, "Wraps extracted content so downstream agents can treat it as untrusted data.")
	out.WriteString("\n")
	return out.String()
}

func renderOnboardDaemonStep(step, total int, installed bool) string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "%s\n", styleOnboardStep(step, total, "Background service"))
	if installed {
		writeOnboardMessage(&out, true, "Installed and started a user-level daemon for PinchTab.")
		writeOnboardNote(&out, "Use `pinchtab daemon`, `start`, `restart`, `stop`, or `uninstall` to manage it.")
		out.WriteString("\n")
		return out.String()
	}
	writeOnboardWarning(&out, "This step installs a user-level background service and starts PinchTab automatically.")
	writeOnboardNote(&out, "After install you can use `pinchtab daemon`, `start`, `restart`, `stop`, and `uninstall`.")
	out.WriteString("\n")
	return out.String()
}

func renderOnboardNextStep(step, total int, installDaemon bool) string {
	var out bytes.Buffer
	if step > 0 && total > 0 {
		fmt.Fprintf(&out, "%s\n", styleOnboardStep(step, total, "Next commands"))
	} else {
		fmt.Fprintf(&out, "%s\n", cliui.ValueStyle.Render("Next commands"))
	}
	if installDaemon {
		writeOnboardCommand(&out, "pinchtab daemon")
		writeOnboardCommand(&out, "pinchtab health")
		writeOnboardCommand(&out, "pinchtab security")
	} else {
		writeOnboardCommand(&out, "pinchtab server")
		writeOnboardCommand(&out, "pinchtab health")
		writeOnboardCommand(&out, "pinchtab security")
	}
	writeOnboardCommand(&out, "pinchtab quick https://example.com")
	return out.String()
}

func renderOnboardTitle() string {
	return cliui.HeadingStyle.Render("PinchTab") + "  " + cliui.CommandStyle.Render("onboarding")
}

func styleOnboardStep(step, total int, title string) string {
	prefix := cliui.StyleStdout(cliui.HeadingStyle, fmt.Sprintf("Step %d/%d", step, total))
	return prefix + "  " + cliui.ValueStyle.Render(title)
}

func writeOnboardKV(out *bytes.Buffer, key, value string) {
	fmt.Fprintf(out, "  %s  %s\n", cliui.MutedStyle.Render(key), cliui.ValueStyle.Render(value))
}

func writeOnboardNote(out *bytes.Buffer, message string, args ...any) {
	fmt.Fprintf(out, "  %s %s\n", cliui.StyleStdout(cliui.MutedStyle, ">"), cliui.StyleStdout(cliui.MutedStyle, fmt.Sprintf(message, args...)))
}

func writeOnboardMessage(out *bytes.Buffer, ok bool, message string, args ...any) {
	fmt.Fprintf(out, "  [%s] %s\n", styleMarker(ok), fmt.Sprintf(message, args...))
}

func writeOnboardWarning(out *bytes.Buffer, message string, args ...any) {
	fmt.Fprintf(out, "  [%s] %s\n", styleMarker(false), fmt.Sprintf(message, args...))
}

func writeOnboardCommand(out *bytes.Buffer, command string) {
	fmt.Fprintf(out, "  %s %s\n", cliui.StyleStdout(cliui.MutedStyle, "$"), cliui.ValueStyle.Render(command))
}

func printDaemonFollowUp() {
	fmt.Println()
	fmt.Println(cliui.StyleStdout(cliui.HeadingStyle, "Follow-up commands:"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab daemon"), cliui.StyleStdout(cliui.MutedStyle, "# Check service health and logs"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab daemon restart"), cliui.StyleStdout(cliui.MutedStyle, "# Apply config changes"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab daemon stop"), cliui.StyleStdout(cliui.MutedStyle, "# Stop background service"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab daemon uninstall"), cliui.StyleStdout(cliui.MutedStyle, "# Remove service file"))
}

func currentDaemonManager() (daemonManager, error) {
	env, err := currentDaemonEnvironment()
	if err != nil {
		return nil, err
	}
	return newDaemonManager(env, osCommandRunner{})
}

func currentDaemonEnvironment() (daemonEnvironment, error) {
	execPath, err := os.Executable()
	if err != nil {
		return daemonEnvironment{}, fmt.Errorf("resolve executable path: %w", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return daemonEnvironment{}, fmt.Errorf("resolve home directory: %w", err)
	}
	currentUser, err := user.Current()
	if err != nil {
		return daemonEnvironment{}, fmt.Errorf("resolve current user: %w", err)
	}

	return daemonEnvironment{
		execPath:      execPath,
		homeDir:       homeDir,
		osName:        runtime.GOOS,
		userID:        currentUser.Uid,
		xdgConfigHome: os.Getenv("XDG_CONFIG_HOME"),
	}, nil
}

func newDaemonManager(env daemonEnvironment, runner commandRunner) (daemonManager, error) {
	switch env.osName {
	case "linux":
		return &systemdUserManager{env: env, runner: runner}, nil
	case "darwin":
		return &launchdManager{env: env, runner: runner}, nil
	default:
		return nil, fmt.Errorf("pinchtab daemon is supported on macOS and Linux; current OS is %s", env.osName)
	}
}

func managerEnvironment(manager daemonManager) daemonEnvironment {
	switch m := manager.(type) {
	case *systemdUserManager:
		return m.env
	case *launchdManager:
		return m.env
	default:
		return daemonEnvironment{}
	}
}

func (m *systemdUserManager) ServicePath() string {
	return filepath.Join(systemdUserConfigHome(m.env), "systemd", "user", pinchtabDaemonUnitName)
}

func (m *systemdUserManager) Install(execPath, configPath string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(m.ServicePath()), 0755); err != nil {
		return "", fmt.Errorf("create systemd user directory: %w", err)
	}
	if err := os.WriteFile(m.ServicePath(), []byte(renderSystemdUnit(execPath, configPath)), 0644); err != nil {
		return "", fmt.Errorf("write systemd unit: %w", err)
	}
	if _, err := runCommand(m.runner, "systemctl", "--user", "daemon-reload"); err != nil {
		return "", err
	}
	if _, err := runCommand(m.runner, "systemctl", "--user", "enable", "--now", pinchtabDaemonUnitName); err != nil {
		return "", err
	}
	return fmt.Sprintf("Installed systemd user service at %s", m.ServicePath()), nil
}

func (m *systemdUserManager) Start() (string, error) {
	if _, err := runCommand(m.runner, "systemctl", "--user", "start", pinchtabDaemonUnitName); err != nil {
		return "", err
	}
	return "Pinchtab daemon started.", nil
}

func (m *systemdUserManager) Restart() (string, error) {
	if _, err := runCommand(m.runner, "systemctl", "--user", "restart", pinchtabDaemonUnitName); err != nil {
		return "", err
	}
	return "Pinchtab daemon restarted.", nil
}

func (m *systemdUserManager) Stop() (string, error) {
	if _, err := runCommand(m.runner, "systemctl", "--user", "stop", pinchtabDaemonUnitName); err != nil {
		return "", err
	}
	return "Pinchtab daemon stopped.", nil
}

func (m *systemdUserManager) Status() (string, error) {
	output, err := runCommand(m.runner, "systemctl", "--user", "status", pinchtabDaemonUnitName, "--no-pager")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(output) == "" {
		return "Pinchtab daemon status returned no output.", nil
	}
	return output, nil
}

func (m *systemdUserManager) Uninstall() (string, error) {
	var errs []error
	if _, err := runCommand(m.runner, "systemctl", "--user", "disable", "--now", pinchtabDaemonUnitName); err != nil {
		errs = append(errs, err)
	}
	if err := os.Remove(m.ServicePath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, fmt.Errorf("remove unit file: %w", err))
	}
	if _, err := runCommand(m.runner, "systemctl", "--user", "daemon-reload"); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return "", errors.Join(errs...)
	}
	return "Pinchtab daemon uninstalled.", nil
}

func (m *systemdUserManager) Pid() (string, error) {
	output, err := runCommand(m.runner, "systemctl", "--user", "show", pinchtabDaemonUnitName, "--property", "MainPID")
	if err != nil {
		return "", err
	}
	// Output is typically MainPID=1234
	if parts := strings.Split(output, "="); len(parts) == 2 {
		pid := strings.TrimSpace(parts[1])
		if pid == "0" {
			return "", nil // Not running
		}
		return pid, nil
	}
	return "", nil
}

func (m *systemdUserManager) Logs(n int) (string, error) {
	return runCommand(m.runner, "journalctl", "--user", "-u", pinchtabDaemonUnitName, "-n", fmt.Sprintf("%d", n), "--no-pager")
}

func (m *systemdUserManager) ManualInstructions() string {
	path := m.ServicePath()
	var b strings.Builder
	fmt.Fprintln(&b, cliui.StyleStdout(cliui.HeadingStyle, "Manual instructions (Linux/systemd):"))
	fmt.Fprintln(&b, cliui.StyleStdout(cliui.MutedStyle, "To install manually:"))
	fmt.Fprintf(&b, "  1. Create %s\n", cliui.StyleStdout(cliui.ValueStyle, path))
	fmt.Fprintln(&b, "  2. Run: "+cliui.StyleStdout(cliui.CommandStyle, "systemctl --user daemon-reload"))
	fmt.Fprintln(&b, "  3. Run: "+cliui.StyleStdout(cliui.CommandStyle, "systemctl --user enable --now pinchtab.service"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, cliui.StyleStdout(cliui.MutedStyle, "To uninstall manually:"))
	fmt.Fprintln(&b, "  1. Run: "+cliui.StyleStdout(cliui.CommandStyle, "systemctl --user disable --now pinchtab.service"))
	fmt.Fprintf(&b, "  2. Remove: %s\n", cliui.StyleStdout(cliui.ValueStyle, path))
	fmt.Fprintln(&b, "  3. Run: "+cliui.StyleStdout(cliui.CommandStyle, "systemctl --user daemon-reload"))
	return b.String()
}

func renderSystemdUnit(execPath, configPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Pinchtab Browser Service
After=network.target

[Service]
Type=simple
ExecStart="%s" server
Environment="PINCHTAB_CONFIG=%s"
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`, execPath, configPath)
}

func (m *launchdManager) ServicePath() string {
	return filepath.Join(m.env.homeDir, "Library", "LaunchAgents", pinchtabLaunchdLabel+".plist")
}

func (m *launchdManager) Install(execPath, configPath string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(m.ServicePath()), 0755); err != nil {
		return "", fmt.Errorf("create LaunchAgents directory: %w", err)
	}
	if err := os.WriteFile(m.ServicePath(), []byte(renderLaunchdPlist(execPath, configPath)), 0644); err != nil {
		return "", fmt.Errorf("write launchd plist: %w", err)
	}
	_, _ = runCommand(m.runner, "launchctl", "bootout", launchdDomainTarget(m.env), m.ServicePath())
	if _, err := runCommand(m.runner, "launchctl", "bootstrap", launchdDomainTarget(m.env), m.ServicePath()); err != nil {
		return "", err
	}
	if _, err := runCommand(m.runner, "launchctl", "kickstart", "-k", launchdDomainTarget(m.env)+"/"+pinchtabLaunchdLabel); err != nil {
		return "", err
	}
	return fmt.Sprintf("Installed launchd agent at %s", m.ServicePath()), nil
}

func (m *launchdManager) Start() (string, error) {
	if _, err := runCommand(m.runner, "launchctl", "bootstrap", launchdDomainTarget(m.env), m.ServicePath()); err != nil && !strings.Contains(err.Error(), "already bootstrapped") {
		return "", err
	}
	if _, err := runCommand(m.runner, "launchctl", "kickstart", launchdDomainTarget(m.env)+"/"+pinchtabLaunchdLabel); err != nil {
		return "", err
	}
	return "Pinchtab daemon started.", nil
}

func (m *launchdManager) Restart() (string, error) {
	if _, err := runCommand(m.runner, "launchctl", "kickstart", "-k", launchdDomainTarget(m.env)+"/"+pinchtabLaunchdLabel); err != nil {
		return "", err
	}
	return "Pinchtab daemon restarted.", nil
}

func (m *launchdManager) Stop() (string, error) {
	_, err := runCommand(m.runner, "launchctl", "bootout", launchdDomainTarget(m.env), m.ServicePath())
	if err != nil && !isLaunchdIgnorableError(err) {
		return "", err
	}
	return "Pinchtab daemon stopped.", nil
}

func (m *launchdManager) Status() (string, error) {
	output, err := runCommand(m.runner, "launchctl", "print", launchdDomainTarget(m.env)+"/"+pinchtabLaunchdLabel)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(output) == "" {
		return "Pinchtab daemon status returned no output.", nil
	}
	return output, nil
}

func (m *launchdManager) Uninstall() (string, error) {
	var errs []error
	_, err := runCommand(m.runner, "launchctl", "bootout", launchdDomainTarget(m.env), m.ServicePath())
	if err != nil && !isLaunchdIgnorableError(err) {
		errs = append(errs, err)
	}
	if err := os.Remove(m.ServicePath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, fmt.Errorf("remove launchd plist: %w", err))
	}
	if len(errs) > 0 {
		return "", errors.Join(errs...)
	}
	return "Pinchtab daemon uninstalled.", nil
}

func (m *launchdManager) Pid() (string, error) {
	output, err := runCommand(m.runner, "launchctl", "print", launchdDomainTarget(m.env)+"/"+pinchtabLaunchdLabel)
	if err != nil {
		return "", err
	}
	// Try to find pid = 1234
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "pid = ") {
			return strings.TrimPrefix(trimmed, "pid = "), nil
		}
	}
	return "", nil
}

func (m *launchdManager) Logs(n int) (string, error) {
	// macOS log paths we added to plist
	logPath := "/tmp/pinchtab.err.log"
	if _, err := os.Stat(logPath); err != nil {
		return "No logs found at " + logPath, nil
	}
	return runCommand(m.runner, "tail", "-n", fmt.Sprintf("%d", n), logPath)
}

func (m *launchdManager) ManualInstructions() string {
	path := m.ServicePath()
	target := launchdDomainTarget(m.env)
	var b strings.Builder
	fmt.Fprintln(&b, cliui.StyleStdout(cliui.HeadingStyle, "Manual instructions (macOS/launchd):"))
	fmt.Fprintln(&b, cliui.StyleStdout(cliui.MutedStyle, "To install manually:"))
	fmt.Fprintf(&b, "  1. Create %s\n", cliui.StyleStdout(cliui.ValueStyle, path))
	fmt.Fprintln(&b, "  2. Run: "+cliui.StyleStdout(cliui.CommandStyle, fmt.Sprintf("launchctl bootstrap %s %s", target, path)))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, cliui.StyleStdout(cliui.MutedStyle, "To uninstall manually:"))
	fmt.Fprintln(&b, "  1. Run: "+cliui.StyleStdout(cliui.CommandStyle, fmt.Sprintf("launchctl bootout %s %s", target, path)))
	fmt.Fprintf(&b, "  2. Remove: %s\n", cliui.StyleStdout(cliui.ValueStyle, path))
	return b.String()
}

func isLaunchdIgnorableError(err error) bool {
	if err == nil {
		return true
	}
	msg := err.Error()
	// Exit status 5 often means already booted out or path not found in domain
	// "No such process" (status 3) or "No such file or directory" (status 2) or "Operation not permitted" (sometimes)
	return strings.Contains(msg, "exit status 5") ||
		strings.Contains(msg, "No such process") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "already bootstrapped")
}

func renderLaunchdPlist(execPath, configPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>server</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>EnvironmentVariables</key>
  <dict>
    <key>PINCHTAB_CONFIG</key>
    <string>%s</string>
  </dict>
  <key>StandardOutPath</key>
  <string>/tmp/pinchtab.out.log</string>
  <key>StandardErrorPath</key>
  <string>/tmp/pinchtab.err.log</string>
</dict>
</plist>
`, pinchtabLaunchdLabel, execPath, configPath)
}

func runCommand(runner commandRunner, name string, args ...string) (string, error) {
	output, err := runner.CombinedOutput(name, args...)
	trimmed := strings.TrimSpace(string(output))
	if err == nil {
		return trimmed, nil
	}
	if trimmed == "" {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return "", fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, trimmed)
}

func launchdDomainTarget(env daemonEnvironment) string {
	return "gui/" + env.userID
}

func systemdUserConfigHome(env daemonEnvironment) string {
	if strings.TrimSpace(env.xdgConfigHome) != "" {
		return env.xdgConfigHome
	}
	return filepath.Join(env.homeDir, ".config")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isInteractiveTerminal() bool {
	in, err := os.Stdin.Stat()
	if err != nil || (in.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	out, err := os.Stdout.Stat()
	if err != nil || (out.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	return true
}
