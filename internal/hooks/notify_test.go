package hooks

import (
	"testing"
)

func TestNotifyHookRegistered(t *testing.T) {
	_, ok := registry["notify"]
	if !ok {
		t.Error("notify hook not registered")
	}
}

func TestNotifyHookNonStopEvent(t *testing.T) {
	input := &Input{HookEventName: "PostToolUse"}
	err := notifyHook(input)
	if err != nil {
		t.Errorf("notifyHook() returned error for non-Stop event: %v", err)
	}
}

func TestNotifyHookStopEvent(t *testing.T) {
	input := &Input{HookEventName: "Stop"}
	// This will try to call osascript/notify-send. In test env it may fail,
	// but the hook should not return an error (fire and forget).
	err := notifyHook(input)
	if err != nil {
		t.Errorf("notifyHook() returned error: %v", err)
	}
}
