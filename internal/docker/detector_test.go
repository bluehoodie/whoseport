package docker

import (
	"testing"

	"github.com/bluehoodie/whoseport/internal/model"
)

func TestExtractContainerIPFromProxy(t *testing.T) {
	detector := &DefaultDetector{dockerAvailable: true}

	tests := []struct {
		name        string
		fullCommand string
		expected    string
	}{
		{
			name:        "valid docker-proxy command",
			fullCommand: "/usr/bin/docker-proxy -proto tcp -host-ip 0.0.0.0 -host-port 8080 -container-ip 172.17.0.2 -container-port 8080",
			expected:    "172.17.0.2",
		},
		{
			name:        "docker-proxy with ipv6",
			fullCommand: "/usr/bin/docker-proxy -proto tcp -host-ip :: -host-port 8080 -container-ip 172.17.0.3 -container-port 80",
			expected:    "172.17.0.3",
		},
		{
			name:        "no container-ip in command",
			fullCommand: "/usr/bin/docker-proxy -proto tcp -host-ip 0.0.0.0 -host-port 8080",
			expected:    "",
		},
		{
			name:        "empty command",
			fullCommand: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &model.ProcessInfo{
				Command:     "docker-proxy",
				FullCommand: tt.fullCommand,
			}

			result := detector.extractContainerIPFromProxy(info)
			if result != tt.expected {
				t.Errorf("extractContainerIPFromProxy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCheckCgroupForContainer(t *testing.T) {
	detector := &DefaultDetector{dockerAvailable: true}

	// Note: This test will only work on systems with /proc filesystem
	// For a process running in Docker, we expect to find a container ID
	// For a regular process, we expect empty string

	// Test with a non-existent PID (should return error or empty string)
	containerID, err := detector.checkCgroupForContainer(999999)
	if err == nil && containerID != "" {
		t.Errorf("Expected no container ID for non-existent PID, got: %s", containerID)
	}
}
