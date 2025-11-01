package model

import (
	"encoding/json"
	"testing"
)

// TestNew verifies the constructor creates a ProcessInfo with correct basic fields
func TestNew(t *testing.T) {
	info := New("node", 12345, "testuser", "21u", "IPv4", "0x123456", "0t0", "TCP", "*:8080 (LISTEN)")

	if info.Command != "node" {
		t.Errorf("Command = %v, want node", info.Command)
	}
	if info.ID != 12345 {
		t.Errorf("ID = %v, want 12345", info.ID)
	}
	if info.User != "testuser" {
		t.Errorf("User = %v, want testuser", info.User)
	}
	if info.FD != "21u" {
		t.Errorf("FD = %v, want 21u", info.FD)
	}
	if info.Type != "IPv4" {
		t.Errorf("Type = %v, want IPv4", info.Type)
	}
	if info.Name != "*:8080 (LISTEN)" {
		t.Errorf("Name = %v, want *:8080 (LISTEN)", info.Name)
	}

	// Enhanced fields should be zero values
	if info.FullCommand != "" {
		t.Errorf("FullCommand should be empty, got %v", info.FullCommand)
	}
	if info.PPid != 0 {
		t.Errorf("PPid should be 0, got %v", info.PPid)
	}
}

// TestProcessInfoJSONMarshaling verifies ProcessInfo can be marshaled/unmarshaled
func TestProcessInfoJSONMarshaling(t *testing.T) {
	original := New("test", 123, "user", "1u", "IPv4", "0x1", "0t0", "TCP", "*:80")
	original.FullCommand = "/usr/bin/test --flag"
	original.PPid = 1
	original.MemoryRSS = 1024

	// Marshal to JSON
	jsonBytes, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded ProcessInfo
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify key fields
	if decoded.Command != original.Command {
		t.Errorf("Command = %v, want %v", decoded.Command, original.Command)
	}
	if decoded.ID != original.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, original.ID)
	}
	if decoded.FullCommand != original.FullCommand {
		t.Errorf("FullCommand = %v, want %v", decoded.FullCommand, original.FullCommand)
	}
	if decoded.PPid != original.PPid {
		t.Errorf("PPid = %v, want %v", decoded.PPid, original.PPid)
	}
	if decoded.MemoryRSS != original.MemoryRSS {
		t.Errorf("MemoryRSS = %v, want %v", decoded.MemoryRSS, original.MemoryRSS)
	}
}

// TestProcessInfoZeroValue verifies the zero value is safe to use
func TestProcessInfoZeroValue(t *testing.T) {
	var info ProcessInfo

	// Should not panic when accessing fields
	_ = info.Command
	_ = info.ID
	_ = info.TCPConns
	_ = info.UDPConns

	// Slices should be nil but safe
	if len(info.TCPConns) != 0 {
		t.Errorf("Zero value TCPConns should have length 0")
	}
}

// TestProcessInfoFieldCount documents the total number of fields
func TestProcessInfoFieldCount(t *testing.T) {
	// This test serves as documentation for the struct size
	// If this test fails, it means fields were added/removed
	info := ProcessInfo{}

	jsonBytes, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &fields); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// ProcessInfo has 40 JSON fields total (9 basic + 19 enhanced + 12 additional)
	expectedFieldCount := 40
	if len(fields) != expectedFieldCount {
		t.Errorf("ProcessInfo has %d JSON fields, expected %d. Update documentation if this is intentional.",
			len(fields), expectedFieldCount)
	}
}
