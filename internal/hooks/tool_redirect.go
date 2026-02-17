package hooks

func init() {
	Register("tool-redirect", toolRedirectHook)
}

// toolRedirectHook runs on PreToolUse to block certain built-in tools and
// redirect Claude to preferred alternatives.
func toolRedirectHook(input *Input) error {
	switch input.ToolName {
	case "WebSearch":
		WriteOutput(&Output{
			HookSpecific: &HookSpecificOuput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       "deny",
				PermissionDecisionReason: "Use MCP web-search tools instead of built-in WebSearch. Try: mcp-cli web-search/search",
			},
		})

	case "WebFetch":
		WriteOutput(&Output{
			HookSpecific: &HookSpecificOuput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       "deny",
				PermissionDecisionReason: "Use MCP web-fetch tools instead of built-in WebFetch. Try: mcp-cli web-fetch/fetch_url",
			},
		})

	case "EnterPlanMode", "ExitPlanMode":
		WriteOutput(&Output{
			HookSpecific: &HookSpecificOuput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       "deny",
				PermissionDecisionReason: "Built-in plan mode is disabled. Do not call EnterPlanMode or ExitPlanMode. Continue working directly without these tools.",
			},
		})

	default:
		ExitOK()
	}

	return nil
}
