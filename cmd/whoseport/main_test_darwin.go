//go:build darwin

package main

import (
	"testing"

	"github.com/bluehoodie/whoseport/internal/procfs"
)

// TestExpandState tests process state expansion (macOS-specific)
func TestExpandState(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"R", "Running"},
		{"S", "Sleeping"},
		{"U", "Uninterruptible wait"},
		{"Z", "Zombie"},
		{"T", "Stopped"},
		{"I", "Idle"},
		{"x", "x"}, // Unknown state returns as-is
		{"", ""},   // Empty returns empty
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := procfs.ExpandState(tt.state)
			if got != tt.want {
				t.Errorf("procfs.ExpandState(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}
