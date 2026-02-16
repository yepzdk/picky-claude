package hooks

import "encoding/json"

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
				PermissionDecisionReason: "Do not use built-in plan mode. Use the /spec command instead.",
			},
		})

	case "Task":
		if input.ToolInput != nil && isExploreAgent(input.ToolInput) {
			WriteOutput(&Output{
				HookSpecific: &HookSpecificOuput{
					HookEventName:            "PreToolUse",
					PermissionDecision:       "deny",
					PermissionDecisionReason: "Do not use the Explore agent. Use vexor search for semantic codebase exploration, or Grep/Glob for exact matches.",
				},
			})
			return nil
		}
		ExitOK()

	default:
		ExitOK()
	}

	return nil
}

func isExploreAgent(toolInput []byte) bool {
	var ti struct {
		SubagentType string `json:"subagent_type"`
	}
	if err := json.Unmarshal(toolInput, &ti); err != nil {
		return false
	}
	return ti.SubagentType == "Explore"
}
