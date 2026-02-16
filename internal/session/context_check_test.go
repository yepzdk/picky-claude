package session

import (
	"testing"
)

func TestCheckContextResult(t *testing.T) {
	tests := []struct {
		name       string
		pct        float64
		wantStatus string
	}{
		{"low", 25.0, "OK"},
		{"mid", 60.0, "OK"},
		{"high", 79.9, "OK"},
		{"threshold", 80.0, "CLEAR_NEEDED"},
		{"critical", 95.0, "CLEAR_NEEDED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckContext(tt.pct)
			if result.Status != tt.wantStatus {
				t.Errorf("CheckContext(%f).Status = %q, want %q", tt.pct, result.Status, tt.wantStatus)
			}
			if result.Percentage != tt.pct {
				t.Errorf("CheckContext(%f).Percentage = %f", tt.pct, result.Percentage)
			}
		})
	}
}

func TestCheckContextFromDir(t *testing.T) {
	dir := t.TempDir()

	// No file → error
	_, err := CheckContextFromDir(dir)
	if err == nil {
		t.Error("expected error for missing context-pct.json")
	}

	// Write and read
	WriteContextPercentage(dir, 47.5)
	result, err := CheckContextFromDir(dir)
	if err != nil {
		t.Fatalf("CheckContextFromDir: %v", err)
	}
	if result.Status != "OK" {
		t.Errorf("Status = %q, want OK", result.Status)
	}
	if result.Percentage != 47.5 {
		t.Errorf("Percentage = %f, want 47.5", result.Percentage)
	}
}
