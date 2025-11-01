//go:build linux

package main

import (
	"testing"

	"github.com/bluehoodie/whoseport/internal/procfs"
)

// TestExpandState tests process state expansion (Linux-specific)
func TestExpandState(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"R", "Running"},
		{"S", "Sleeping (interruptible)"},
		{"D", "Waiting (uninterruptible)"},
		{"Z", "Zombie"},
		{"T", "Stopped"},
		{"t", "Tracing stop"},
		{"X", "Dead"},
		{"x", "x"}, // Unknown state returns as-is
		{"I", "Idle"},
		{"?", "?"}, // Unknown state returns as-is
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

// TestHexToIP tests hex IP address conversion
func TestHexToIP(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		want string
	}{
		{"localhost", "0100007F", "127.0.0.1"},
		{"all zeros", "00000000", "0.0.0.0"},
		{"example IP", "0101A8C0", "192.168.1.1"},
		{"another IP", "0A00020A", "10.2.0.10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := procfs.HexToIP(tt.hex)
			if got != tt.want {
				t.Errorf("procfs.HexToIP(%q) = %v, want %v", tt.hex, got, tt.want)
			}
		})
	}
}

// TestParseTCPState tests TCP state code parsing
func TestParseTCPState(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"01", "ESTABLISHED"},
		{"02", "SYN_SENT"},
		{"03", "SYN_RECV"},
		{"04", "FIN_WAIT1"},
		{"05", "FIN_WAIT2"},
		{"06", "TIME_WAIT"},
		{"07", "CLOSE"},
		{"08", "CLOSE_WAIT"},
		{"09", "LAST_ACK"},
		{"0A", "LISTEN"},
		{"0B", "CLOSING"},
		{"FF", "FF"}, // Unknown state returns hex string
		{"", ""},     // Empty returns empty
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			got := procfs.ParseTCPState(tt.hex)
			if got != tt.want {
				t.Errorf("procfs.ParseTCPState(%q) = %v, want %v", tt.hex, got, tt.want)
			}
		})
	}
}

// TestParseAddress tests network address parsing
func TestParseAddress(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		want string
	}{
		{"localhost:8080", "0100007F:1F90", "127.0.0.1:8080"},
		{"all interfaces:3000", "00000000:0BB8", "*:3000"},
		{"localhost:3306", "0100007F:0CEA", "127.0.0.1:3306"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := procfs.ParseAddress(tt.hex)
			if got != tt.want {
				t.Errorf("procfs.ParseAddress(%q) = %v, want %v", tt.hex, got, tt.want)
			}
		})
	}
}
