package config

import (
	"testing"
	"time"
)

func TestLoadLSPDebounce(t *testing.T) {
	tests := map[string]struct {
		env  string
		want time.Duration
	}{
		"default":  {"", DefaultLSPDebounce},
		"duration": {"25ms", 25 * time.Millisecond},
		"zero":     {"0", 0},
		"invalid":  {"soon", DefaultLSPDebounce},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(EnvLSPDebounce, tt.env)

			if got := Load().LSPDebounce; got != tt.want {
				t.Fatalf("LSPDebounce = %v, want %v", got, tt.want)
			}
		})
	}
}
