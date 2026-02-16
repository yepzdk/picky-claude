package checkers

import (
	"testing"
)

func TestForExtension(t *testing.T) {
	tests := []struct {
		ext      string
		wantName string
	}{
		{".py", "python"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".js", "typescript"},
		{".jsx", "typescript"},
		{".go", "go"},
		{".rs", ""},
		{".rb", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			c := ForExtension(tt.ext)
			if tt.wantName == "" {
				if c != nil {
					t.Errorf("ForExtension(%q) = %v, want nil", tt.ext, c.Name())
				}
				return
			}
			if c == nil {
				t.Fatalf("ForExtension(%q) = nil, want %q", tt.ext, tt.wantName)
			}
			if c.Name() != tt.wantName {
				t.Errorf("ForExtension(%q).Name() = %q, want %q", tt.ext, c.Name(), tt.wantName)
			}
		})
	}
}

func TestAllCheckersRegistered(t *testing.T) {
	names := map[string]bool{}
	for _, c := range registry {
		names[c.Name()] = true
	}

	want := []string{"python", "typescript", "go"}
	for _, name := range want {
		if !names[name] {
			t.Errorf("checker %q not registered", name)
		}
	}
}
