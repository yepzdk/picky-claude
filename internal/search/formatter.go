package search

import "fmt"

// FormatResult returns a human-readable string for a search result.
func FormatResult(r *HybridResult) string {
	return fmt.Sprintf("[%d] (%.2f) [%s] %s", r.ID, r.Score, r.ObsType, r.Title)
}

// FormatResults returns a formatted string for multiple results.
func FormatResults(results []HybridResult) string {
	if len(results) == 0 {
		return "No results found."
	}
	var s string
	for i := range results {
		s += FormatResult(&results[i]) + "\n"
	}
	return s
}
