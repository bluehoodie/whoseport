package format

import "testing"

func TestVisualWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"ascii", "hello", 5},
		{"emoji", "🔍", 2},
		{"emoji variation selector", "⚙️  PROCESS IDENTITY", 20},
		{"multiple emoji", "🔍⚙️📊", 6},
		{"zwj sequence", "👩‍💻", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VisualWidth(tt.input)
			if got != tt.want {
				t.Fatalf("VisualWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
