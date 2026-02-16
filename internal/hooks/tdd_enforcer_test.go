package hooks

import "testing"

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Go
		{"/src/config/config_test.go", true},
		{"/src/config/config.go", false},

		// Python
		{"/tests/test_auth.py", true},
		{"/src/test_auth.py", true},
		{"/src/auth.py", false},

		// TypeScript/JavaScript
		{"/src/App.test.tsx", true},
		{"/src/App.spec.ts", true},
		{"/src/__tests__/App.tsx", true},
		{"/src/App.tsx", false},

		// Directory patterns
		{"/tests/unit/foo.py", true},
		{"/test/integration/bar.js", true},
		{"/spec/models/user_spec.rb", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isTestFile(tt.path)
			if got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestTddEnforcerRegistered(t *testing.T) {
	_, ok := registry["tdd-enforcer"]
	if !ok {
		t.Error("tdd-enforcer not registered")
	}
}
