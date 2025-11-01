package action

import (
	"fmt"
	"os"
	"syscall"
)

// Killer defines the interface for process termination
// Following SOLID Interface Segregation Principle
type Killer interface {
	Kill(pid int, signal syscall.Signal) error
}

// killer implements the Killer interface
type killer struct{}

// NewKiller creates a new Killer implementation
func NewKiller() Killer {
	return &killer{}
}

// Kill terminates a process by PID with the specified signal
// Follows Single Responsibility Principle: only sends signals to processes
func (k *killer) Kill(pid int, signal syscall.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Check if the process exists by sending signal 0
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return fmt.Errorf("process with PID %d does not exist or is not accessible: %w", pid, err)
	}

	// Send the requested signal
	err = process.Signal(signal)
	if err != nil {
		return fmt.Errorf("failed to send signal %v: %w", signal, err)
	}

	return nil
}
