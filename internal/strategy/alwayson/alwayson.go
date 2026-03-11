package alwayson

import (
	"github.com/pinchtab/pinchtab/internal/strategy"
	"github.com/pinchtab/pinchtab/internal/strategy/autorestart"
)

func init() {
	strategy.MustRegister("always-on", func() strategy.Strategy {
		return autorestart.New(autorestart.AutorestartConfig{
			MaxRestarts:  20,
			StrategyName: "always-on",
			StatusPath:   "/always-on/status",
		})
	})
}
