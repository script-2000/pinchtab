package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pinchtab/pinchtab/internal/browsercli"
	"github.com/pinchtab/pinchtab/internal/config"
	"github.com/spf13/cobra"
)

var quickCmd = &cobra.Command{
	Use:   "quick <url>",
	Short: "Navigate + analyze page (beginner-friendly)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Quick(client, base, token, args)
		})
	},
}

var navCmd = &cobra.Command{
	Use:   "nav <url>",
	Short: "Navigate to URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Navigate(client, base, token, args)
		})
	},
}

var snapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Snapshot accessibility tree",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Snapshot(client, base, token, args)
		})
	},
}

var clickCmd = &cobra.Command{
	Use:   "click <ref>",
	Short: "Click element",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Action(client, base, token, "click", args)
		})
	},
}

var typeCmd = &cobra.Command{
	Use:   "type <ref> <text>",
	Short: "Type into element",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Action(client, base, token, "type", args)
		})
	},
}

var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Take a screenshot",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Screenshot(client, base, token, args)
		})
	},
}

var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "List or manage tabs",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Tabs(client, base, token, args)
		})
	},
}

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List or manage instances",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Instances(client, base, token)
		})
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check server health",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Health(client, base, token)
		})
	},
}

var pressCmd = &cobra.Command{
	Use:   "press <key>",
	Short: "Press key (Enter, Tab, Escape...)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Action(client, base, token, "press", args)
		})
	},
}

var fillCmd = &cobra.Command{
	Use:   "fill <ref|selector> <text>",
	Short: "Fill input directly",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Action(client, base, token, "fill", args)
		})
	},
}

var hoverCmd = &cobra.Command{
	Use:   "hover <ref>",
	Short: "Hover element",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Action(client, base, token, "hover", args)
		})
	},
}

var scrollCmd = &cobra.Command{
	Use:   "scroll <ref|pixels>",
	Short: "Scroll to element or by pixels",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Action(client, base, token, "scroll", args)
		})
	},
}

var evalCmd = &cobra.Command{
	Use:   "eval <expression>",
	Short: "Evaluate JavaScript",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Evaluate(client, base, token, args)
		})
	},
}

var pdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Export the current page as PDF",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.PDF(client, base, token, args)
		})
	},
}

var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Extract page text",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Text(client, base, token, args)
		})
	},
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List browser profiles",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browsercli.Profiles(client, base, token)
		})
	},
}

var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage browser instances",
}

func init() {
	rootCmd.AddCommand(quickCmd)
	rootCmd.AddCommand(navCmd)
	rootCmd.AddCommand(snapCmd)
	rootCmd.AddCommand(clickCmd)
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(tabsCmd)
	rootCmd.AddCommand(instancesCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(pressCmd)
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(hoverCmd)
	rootCmd.AddCommand(scrollCmd)
	rootCmd.AddCommand(evalCmd)
	rootCmd.AddCommand(pdfCmd)
	rootCmd.AddCommand(textCmd)
	rootCmd.AddCommand(profilesCmd)

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "start <name>",
		Short: "Start a browser instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browsercli.InstanceStart(client, base, token, args)
			})
		},
	})
	instanceCmd.AddCommand(&cobra.Command{
		Use:   "navigate <id> <url>",
		Short: "Navigate an instance to a URL",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browsercli.InstanceNavigate(client, base, token, args)
			})
		},
	})
	instanceCmd.AddCommand(&cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a browser instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browsercli.InstanceStop(client, base, token, args)
			})
		},
	})
	instanceCmd.AddCommand(&cobra.Command{
		Use:   "logs <id>",
		Short: "Get instance logs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browsercli.InstanceLogs(client, base, token, args)
			})
		},
	})
	rootCmd.AddCommand(instanceCmd)
}

func runCLIWith(cfg *config.RuntimeConfig, fn func(client *http.Client, base, token string)) {
	client := &http.Client{Timeout: 60 * time.Second}
	dashPort := cfg.Port
	if dashPort == "" {
		dashPort = "9870"
	}
	base := fmt.Sprintf("http://localhost:%s", dashPort)
	token := cfg.Token

	fn(client, base, token)
}

func printHelp() {
	renderCLIHelp()
}

var cliCommands = map[string]bool{
	"nav": true, "navigate": true,
	"snap": true, "snapshot": true,
	"click": true, "type": true, "press": true, "fill": true,
	"hover": true, "scroll": true, "select": true, "focus": true,
	"text": true, "tabs": true, "tab": true,
	"screenshot": true, "ss": true,
	"eval": true, "evaluate": true,
	"pdf": true, "health": true,
	"help": true, "quick": true,
	"instance": true, "instances": true,
	"profiles": true,
}

