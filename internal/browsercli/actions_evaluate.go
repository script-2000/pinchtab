package browsercli

import (
	"github.com/pinchtab/pinchtab/internal/cliui"
	"net/http"
	"strings"
)

func Evaluate(client *http.Client, base, token string, args []string) {
	if len(args) < 1 {
		cliui.Fatal("Usage: pinchtab eval <expression>")
	}
	expr := strings.Join(args, " ")
	DoPost(client, base, token, "/evaluate", map[string]any{
		"expression": expr,
	})
}
