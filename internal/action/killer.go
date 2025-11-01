package action

import (
	"fmt"
	"os"
	"syscall"
)

// Killer handles process termination
type Killer struct{}

// NewKiller creates a new Killer
func NewKiller() *Killer {
	return &Killer{}
}

// Kill terminates a process by PID
func (k *Killer) Kill(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Check if the process exists by sending signal 0
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return fmt.Errorf("process with PID %d does not exist or is not accessible: %w", pid, err)
	}

	// Attempt to send SIGTERM
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	return nil
}
