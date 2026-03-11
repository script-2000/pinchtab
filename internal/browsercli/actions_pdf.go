package browsercli

import (
	"fmt"
	"github.com/pinchtab/pinchtab/internal/cliui"
	"net/http"
	"net/url"
	"os"
	"time"
)

func PDF(client *http.Client, base, token string, args []string) {
	params := url.Values{}
	params.Set("raw", "true")
	outFile := ""
	tabID := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				outFile = args[i]
			}
		case "--landscape":
			params.Set("landscape", "true")
		case "--scale":
			if i+1 < len(args) {
				i++
				params.Set("scale", args[i])
			}
		case "--tab":
			if i+1 < len(args) {
				i++
				tabID = args[i]
			}
		// Paper dimensions
		case "--paper-width":
			if i+1 < len(args) {
				i++
				params.Set("paperWidth", args[i])
			}
		case "--paper-height":
			if i+1 < len(args) {
				i++
				params.Set("paperHeight", args[i])
			}
		// Margins
		case "--margin-top":
			if i+1 < len(args) {
				i++
				params.Set("marginTop", args[i])
			}
		case "--margin-bottom":
			if i+1 < len(args) {
				i++
				params.Set("marginBottom", args[i])
			}
		case "--margin-left":
			if i+1 < len(args) {
				i++
				params.Set("marginLeft", args[i])
			}
		case "--margin-right":
			if i+1 < len(args) {
				i++
				params.Set("marginRight", args[i])
			}
		// Content options
		case "--page-ranges":
			if i+1 < len(args) {
				i++
				params.Set("pageRanges", args[i])
			}
		case "--prefer-css-page-size":
			params.Set("preferCSSPageSize", "true")
		// Header/Footer
		case "--display-header-footer":
			params.Set("displayHeaderFooter", "true")
		case "--header-template":
			if i+1 < len(args) {
				i++
				params.Set("headerTemplate", args[i])
			}
		case "--footer-template":
			if i+1 < len(args) {
				i++
				params.Set("footerTemplate", args[i])
			}
		// Accessibility
		case "--generate-tagged-pdf":
			params.Set("generateTaggedPDF", "true")
		case "--generate-document-outline":
			params.Set("generateDocumentOutline", "true")
		// Output options
		case "--file-output":
			params.Del("raw")
			params.Set("output", "file")
		case "--path":
			if i+1 < len(args) {
				i++
				params.Set("path", args[i])
			}
		case "--raw":
			params.Set("raw", "true")
		}
	}

	if outFile == "" {
		outFile = fmt.Sprintf("page-%s.pdf", time.Now().Format("20060102-150405"))
	}

	var data []byte
	if tabID != "" {
		data = DoGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/pdf", tabID), params)
	} else {
		data = DoGetRaw(client, base, token, "/pdf", params)
	}
	if data == nil {
		return
	}
	if err := os.WriteFile(outFile, data, 0600); err != nil {
		cliui.Fatal("Write failed: %v", err)
	}
	fmt.Println(cliui.StyleStdout(cliui.SuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
}
