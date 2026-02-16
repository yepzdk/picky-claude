// Package context builds startup context injections from observations and
// summaries, respecting a token budget.
package context

// EstimateTokens returns a rough token count for a string.
// Uses the ~4 characters per token heuristic common for English text.
func EstimateTokens(s string) int {
	if len(s) == 0 {
		return 0
	}
	return (len(s) + 3) / 4
}
