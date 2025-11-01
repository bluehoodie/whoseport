// Package process provides interfaces and implementations for retrieving process information.
package process

import "github.com/bluehoodie/whoseport/internal/model"

// Executor executes the lsof command to find processes listening on a port.
type Executor interface {
	Execute(port int) ([]byte, error)
}

// Parser parses lsof command output into a ProcessInfo structure.
type Parser interface {
	Parse(output []byte) (*model.ProcessInfo, error)
}

// Retriever retrieves process information for a given port.
// It orchestrates the execution and parsing steps.
type Retriever interface {
	GetProcessByPort(port int) (*model.ProcessInfo, error)
}
