package statusline

// selectTip returns a contextual tip based on the current state.
// Returns empty string if no tip is relevant.
func selectTip(input *Input) string {
	if input.ContextPct >= 90 {
		return "TIP: Handoff imminent"
	}
	if input.ContextPct >= 80 {
		return "TIP: Wrap up current task"
	}
	if input.Plan != nil && input.Plan.Status == "VERIFIED" {
		return "TIP: Plan verified, done!"
	}
	if input.Messages == 0 {
		return "TIP: Session started"
	}
	return ""
}
