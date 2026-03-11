package main

import (
	"github.com/pinchtab/pinchtab/internal/cliui"
	"testing"
)

func TestSecurityLevelColor(t *testing.T) {
	tests := []struct {
		level string
		want  string
	}{
		{level: "LOCKED", want: string(cliui.ColorSuccess)},
		{level: "GUARDED", want: string(cliui.ColorWarning)},
		{level: "ELEVATED", want: string(cliui.ColorDanger)},
		{level: "EXPOSED", want: string(cliui.ColorDanger)},
		{level: "UNKNOWN", want: string(cliui.ColorDanger)},
	}

	for _, tt := range tests {
		if got := securityLevelColor(tt.level); got != tt.want {
			t.Fatalf("securityLevelColor(%q) = %q, want %q", tt.level, got, tt.want)
		}
	}
}
