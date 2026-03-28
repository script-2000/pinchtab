package bridge

import (
	"context"

	bridgeruntime "github.com/pinchtab/pinchtab/internal/bridge/runtime"
	"github.com/pinchtab/pinchtab/internal/config"
)

const popupGuardInitScript = bridgeruntime.PopupGuardInitScript

// InitChrome initializes a Chrome browser for a Bridge instance.
func InitChrome(cfg *config.RuntimeConfig) (context.Context, context.CancelFunc, context.Context, context.CancelFunc, error) {
	return bridgeruntime.InitChrome(cfg, bridgeruntime.Hooks{
		SetHumanRandSeed:          SetHumanRandSeed,
		IsChromeProfileLockError:  isChromeProfileLockError,
		ClearStaleChromeProfile:   clearStaleChromeProfileLock,
		ConfigureChromeProcessCmd: configureChromeProcess,
	})
}

func defaultChromeFlagArgs() []string {
	return bridgeruntime.DefaultChromeFlagArgs()
}

func buildChromeArgs(cfg *config.RuntimeConfig, port int) []string {
	return bridgeruntime.BuildChromeArgs(cfg, port)
}

// parseChromeFlags converts a space-separated Chrome flags string (e.g.
// "--disable-gpu --flag=value") into chromedp ExecAllocatorOptions.
// Boolean flags (no '=') are passed as true; value flags pass the literal value.
func parseChromeFlags(s string) []chromedp.ExecAllocatorOption {
	var opts []chromedp.ExecAllocatorOption
	for _, field := range strings.Fields(s) {
		f := strings.TrimPrefix(field, "--")
		if i := strings.IndexByte(f, '='); i >= 0 {
			opts = append(opts, chromedp.Flag(f[:i], f[i+1:]))
		} else if f != "" {
			opts = append(opts, chromedp.Flag(f, true))
		}
	}
	return opts
}

// sanitiseProxyURL replaces the password in a proxy URL with "***" for safe logging.
func sanitiseProxyURL(raw string) string {
	// Locate "://" then find the first "@" after it — credentials are between them.
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd < 0 {
		return raw
	}
	after := raw[schemeEnd+3:]
	atIdx := strings.LastIndex(after, "@")
	if atIdx < 0 {
		return raw // no credentials
	}
	credentials := after[:atIdx]
	colonIdx := strings.Index(credentials, ":")
	if colonIdx < 0 {
		return raw // no password to redact
	}
	user := credentials[:colonIdx]
	host := after[atIdx+1:]
	return raw[:schemeEnd+3] + user + ":***@" + host
}
