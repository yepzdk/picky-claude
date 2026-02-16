package hooks

import (
	"github.com/jesperpedersen/picky-claude/internal/notify"
)

func init() {
	Register("notify", notifyHook)
}

// notifyHook sends a desktop notification based on the hook event.
// It is non-blocking and never prevents tool execution.
func notifyHook(input *Input) error {
	var event notify.NotifyEvent
	var detail string

	switch input.HookEventName {
	case "Stop":
		event = notify.EventSessionComplete
		detail = "Claude Code session ended"
	default:
		// For other events, no notification unless triggered with a system message.
		return nil
	}

	// Fire and forget — notification failures should not block.
	notify.Send(event, detail)
	return nil
}
