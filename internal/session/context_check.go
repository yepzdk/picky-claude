package session

// ContextResult is the JSON output for the check-context command.
type ContextResult struct {
	Status     string  `json:"status"`
	Percentage float64 `json:"percentage"`
}

// CheckContext evaluates a context percentage and returns a result.
// Status is "CLEAR_NEEDED" when percentage >= 80, otherwise "OK".
func CheckContext(pct float64) ContextResult {
	status := "OK"
	if pct >= 80 {
		status = "CLEAR_NEEDED"
	}
	return ContextResult{
		Status:     status,
		Percentage: pct,
	}
}

// CheckContextFromDir reads the context percentage from the session directory
// and returns a ContextResult.
func CheckContextFromDir(sessionDir string) (ContextResult, error) {
	pct, err := ReadContextPercentage(sessionDir)
	if err != nil {
		return ContextResult{}, err
	}
	return CheckContext(pct), nil
}
