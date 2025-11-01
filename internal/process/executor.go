package process

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

// LsofExecutor executes lsof command with grep to find listening processes.
type LsofExecutor struct{}

// NewLsofExecutor creates a new LsofExecutor.
func NewLsofExecutor() *LsofExecutor {
	return &LsofExecutor{}
}

// Execute runs lsof piped through grep to find processes listening on the specified port.
func (e *LsofExecutor) Execute(port int) ([]byte, error) {
	c1 := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	c2 := exec.Command("grep", "LISTEN")

	r, w := io.Pipe()
	defer r.Close()

	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	if err := c1.Start(); err != nil {
		return nil, fmt.Errorf("failed to start lsof: %w", err)
	}

	if err := c2.Start(); err != nil {
		return nil, fmt.Errorf("failed to start grep: %w", err)
	}

	c1.Wait()
	w.Close()
	c2.Wait()

	return b2.Bytes(), nil
}
