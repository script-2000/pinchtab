package browsercli

import (
	"fmt"
	"github.com/pinchtab/pinchtab/internal/cliui"
	"io"
	"net/http"
	"os"
	"strings"
)

// CheckServerAndGuide checks if pinchtab server is running and provides guidance
func CheckServerAndGuide(client *http.Client, base, token string) bool {
	req, _ := http.NewRequest("GET", base+"/health", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp") {
			fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, fmt.Sprintf("Pinchtab server is not running on %s", base)))
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.HeadingStyle, "To start the server"))
			fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab"), cliui.StyleStderr(cliui.MutedStyle, "# Run in foreground (recommended for beginners)"))
			fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab &"), cliui.StyleStderr(cliui.MutedStyle, "# Run in background"))
			fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "PINCHTAB_PORT=9868 pinchtab"), cliui.StyleStderr(cliui.MutedStyle, "# Use different port"))
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.HeadingStyle, "Then try your command again"))
			fmt.Fprintf(os.Stderr, "  %s\n", cliui.StyleStderr(cliui.CommandStyle, strings.Join(os.Args, " ")))
			fmt.Fprintln(os.Stderr)
			fmt.Fprintf(os.Stderr, "%s %s\n", cliui.StyleStderr(cliui.MutedStyle, "Learn more:"), cliui.StyleStderr(cliui.CommandStyle, "https://github.com/pinchtab/pinchtab#quick-start"))
			return false
		}
		// Other connection errors
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, fmt.Sprintf("Cannot connect to Pinchtab server: %v", err)))
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, "Authentication required. Set PINCHTAB_TOKEN."))
		return false
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.ErrorStyle, fmt.Sprintf("Server error %d: %s", resp.StatusCode, string(body))))
		return false
	}

	return true
}

// SuggestNextAction provides helpful suggestions based on the current command and state
func SuggestNextAction(cmd string, result map[string]any) {
	switch cmd {
	case "nav", "navigate":
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.HeadingStyle, "Next steps"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab snap"), cliui.StyleStderr(cliui.MutedStyle, "# See page structure"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab screenshot"), cliui.StyleStderr(cliui.MutedStyle, "# Capture visual"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab click <ref>"), cliui.StyleStderr(cliui.MutedStyle, "# Click an element"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab pdf --tab <id> -o output.pdf"), cliui.StyleStderr(cliui.MutedStyle, "# Save tab as PDF"))

	case "snap", "snapshot":
		refs := extractRefs(result)
		if len(refs) > 0 {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.HeadingStyle, fmt.Sprintf("Found %d interactive elements", len(refs))))
			for i, ref := range refs[:min(3, len(refs))] {
				fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, fmt.Sprintf("pinchtab click %s", ref.id)), cliui.StyleStderr(cliui.MutedStyle, "# "+ref.desc))
				if i >= 2 {
					break
				}
			}
			if len(refs) > 3 {
				fmt.Fprintf(os.Stderr, "  %s\n", cliui.StyleStderr(cliui.MutedStyle, fmt.Sprintf("... and %d more", len(refs)-3)))
			}
		}

	case "click", "type", "fill":
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, cliui.StyleStderr(cliui.HeadingStyle, "Action completed"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab snap"), cliui.StyleStderr(cliui.MutedStyle, "# See updated page"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", cliui.StyleStderr(cliui.CommandStyle, "pinchtab screenshot"), cliui.StyleStderr(cliui.MutedStyle, "# Visual confirmation"))
	}
}

type refInfo struct {
	id   string
	desc string
}

func extractRefs(data map[string]any) []refInfo {
	var refs []refInfo

	// Handle different snapshot formats
	if elements, ok := data["elements"].([]any); ok {
		for _, elem := range elements {
			if m, ok := elem.(map[string]any); ok {
				if ref, ok := m["ref"].(string); ok && ref != "" {
					desc := ""
					if role, ok := m["role"].(string); ok {
						desc = role
					}
					if name, ok := m["name"].(string); ok && name != "" {
						desc += ": " + name
					}
					// Only include interactive elements
					if role, ok := m["role"].(string); ok {
						if role == "button" || role == "link" || role == "textbox" ||
							role == "checkbox" || role == "radio" || role == "combobox" {
							refs = append(refs, refInfo{id: ref, desc: desc})
						}
					}
				}
			}
		}
	}

	return refs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
