package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pinchtab/pinchtab/internal/config"
)

type fakeCommandRunner struct {
	calls   []string
	outputs map[string]string
	errors  map[string]error
}

func (f *fakeCommandRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	call := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, call)
	if out, ok := f.outputs[call]; ok {
		return []byte(out), f.errors[call]
	}
	return nil, f.errors[call]
}

func TestEnsureOnboardConfigCreatesDefaultConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "pinchtab", "config.json")
	t.Setenv("PINCHTAB_CONFIG", configPath)
	t.Setenv("PINCHTAB_TOKEN", "")
	t.Setenv("PINCHTAB_BIND", "")

	gotPath, cfg, status, err := ensureOnboardConfig(false)
	if err != nil {
		t.Fatalf("ensureOnboardConfig returned error: %v", err)
	}
	if status != onboardConfigCreated {
		t.Fatalf("status = %q, want %q", status, onboardConfigCreated)
	}
	if gotPath != configPath {
		t.Fatalf("config path = %q, want %q", gotPath, configPath)
	}
	if cfg.Bind != "127.0.0.1" {
		t.Fatalf("bind = %q, want 127.0.0.1", cfg.Bind)
	}
	if strings.TrimSpace(cfg.Token) == "" {
		t.Fatal("expected generated token to be set")
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `"bind": "127.0.0.1"`) {
		t.Fatalf("expected config file to include bind, got %s", content)
	}
	if !strings.Contains(content, `"token": "`) {
		t.Fatalf("expected config file to include token, got %s", content)
	}
}

func TestEnsureOnboardConfigRecoversExistingSecuritySettings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "pinchtab", "config.json")
	t.Setenv("PINCHTAB_CONFIG", configPath)
	input := `{
  "server": {
    "bind": "0.0.0.0",
    "port": "9999",
    "token": ""
  },
  "browser": {
    "binary": "/custom/chrome"
  },
  "security": {
    "allowEvaluate": true
  }
}
`
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(input), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	_, cfg, status, err := ensureOnboardConfig(false)
	if err != nil {
		t.Fatalf("ensureOnboardConfig returned error: %v", err)
	}
	if status != onboardConfigRecovered {
		t.Fatalf("status = %q, want %q", status, onboardConfigRecovered)
	}
	if cfg.Bind != "127.0.0.1" {
		t.Fatalf("bind = %q, want 127.0.0.1", cfg.Bind)
	}
	if cfg.Port != "9999" {
		t.Fatalf("port = %q, want 9999", cfg.Port)
	}
	if cfg.ChromeBinary != "/custom/chrome" {
		t.Fatalf("chrome binary = %q, want /custom/chrome", cfg.ChromeBinary)
	}
	if cfg.AllowEvaluate {
		t.Fatal("expected allowEvaluate to be reset to false")
	}
	if strings.TrimSpace(cfg.Token) == "" {
		t.Fatal("expected recovery to generate a token")
	}
}

func TestRenderOnboardGuideIncludesSecurityValues(t *testing.T) {
	cfg := testRuntimeConfig()
	output := renderOnboardGuide("/tmp/pinchtab/config.json", cfg, onboardConfigCreated, true)

	required := []string{
		"Step 2/7  API access",
		"server.bind  127.0.0.1",
		"security.allowEvaluate  false",
		"security.attach.enabled  false",
		"security.idpi.strictMode  true",
		"pinchtab daemon",
	}
	for _, needle := range required {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected onboarding guide to contain %q\n%s", needle, output)
		}
	}
}

