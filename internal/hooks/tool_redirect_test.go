package hooks

import (
	"encoding/json"
	"testing"
)

func TestIsExploreAgent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"explore", `{"subagent_type": "Explore", "prompt": "find files"}`, true},
		{"general", `{"subagent_type": "general-purpose", "prompt": "research"}`, false},
		{"empty", `{}`, false},
		{"invalid", `not json`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExploreAgent(json.RawMessage(tt.input))
			if got != tt.want {
				t.Errorf("isExploreAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolRedirectRegistered(t *testing.T) {
	_, ok := registry["tool-redirect"]
	if !ok {
		t.Error("tool-redirect not registered")
	}
}
