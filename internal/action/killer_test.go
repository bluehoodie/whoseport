package action

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

// TestKillerWithSIGTERM tests killing a process with SIGTERM (graceful shutdown)
func TestKillerWithSIGTERM(t *testing.T) {
	// Start a dummy process that can be killed
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill() // Cleanup in case test fails

	pid := cmd.Process.Pid
	killer := NewKiller()

	// Kill with SIGTERM
	err := killer.Kill(pid, syscall.SIGTERM)
	if err != nil {
		t.Errorf("Kill() with SIGTERM failed: %v", err)
	}

	// Wait for process to actually exit
	err = cmd.Wait()
	if err == nil {
		t.Error("Process should have been killed but Wait() returned no error")
	}

	// Verify process is dead by checking if we can signal it
	process, _ := os.FindProcess(pid)
	if err := process.Signal(syscall.Signal(0)); err == nil {
		t.Error("Process should be dead after SIGTERM but is still running")
	}
}

// TestKillerWithSIGKILL tests killing a process with SIGKILL (force kill)
func TestKillerWithSIGKILL(t *testing.T) {
	// Start a dummy process
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill()

	pid := cmd.Process.Pid
	killer := NewKiller()

	// Kill with SIGKILL
	err := killer.Kill(pid, syscall.SIGKILL)
	if err != nil {
		t.Errorf("Kill() with SIGKILL failed: %v", err)
	}

	// Wait for process to actually exit
	err = cmd.Wait()
	if err == nil {
		t.Error("Process should have been killed but Wait() returned no error")
	}

	// Verify process is dead
	process, _ := os.FindProcess(pid)
	if err := process.Signal(syscall.Signal(0)); err == nil {
		t.Error("Process should be dead after SIGKILL but is still running")
	}
}

// TestKillerWithSIGHUP tests killing a process with SIGHUP
func TestKillerWithSIGHUP(t *testing.T) {
	// SIGHUP typically causes process termination (unless handled)
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill()

	pid := cmd.Process.Pid
	killer := NewKiller()

	// Send SIGHUP
	err := killer.Kill(pid, syscall.SIGHUP)
	if err != nil {
		t.Errorf("Kill() with SIGHUP failed: %v", err)
	}

	// Wait for process to react (sleep exits on SIGHUP)
	err = cmd.Wait()
	if err == nil {
		t.Error("Process should have been killed but Wait() returned no error")
	}

	// Verify process received the signal
	process, _ := os.FindProcess(pid)
	if err := process.Signal(syscall.Signal(0)); err == nil {
		t.Error("Process should have exited after SIGHUP but is still running")
	}
}

// TestKillerWithSIGINT tests killing a process with SIGINT
func TestKillerWithSIGINT(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill()

	pid := cmd.Process.Pid
	killer := NewKiller()

	// Send SIGINT
	err := killer.Kill(pid, syscall.SIGINT)
	if err != nil {
		t.Errorf("Kill() with SIGINT failed: %v", err)
	}

	// Wait for process to react
	err = cmd.Wait()
	if err == nil {
		t.Error("Process should have been killed but Wait() returned no error")
	}

	// Verify process received the signal
	process, _ := os.FindProcess(pid)
	if err := process.Signal(syscall.Signal(0)); err == nil {
		t.Error("Process should have exited after SIGINT but is still running")
	}
}

// TestKillerWithCustomSignal tests killing with a custom numeric signal
func TestKillerWithCustomSignal(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill()

	pid := cmd.Process.Pid
	killer := NewKiller()

	// Send custom signal (SIGUSR1 = 10)
	err := killer.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		t.Errorf("Kill() with custom signal (SIGUSR1) failed: %v", err)
	}

	// SIGUSR1 doesn't kill by default, so process should still be alive
	// No need to wait; the signal was sent successfully
	process, _ := os.FindProcess(pid)
	if err := process.Signal(syscall.Signal(0)); err != nil {
		t.Error("Process should still be alive after SIGUSR1 (unless it has a handler)")
	}

	// Clean up the process
	cmd.Process.Kill()
}

// TestKillerWithInvalidPID tests error handling for non-existent PIDs
func TestKillerWithInvalidPID(t *testing.T) {
	killer := NewKiller()

	// Use a PID that definitely doesn't exist (very high number)
	invalidPID := 999999

	err := killer.Kill(invalidPID, syscall.SIGTERM)
	if err == nil {
		t.Error("Kill() should return error for invalid PID")
	}
}

// TestKillerBackwardCompatibility tests that the old behavior still works
// This ensures we don't break existing code that may be calling Kill()
func TestKillerBackwardCompatibility(t *testing.T) {
	// This test verifies that SIGTERM is a valid signal choice
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	defer cmd.Process.Kill()

	pid := cmd.Process.Pid
	killer := NewKiller()

	// Use SIGTERM as default (backward compatible behavior)
	err := killer.Kill(pid, syscall.SIGTERM)
	if err != nil {
		t.Errorf("Kill() with SIGTERM (default) failed: %v", err)
	}

	// Wait for process to exit
	err = cmd.Wait()
	if err == nil {
		t.Error("Process should have been killed but Wait() returned no error")
	}

	process, _ := os.FindProcess(pid)
	if err := process.Signal(syscall.Signal(0)); err == nil {
		t.Error("Process should be dead after SIGTERM")
	}
}

// TestKillerInterface verifies that Killer implements the expected interface
func TestKillerInterface(t *testing.T) {
	var _ Killer = (*killer)(nil) // Compile-time interface check
}