func TestSystemdUserManagerInstallWritesUnitAndEnablesService(t *testing.T) {
	root := t.TempDir()
	runner := &fakeCommandRunner{}
	manager := &systemdUserManager{
		env: daemonEnvironment{
			homeDir:       root,
			osName:        "linux",
			execPath:      "/usr/local/bin/pinchtab",
			xdgConfigHome: filepath.Join(root, ".config"),
		},
		runner: runner,
	}

	message, err := manager.Install("/usr/local/bin/pinchtab", "/tmp/pinchtab/config.json")
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if !strings.Contains(message, manager.ServicePath()) {
		t.Fatalf("install message = %q, want path %q", message, manager.ServicePath())
	}

	data, err := os.ReadFile(manager.ServicePath())
	if err != nil {
		t.Fatalf("reading service file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `ExecStart="/usr/local/bin/pinchtab" server`) {
		t.Fatalf("unexpected unit content: %s", content)
	}
	if !strings.Contains(content, `Environment="PINCHTAB_CONFIG=/tmp/pinchtab/config.json"`) {
		t.Fatalf("expected config env in unit content: %s", content)
	}

	expectedCalls := []string{
		"systemctl --user daemon-reload",
		"systemctl --user enable --now pinchtab.service",
	}
	if strings.Join(runner.calls, "\n") != strings.Join(expectedCalls, "\n") {
		t.Fatalf("systemd calls = %v, want %v", runner.calls, expectedCalls)
	}
}

func TestLaunchdManagerInstallWritesPlistAndBootstrapsAgent(t *testing.T) {
	root := t.TempDir()
	runner := &fakeCommandRunner{}
	manager := &launchdManager{
		env: daemonEnvironment{
			homeDir:  root,
			osName:   "darwin",
			execPath: "/Applications/Pinchtab.app/Contents/MacOS/pinchtab",
			userID:   "501",
		},
		runner: runner,
	}

	message, err := manager.Install("/Applications/Pinchtab.app/Contents/MacOS/pinchtab", "/tmp/pinchtab/config.json")
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if !strings.Contains(message, manager.ServicePath()) {
		t.Fatalf("install message = %q, want path %q", message, manager.ServicePath())
	}

	data, err := os.ReadFile(manager.ServicePath())
	if err != nil {
		t.Fatalf("reading launchd plist: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<string>com.pinchtab.pinchtab</string>") {
		t.Fatalf("expected launchd label in plist: %s", content)
	}
	if !strings.Contains(content, "<string>/Applications/Pinchtab.app/Contents/MacOS/pinchtab</string>") {
		t.Fatalf("expected executable path in plist: %s", content)
	}
	if !strings.Contains(content, "<string>/tmp/pinchtab/config.json</string>") {
		t.Fatalf("expected config path in plist: %s", content)
	}

	expectedCalls := []string{
		"launchctl bootout gui/501 " + manager.ServicePath(),
		"launchctl bootstrap gui/501 " + manager.ServicePath(),
		"launchctl kickstart -k gui/501/com.pinchtab.pinchtab",
	}
	if strings.Join(runner.calls, "\n") != strings.Join(expectedCalls, "\n") {
		t.Fatalf("launchctl calls = %v, want %v", runner.calls, expectedCalls)
	}
}

func TestNewDaemonManagerRejectsUnsupportedOS(t *testing.T) {
	_, err := newDaemonManager(daemonEnvironment{osName: "windows"}, &fakeCommandRunner{})
	if err == nil {
		t.Fatal("expected unsupported OS error")
	}
}

func testRuntimeConfig() *config.RuntimeConfig {
	return &config.RuntimeConfig{
		Bind:               "127.0.0.1",
		Token:              "abcd1234efgh5678",
		AllowEvaluate:      false,
		AllowMacro:         false,
		AllowScreencast:    false,
		AllowDownload:      false,
		AllowUpload:        false,
		AttachEnabled:      false,
		AttachAllowHosts:   []string{"127.0.0.1", "localhost", "::1"},
		AttachAllowSchemes: []string{"ws", "wss"},
		IDPI: config.IDPIConfig{
			Enabled:        true,
			AllowedDomains: []string{"127.0.0.1", "localhost", "::1"},
			StrictMode:     true,
			ScanContent:    true,
			WrapContent:    true,
		},
	}
}
