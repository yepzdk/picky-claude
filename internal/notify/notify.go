// Package notify provides cross-platform desktop notifications.
// macOS uses osascript; Linux uses notify-send.
package notify

import (
	"fmt"
	"os/exec"
	"runtime"
)

// NotifyEvent identifies the type of notification.
type NotifyEvent string

const (
	EventContextWarning  NotifyEvent = "context-warning"
	EventSessionComplete NotifyEvent = "session-complete"
	EventVerificationDone NotifyEvent = "verification-done"
	EventPlanApproved    NotifyEvent = "plan-approved"
)

// Send sends a desktop notification for the given event with a detail message.
// Returns nil if the notification was sent (or notifications are unsupported).
func Send(event NotifyEvent, detail string) error {
	title, body := eventStrings(event, detail)

	switch runtime.GOOS {
	case "darwin":
		return runOsascript(title, body)
	case "linux":
		return runNotifySend(title, body)
	default:
		// Silently ignore unsupported platforms.
		return nil
	}
}

// SendRaw sends a notification with a custom title and body.
func SendRaw(title, body string) error {
	switch runtime.GOOS {
	case "darwin":
		return runOsascript(title, body)
	case "linux":
		return runNotifySend(title, body)
	default:
		return nil
	}
}

func eventStrings(event NotifyEvent, detail string) (string, string) {
	switch event {
	case EventContextWarning:
		return "Context Warning", detail
	case EventSessionComplete:
		return "Session Complete", detail
	case EventVerificationDone:
		return "Verification Done", detail
	case EventPlanApproved:
		return "Plan Approved", detail
	default:
		return "Picky Claude", detail
	}
}

func runOsascript(title, body string) error {
	args := buildOsascriptArgs(title, body)
	return exec.Command("osascript", args...).Run()
}

func buildOsascriptArgs(title, body string) []string {
	script := fmt.Sprintf(
		`display notification %q with title %q`,
		body, title,
	)
	return []string{"-e", script}
}

func runNotifySend(title, body string) error {
	args := buildNotifySendArgs(title, body)
	return exec.Command("notify-send", args...).Run()
}

func buildNotifySendArgs(title, body string) []string {
	return []string{title, body}
}
