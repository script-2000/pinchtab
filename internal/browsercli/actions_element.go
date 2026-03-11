package browsercli

import (
	"github.com/pinchtab/pinchtab/internal/cliui"
	"net/http"
	"strconv"
	"strings"
)

func Action(client *http.Client, base, token, kind string, args []string) {
	body := map[string]any{"kind": kind}

	switch kind {
	case "click", "hover", "focus":
		var cssSelector string
		var refArg string
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--css":
				if i+1 < len(args) {
					i++
					cssSelector = args[i]
				}
			case "--wait-nav":
				body["waitNav"] = true
			default:
				if refArg == "" {
					refArg = args[i]
				}
			}
		}
		if cssSelector != "" {
			body["selector"] = cssSelector
		} else if refArg != "" {
			body["ref"] = refArg
		} else {
			cliui.Fatal("Usage: pinchtab %s <ref> [--css <selector>] [--wait-nav]", kind)
		}
	case "type":
		if len(args) < 2 {
			cliui.Fatal("Usage: pinchtab type <ref> <text>")
		}
		body["ref"] = args[0]
		body["text"] = strings.Join(args[1:], " ")
	case "fill":
		if len(args) < 2 {
			cliui.Fatal("Usage: pinchtab fill <ref|selector> <text>")
		}
		if strings.HasPrefix(args[0], "e") {
			body["ref"] = args[0]
		} else {
			body["selector"] = args[0]
		}
		body["text"] = strings.Join(args[1:], " ")
	case "press":
		if len(args) < 1 {
			cliui.Fatal("Usage: pinchtab press <key>  (e.g. Enter, Tab, Escape)")
		}
		body["key"] = args[0]
	case "scroll":
		if len(args) < 1 {
			cliui.Fatal("Usage: pinchtab scroll <ref|pixels|direction>  (e.g. e5, 800, or down)")
		}
		if strings.HasPrefix(args[0], "e") {
			body["ref"] = args[0]
		} else if px, err := strconv.Atoi(args[0]); err == nil {
			body["scrollY"] = px
		} else {
			switch strings.ToLower(args[0]) {
			case "down":
				body["scrollY"] = 800
			case "up":
				body["scrollY"] = -800
			case "right":
				body["scrollX"] = 800
			case "left":
				body["scrollX"] = -800
			default:
				cliui.Fatal("Usage: pinchtab scroll <ref|pixels|direction>  (e.g. e5, 800, or down)")
			}
		}
	case "select":
		if len(args) < 2 {
			cliui.Fatal("Usage: pinchtab select <ref> <value>")
		}
		body["ref"] = args[0]
		body["value"] = args[1]
	}

	DoPost(client, base, token, "/action", body)
}
