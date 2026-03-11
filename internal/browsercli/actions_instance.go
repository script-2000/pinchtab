package browsercli

import (
	"fmt"
	"github.com/pinchtab/pinchtab/internal/cliui"
	"net/http"
)

func Instance(client *http.Client, base, token string, args []string) {
	if len(args) < 1 {
		cliui.Fatal("Usage: pinchtab instance <subcommand> [options]\nSubcommands: start, launch (alias), navigate, logs, stop")
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "start", "launch": // "start" is new Phase 2 API, "launch" is legacy
		InstanceStart(client, base, token, subArgs)
	case "navigate":
		InstanceNavigate(client, base, token, subArgs)
	case "logs":
		InstanceLogs(client, base, token, subArgs)
	case "stop":
		InstanceStop(client, base, token, subArgs)
	default:
		cliui.Fatal("Unknown subcommand: %s", subCmd)
	}
}

func InstanceStart(client *http.Client, base, token string, args []string) {
	body := map[string]any{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--profileId":
			if i+1 < len(args) {
				body["profileId"] = args[i+1]
				i++
			}
		case "--mode":
			if i+1 < len(args) {
				body["mode"] = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				body["port"] = args[i+1]
				i++
			}
		}
	}

	// Use new /instances/start endpoint if available, fall back to /instances/launch for backward compat
	endpoint := "/instances/start"
	DoPost(client, base, token, endpoint, body)
}

func InstanceNavigate(client *http.Client, base, token string, args []string) {
	if len(args) < 2 {
		cliui.Fatal("Usage: pinchtab instance navigate <instance-id> <url>")
	}

	instID := args[0]
	targetURL := args[1]

	// Instance navigate now works via tab-scoped navigation:
	// open a tab on the instance, then navigate that tab.
	openResp := DoPost(client, base, token, fmt.Sprintf("/instances/%s/tabs/open", instID), map[string]any{
		"url": "about:blank",
	})
	tabID, _ := openResp["tabId"].(string)
	if tabID == "" {
		cliui.Fatal("failed to open tab for instance %s", instID)
	}

	// DoPost auto-prints JSON response.
	DoPost(client, base, token, fmt.Sprintf("/tabs/%s/navigate", tabID), map[string]any{
		"url": targetURL,
	})
}

func InstanceLogs(client *http.Client, base, token string, args []string) {
	var instID string

	if len(args) == 0 {
		cliui.Fatal("Usage: pinchtab instance logs <instance-id> OR pinchtab instance logs --id <instance-id>")
	}

	if args[0] == "--id" {
		if len(args) < 2 {
			cliui.Fatal("Usage: --id requires instance ID")
		}
		instID = args[1]
	} else {
		instID = args[0]
	}

	logs := DoGetRaw(client, base, token, fmt.Sprintf("/instances/%s/logs", instID), nil)
	fmt.Println(string(logs))
}

func InstanceStop(client *http.Client, base, token string, args []string) {
	var instID string

	if len(args) == 0 {
		cliui.Fatal("Usage: pinchtab instance stop <instance-id> OR pinchtab instance stop --id <instance-id>")
	}

	if args[0] == "--id" {
		if len(args) < 2 {
			cliui.Fatal("Usage: --id requires instance ID")
		}
		instID = args[1]
	} else {
		instID = args[0]
	}

	DoPost(client, base, token, fmt.Sprintf("/instances/%s/stop", instID), nil)
}
