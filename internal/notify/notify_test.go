package notify

import (
	"testing"
)

func TestBuildOsascriptArgs(t *testing.T) {
	args := buildOsascriptArgs("Test Title", "Test body message")
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "-e" {
		t.Errorf("args[0] = %q, want %q", args[0], "-e")
	}
	// Should contain the title and message in the AppleScript
	script := args[1]
	if script == "" {
		t.Error("script is empty")
	}
}

func TestBuildNotifySendArgs(t *testing.T) {
	args := buildNotifySendArgs("Test Title", "Test body message")
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "Test Title" {
		t.Errorf("args[0] = %q, want %q", args[0], "Test Title")
	}
	if args[1] != "Test body message" {
		t.Errorf("args[1] = %q, want %q", args[1], "Test body message")
	}
}

func TestNotificationMessage(t *testing.T) {
	tests := []struct {
		event NotifyEvent
		title string
	}{
		{EventContextWarning, "Context Warning"},
		{EventSessionComplete, "Session Complete"},
		{EventVerificationDone, "Verification Done"},
		{EventPlanApproved, "Plan Approved"},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			title, _ := eventStrings(tt.event, "details")
			if title != tt.title {
				t.Errorf("eventStrings(%q) title = %q, want %q", tt.event, title, tt.title)
			}
		})
	}
}

func TestEventStringsUnknown(t *testing.T) {
	title, body := eventStrings("unknown-event", "msg")
	if title != "Picky Claude" {
		t.Errorf("title = %q, want %q", title, "Picky Claude")
	}
	if body != "msg" {
		t.Errorf("body = %q, want %q", body, "msg")
	}
}
