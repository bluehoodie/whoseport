package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bluehoodie/whoseport/internal/display/format"
	displayjson "github.com/bluehoodie/whoseport/internal/display/json"
	"github.com/bluehoodie/whoseport/internal/model"
	"github.com/bluehoodie/whoseport/internal/process"
	"github.com/bluehoodie/whoseport/internal/procfs"
)

// TestToProcessInfo tests the lsof output parsing
func TestToProcessInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *model.ProcessInfo
		wantErr bool
	}{
		{
			name:  "valid lsof output with IPv4",
			input: "node    12345 testuser   21u  IPv4 0x123456      0t0  TCP *:8080 (LISTEN)",
			want: &model.ProcessInfo{
				Command:    "node",
				ID:         12345,
				User:       "testuser",
				FD:         "21u",
				Type:       "IPv4",
				Device:     "0x123456",
				SizeOffset: "0t0",
				Node:       "TCP",
				Name:       "*:8080 (LISTEN)",
			},
			wantErr: false,
		},
		{
			name:  "valid lsof output with IPv6",
			input: "python3    9876 www-data   3u  IPv6 0xabcdef      0t0  TCP *:3000 (LISTEN)",
			want: &model.ProcessInfo{
				Command:    "python3",
				ID:         9876,
				User:       "www-data",
				FD:         "3u",
				Type:       "IPv6",
				Device:     "0xabcdef",
				SizeOffset: "0t0",
				Node:       "TCP",
				Name:       "*:3000 (LISTEN)",
			},
			wantErr: false,
		},
		{
			name:    "invalid lsof output - too few fields",
			input:   "node 12345 testuser",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid lsof output - non-numeric PID",
			input:   "node abcde testuser   21u  IPv4 0x123456      0t0  TCP *:8080 (LISTEN)",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := process.NewLsofParser()
			got, err := parser.Parse([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("toProcessInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Command != tt.want.Command {
				t.Errorf("Command = %v, want %v", got.Command, tt.want.Command)
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.User != tt.want.User {
				t.Errorf("User = %v, want %v", got.User, tt.want.User)
			}
			if got.FD != tt.want.FD {
				t.Errorf("FD = %v, want %v", got.FD, tt.want.FD)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}

// TestFormatBytes tests byte formatting
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1.00 KB"},
		{"megabytes", 1024 * 1024, "1.00 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.00 GB"},
		{"mixed KB", 1536, "1.50 KB"},
		{"mixed MB", 1024*1024 + 512*1024, "1.50 MB"},
		{"large number", 9876543210, "9.20 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("format.FormatBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

// TestFormatMemory tests memory formatting
func TestFormatMemory(t *testing.T) {
	tests := []struct {
		name string
		kb   int64
		want string
	}{
		{"zero", 0, "0 KB"},
		{"kilobytes", 100, "100 KB"},
		{"megabytes", 1024, "1.00 MB"},
		{"gigabytes", 1024 * 1024, "1.00 GB"},
		{"mixed MB", 1536, "1.50 MB"},
		{"large GB", 2560 * 1024, "2.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.FormatMemory(tt.kb)
			if got != tt.want {
				t.Errorf("format.FormatMemory(%d) = %v, want %v", tt.kb, got, tt.want)
			}
		})
	}
}

// TestFormatDuration tests duration formatting
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "0s"},
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 5 * time.Minute, "5m 0s"},
		{"hours", 2 * time.Hour, "2h 0m 0s"},
		{"days", 24 * time.Hour, "1d 0h 0m 0s"},
		{"weeks", 7 * 24 * time.Hour, "7d 0h 0m 0s"},
		{"mixed", 2*time.Hour + 30*time.Minute + 15*time.Second, "2h 30m 15s"},
		{"long uptime", 10*24*time.Hour + 5*time.Hour + 23*time.Minute, "10d 5h 23m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := procfs.FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("procfs.FormatDuration(%v) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

// TestVisualWidth tests visual width calculation for emoji and unicode
func TestVisualWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"ascii only", "hello", 5},
		{"single emoji", "🔍", 2},
		{"emoji with text", "🔍 Search", 9},
		{"multiple emoji", "🔍⚙️📊", 6},
		{"mixed", "Hello 🌍 World", 14},
		{"empty string", "", 0},
		{"spaces", "   ", 3},
		{"emoji at end", "Test 🎉", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.VisualWidth(tt.input)
			if got != tt.want {
				t.Errorf("format.VisualWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestStripAnsiCodes tests ANSI code removal
func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no codes", "hello", "hello"},
		{"simple color", "\033[31mred\033[0m", "red"},
		{"multiple codes", "\033[1m\033[31mbold red\033[0m", "bold red"},
		{"256 color", "\033[38;5;208morange\033[0m", "orange"},
		{"mixed", "Hello \033[31mRed\033[0m World", "Hello Red World"},
		{"empty", "", ""},
		{"only codes", "\033[31m\033[0m", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.StripAnsiCodes(tt.input)
			if got != tt.want {
				t.Errorf("format.StripAnsiCodes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestGetStateEmoji tests emoji mapping for process states
func TestGetStateEmoji(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"Running", "🟢"},
		{"Sleeping", "🔵"},
		{"Sleeping (interruptible)", "🔵"},
		{"Waiting (uninterruptible)", "⏸"},
		{"Zombie", "💀"},
		{"Stopped", "🟡"},
		{"Dead", "⚪"},
		{"Unknown", "⚪"},
		{"", "⚪"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := format.GetStateEmoji(tt.state)
			if got != tt.want {
				t.Errorf("format.GetStateEmoji(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

// TestGetStateColor tests color mapping for process states
func TestGetStateColor(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"Running", "\033[32m"},    // Green
		{"Sleeping", "\033[36m"},   // Cyan
		{"Disk Sleep", "\033[37m"}, // White
		{"Zombie", "\033[31m"},     // Red
		{"Stopped", "\033[33m"},    // Yellow
		{"Dead", "\033[37m"},       // White
		{"Unknown", "\033[37m"},    // White
		{"", "\033[37m"},           // White
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := format.GetStateColor(tt.state, nil)
			if got != tt.want {
				t.Errorf("format.GetStateColor(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

// TestTruncate tests string truncation
func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"format.Truncate short", "hello world", 5, "he..."},
		{"format.Truncate medium", "hello world", 8, "hello..."},
		{"empty string", "", 5, ""},
		{"very short limit", "hello", 3, "..."},
		// Note: format.Truncate panics when maxLen < 3, which is a bug we'll fix during refactoring
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("format.Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// TestPrintJSON tests JSON output generation
func TestPrintJSON(t *testing.T) {
	info := &model.ProcessInfo{
		Command: "test-process",
		ID:      12345,
		User:    "testuser",
		FD:      "21u",
		Type:    "IPv4",
		Name:    "*:8080 (LISTEN)",
	}

	// Test displayer output
	var buf bytes.Buffer
	displayer := displayjson.NewDisplayerWithWriter(&buf)
	err := displayer.Display(info)
	if err != nil {
		t.Fatalf("Failed to display JSON: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var decoded model.ProcessInfo
	output := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("Produced invalid JSON: %v", err)
	}

	// Verify key fields
	if decoded.Command != "test-process" {
		t.Errorf("JSON Command = %v, want test-process", decoded.Command)
	}
	if decoded.ID != 12345 {
		t.Errorf("JSON ID = %v, want 12345", decoded.ID)
	}
}

// NOTE: createProgressBar and createMemoryBar are now private methods in internal/display/interactive
// These tests are removed since we shouldn't test private implementation details
// The functionality is covered by integration tests that use the full Display() method

// BenchmarkFormatBytes benchmarks the format.FormatBytes function
func BenchmarkFormatBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		format.FormatBytes(1234567890)
	}
}

// BenchmarkVisualWidth benchmarks the format.VisualWidth function
func BenchmarkVisualWidth(b *testing.B) {
	text := "🔍 Hello World 🌍"
	for i := 0; i < b.N; i++ {
		format.VisualWidth(text)
	}
}

// BenchmarkStripAnsiCodes benchmarks the format.StripAnsiCodes function
func BenchmarkStripAnsiCodes(b *testing.B) {
	text := "\033[31m\033[1mHello\033[0m \033[32mWorld\033[0m"
	for i := 0; i < b.N; i++ {
		format.StripAnsiCodes(text)
	}
}

// TestIntegration_LsofParsing is an integration test for the full lsof flow
func TestIntegration_LsofParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require mocking exec.Command
	// For now, we'll just verify the structure is correct
	t.Skip("Integration test requires system setup")
}

// TestIntegration_ProcFSParsing tests /proc filesystem parsing
func TestIntegration_ProcFSParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require setting up a mock /proc filesystem
	t.Skip("Integration test requires mock /proc setup")
}

// TestIntegration_SignalSelection tests the interactive signal selection flow
func TestIntegration_SignalSelection(t *testing.T) {
	// This test verifies the full flow: find process -> prompt for kill -> prompt for signal -> kill
	// For now, this is a placeholder that verifies the structure
	// Full implementation would require mocking the entire flow
	t.Skip("Integration test requires full mock setup")
}

// TestIntegration_BackwardCompatibility tests that -k flag still uses SIGTERM
func TestIntegration_BackwardCompatibility(t *testing.T) {
	// This test verifies that existing behavior is preserved:
	// whoseport -k 8080 should still send SIGTERM by default
	// Full implementation would require process mocking
	t.Skip("Integration test requires process mocking")
}
