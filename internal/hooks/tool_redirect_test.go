package hooks

import "testing"

func TestToolRedirectRegistered(t *testing.T) {
	_, ok := registry["tool-redirect"]
	if !ok {
		t.Error("tool-redirect not registered")
	}
}
