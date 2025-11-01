package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluehoodie/whoseport/internal/model"
)

// MockFileSystem simulates /proc filesystem for testing
type MockFileSystem struct {
	files map[string]string
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string]string),
	}
}

// AddFile adds a file to the mock filesystem
func (m *MockFileSystem) AddFile(path, content string) {
	m.files[path] = content
}

// ReadFile reads a file from the mock filesystem
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, ok := m.files[path]; ok {
		return []byte(content), nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// SetupMockProcFS creates a complete mock /proc filesystem for a PID
func SetupMockProcFS(pid string) *MockFileSystem {
	fs := NewMockFileSystem()

	fs.AddFile(fmt.Sprintf("/proc/%s/status", pid), SampleProcStatus())
	fs.AddFile(fmt.Sprintf("/proc/%s/stat", pid), SampleProcStat())
	fs.AddFile(fmt.Sprintf("/proc/%s/io", pid), SampleProcIO())
	fs.AddFile(fmt.Sprintf("/proc/%s/limits", pid), SampleProcLimits())
	fs.AddFile(fmt.Sprintf("/proc/%s/cmdline", pid), "node\x00/app/server.js\x00")
	fs.AddFile(fmt.Sprintf("/proc/%s/cwd", pid), "/app")
	fs.AddFile(fmt.Sprintf("/proc/%s/exe", pid), "/usr/local/bin/node")
	fs.AddFile(fmt.Sprintf("/proc/%s/environ", pid), strings.Repeat("KEY=value\x00", 15))

	fs.AddFile("/proc/net/tcp", SampleProcNetTCP())
	fs.AddFile("/proc/net/udp", SampleProcNetUDP())
	fs.AddFile("/proc/stat", SampleProcBootStat())

	return fs
}

// MockLsofExecutor simulates lsof command execution
type MockLsofExecutor struct {
	Output string
	Err    error
}

// Execute returns the mocked lsof output
func (m *MockLsofExecutor) Execute(port int) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return []byte(m.Output), nil
}

// TempDirHelper defines the interface for testing types that support TempDir
type TempDirHelper interface {
	TempDir() string
	Helper()
}

// CreateTempProcFS creates a temporary /proc-like directory structure for integration tests
func CreateTempProcFS(t TempDirHelper, pid string) string {
	t.Helper()

	tmpDir := t.TempDir()
	procDir := filepath.Join(tmpDir, "proc")

	// Create directory structure
	pidDir := filepath.Join(procDir, pid)
	os.MkdirAll(pidDir, 0755)
	os.MkdirAll(filepath.Join(procDir, "net"), 0755)

	// Write sample files
	os.WriteFile(filepath.Join(pidDir, "status"), []byte(SampleProcStatus()), 0644)
	os.WriteFile(filepath.Join(pidDir, "stat"), []byte(SampleProcStat()), 0644)
	os.WriteFile(filepath.Join(pidDir, "io"), []byte(SampleProcIO()), 0644)
	os.WriteFile(filepath.Join(pidDir, "limits"), []byte(SampleProcLimits()), 0644)
	os.WriteFile(filepath.Join(pidDir, "cmdline"), []byte("node\x00/app/server.js\x00"), 0644)

	os.WriteFile(filepath.Join(procDir, "net", "tcp"), []byte(SampleProcNetTCP()), 0644)
	os.WriteFile(filepath.Join(procDir, "net", "udp"), []byte(SampleProcNetUDP()), 0644)
	os.WriteFile(filepath.Join(procDir, "stat"), []byte(SampleProcBootStat()), 0644)

	return procDir
}

// MockProcessKiller simulates process killing without actually sending signals
type MockProcessKiller struct {
	KilledPIDs []int
	ShouldFail bool
}

// Kill records the PID and optionally returns an error
func (m *MockProcessKiller) Kill(pid int) error {
	if m.ShouldFail {
		return fmt.Errorf("mock kill failed for pid %d", pid)
	}
	m.KilledPIDs = append(m.KilledPIDs, pid)
	return nil
}

// MockPrompter simulates user prompts
type MockPrompter struct {
	Response bool
	Err      error
}

// Prompt returns the mocked response
func (m *MockPrompter) Prompt(info *model.ProcessInfo) (bool, error) {
	return m.Response, m.Err
}
