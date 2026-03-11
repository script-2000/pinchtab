package browsercli

import (
	"fmt"
	"github.com/pinchtab/pinchtab/internal/cliui"
	"net/http"
	"net/url"
	"time"
)

func Quick(client *http.Client, base, token string, args []string) {
	if len(args) < 1 {
		cliui.Fatal("Usage: pinchtab quick <url>")
	}

	fmt.Println(cliui.StyleStdout(cliui.HeadingStyle, fmt.Sprintf("Navigating to %s...", args[0])))

	// Navigate
	navBody := map[string]any{"url": args[0]}
	navResult := DoPost(client, base, token, "/navigate", navBody)

	// Small delay for page to stabilize
	time.Sleep(1 * time.Second)

	fmt.Println()
	fmt.Println(cliui.StyleStdout(cliui.HeadingStyle, "Page structure"))

	// Snapshot with interactive filter
	snapParams := url.Values{}
	snapParams.Set("filter", "interactive")
	snapParams.Set("compact", "true")
	DoGet(client, base, token, "/snapshot", snapParams)

	// Extract info from navigation result
	if title, ok := navResult["title"].(string); ok {
		fmt.Println()
		fmt.Printf("%s %s\n", cliui.StyleStdout(cliui.MutedStyle, "Title:"), cliui.StyleStdout(cliui.ValueStyle, title))
	}
	if urlStr, ok := navResult["url"].(string); ok {
		fmt.Printf("%s %s\n", cliui.StyleStdout(cliui.MutedStyle, "URL:"), cliui.StyleStdout(cliui.ValueStyle, urlStr))
	}

	fmt.Println()
	fmt.Println(cliui.StyleStdout(cliui.HeadingStyle, "Quick actions"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab click <ref>"), cliui.StyleStdout(cliui.MutedStyle, "# Click an element (use refs from above)"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab type <ref> <text>"), cliui.StyleStdout(cliui.MutedStyle, "# Type into input field"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab screenshot"), cliui.StyleStdout(cliui.MutedStyle, "# Take a screenshot"))
	fmt.Printf("  %s %s\n", cliui.StyleStdout(cliui.CommandStyle, "pinchtab pdf --tab <id> -o output.pdf"), cliui.StyleStdout(cliui.MutedStyle, "# Save tab as PDF"))
}