func isCLICommand(cmd string) bool {
	return cliCommands[cmd]
}

func runCLI(cfg *config.RuntimeConfig) {
	cmd := os.Args[1]
	rawArgs := os.Args[2:]

	var instanceID string
	args := make([]string, 0, len(rawArgs))
	for i := 0; i < len(rawArgs); i++ {
		if (rawArgs[i] == "--instance" || rawArgs[i] == "-I") && i+1 < len(rawArgs) {
			instanceID = rawArgs[i+1]
			i++
		} else {
			args = append(args, rawArgs[i])
		}
	}

	orchBase := fmt.Sprintf("http://%s:%s", cfg.Bind, cfg.Port)
	if envURL := os.Getenv("PINCHTAB_URL"); envURL != "" {
		orchBase = strings.TrimRight(envURL, "/")
	}

	token := cfg.Token
	if envToken := os.Getenv("PINCHTAB_TOKEN"); envToken != "" {
		token = envToken
	}

	base := orchBase
	if instanceID != "" {
		base = browsercli.ResolveInstanceBase(orchBase, token, instanceID, cfg.Bind)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	if cmd != "help" {
		if !browsercli.CheckServerAndGuide(client, base, token) {
			return
		}
	}

	switch cmd {
	case "nav", "navigate":
		browsercli.Navigate(client, base, token, args)
	case "snap", "snapshot":
		browsercli.Snapshot(client, base, token, args)
	case "click", "type", "press", "fill", "hover", "scroll", "select", "focus":
		browsercli.Action(client, base, token, cmd, args)
	case "text":
		browsercli.Text(client, base, token, args)
	case "tabs", "tab":
		browsercli.Tabs(client, base, token, args)
	case "screenshot", "ss":
		browsercli.Screenshot(client, base, token, args)
	case "eval", "evaluate":
		browsercli.Evaluate(client, base, token, args)
	case "pdf":
		browsercli.PDF(client, base, token, args)
	case "health":
		browsercli.Health(client, base, token)
	case "instance":
		browsercli.Instance(client, base, token, args)
	case "instances":
		browsercli.Instances(client, base, token)
	case "profiles":
		browsercli.Profiles(client, base, token)
	case "quick":
		browsercli.Quick(client, base, token, args)
	case "help":
		cliHelp()
	}
}

func cliHelp() {
	renderCLIHelp()
	os.Exit(0)
}

func renderCLIHelp() {
	fmt.Print(`Pinchtab CLI - browser control from the command line

Usage: pinchtab <command> [args] [flags]

QUICK START:
  pinchtab quick <url>    Navigate and show page structure (combines nav + snap)

WORKFLOW:
  1. Start server:        pinchtab                  (or: pinchtab server)
  2. Navigate:           pinchtab nav https://pinchtab.com
  3. See page:           pinchtab snap             (shows clickable refs)
  4. Interact:           pinchtab click e5         (click element)
  5. Check result:       pinchtab snap             (see changes)

Commands:
  quick <url>             Navigate and analyze page (beginner-friendly)

  INSTANCE MANAGEMENT:
  instance launch         Create new instance (--mode headed, --port 9999)
  instance logs <id>      Get instance logs (for debugging)
  instance stop <id>      Stop instance
  instances               List all running instances

  BROWSER CONTROL:
  nav, navigate <url>     Navigate to URL (--new-tab, --block-images, --block-ads)
  snap, snapshot          Accessibility tree snapshot (-i, -c, -d, --max-tokens N)
  click <ref>             Click element by ref
  type <ref> <text>       Type text into element
  fill <ref> <text>       Set input value (no key events)
  press <key>             Press a key (Enter, Tab, Escape, ...)
  hover <ref>             Hover over element
  scroll <direction>      Scroll page (up, down, left, right)
  select <ref> <value>    Select dropdown option
  focus <ref>             Focus element
  text                    Extract page text (--raw for innerText)
  tabs                    List open tabs
  tabs new <url>          Open new tab
  tabs close <tabId>      Close tab
  ss, screenshot          Take screenshot (-o file, -q quality)
  eval <expression>       Evaluate JavaScript
  pdf                     Export page as PDF (-o file, --landscape, --scale N)

  OTHER:
  health                  Server health check
  help                    Show this help

Environment:
  PINCHTAB_URL            Server URL (default: http://localhost:9867)
  PINCHTAB_TOKEN          Auth token (sent as Bearer)

Flags (global):
  --instance <id>, -I <id>  Target a specific instance (e.g., pinchtab snap --instance abc123)

Pipe with jq:
  pinchtab snap -i | jq '.nodes[] | select(.role=="link")'
`)
}
