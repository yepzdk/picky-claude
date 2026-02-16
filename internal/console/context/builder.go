package context

import (
	"fmt"
	"strings"

	"github.com/jesperpedersen/picky-claude/internal/db"
)

// Builder generates startup context injection strings from observations and
// summaries, staying within a configurable token budget.
type Builder struct {
	maxTokens int
}

// NewBuilder creates a context builder with the given token budget.
func NewBuilder(maxTokens int) *Builder {
	return &Builder{maxTokens: maxTokens}
}

// Build constructs a context injection string from recent observations and
// summaries. It prioritizes summaries (high-level context) then fills
// remaining budget with observations (detailed discoveries).
// Returns empty string if there is nothing to inject.
func (b *Builder) Build(observations []*db.Observation, summaries []*db.Summary) string {
	if len(observations) == 0 && len(summaries) == 0 {
		return ""
	}

	var parts []string
	usedTokens := 0

	// Add summaries first — they provide high-level session context
	if len(summaries) > 0 {
		header := "## Recent Session Summaries\n"
		usedTokens += EstimateTokens(header)
		var summaryLines []string

		for _, s := range summaries {
			line := fmt.Sprintf("- [Session %s] %s", s.SessionID, s.Text)
			lineTokens := EstimateTokens(line)
			if usedTokens+lineTokens > b.maxTokens {
				break
			}
			summaryLines = append(summaryLines, line)
			usedTokens += lineTokens
		}

		if len(summaryLines) > 0 {
			parts = append(parts, header+strings.Join(summaryLines, "\n"))
		}
	}

	// Fill remaining budget with observations
	if len(observations) > 0 {
		header := "## Recent Observations\n"
		headerTokens := EstimateTokens(header)
		if usedTokens+headerTokens < b.maxTokens {
			usedTokens += headerTokens
			var obsLines []string

			for _, o := range observations {
				line := fmt.Sprintf("- [#%d %s] **%s**: %s", o.ID, o.Type, o.Title, o.Text)
				lineTokens := EstimateTokens(line)
				if usedTokens+lineTokens > b.maxTokens {
					break
				}
				obsLines = append(obsLines, line)
				usedTokens += lineTokens
			}

			if len(obsLines) > 0 {
				parts = append(parts, header+strings.Join(obsLines, "\n"))
			}
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n\n")
}
