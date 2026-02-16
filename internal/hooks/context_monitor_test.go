package hooks

import "testing"

func TestCurrentThreshold(t *testing.T) {
	tests := []struct {
		pct  float64
		want int
	}{
		{10, 0},
		{39, 0},
		{40, 40},
		{55, 40},
		{60, 60},
		{79, 60},
		{80, 80},
		{89, 80},
		{90, 90},
		{94, 90},
		{95, 95},
		{100, 95},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := currentThreshold(tt.pct)
			if got != tt.want {
				t.Errorf("currentThreshold(%.0f) = %d, want %d", tt.pct, got, tt.want)
			}
		})
	}
}

func TestThresholdMessage(t *testing.T) {
	msg := thresholdMessage(95, 96.5, "test-sess")
	if msg == "" {
		t.Error("expected non-empty message for threshold 95")
	}

	msg = thresholdMessage(80, 82.0, "test-sess")
	if msg == "" {
		t.Error("expected non-empty message for threshold 80")
	}
}

func TestContextMonitorRegistered(t *testing.T) {
	_, ok := registry["context-monitor"]
	if !ok {
		t.Error("context-monitor not registered")
	}
}
