package hooks

import (
	"testing"
)

func TestRegisterAndLookup(t *testing.T) {
	called := false
	Register("test-hook", func(input *Input) error {
		called = true
		return nil
	})

	names := RegisteredHooks()
	found := false
	for _, n := range names {
		if n == "test-hook" {
			found = true
			break
		}
	}
	if !found {
		t.Error("test-hook not found in RegisteredHooks()")
	}

	// We can't easily test Dispatch without mocking stdin, but we can
	// verify the registry lookup works.
	_, ok := registry["test-hook"]
	if !ok {
		t.Error("test-hook not in registry")
	}
	_ = called
}

func TestDispatchUnknownHook(t *testing.T) {
	// Dispatch with stdin will fail, but unknown hook should error first
	// We need to provide something on stdin for ReadInput not to block.
	// Instead, just test the registry lookup part.
	_, ok := registry["nonexistent-hook-xyz"]
	if ok {
		t.Error("nonexistent hook should not be in registry")
	}
}
