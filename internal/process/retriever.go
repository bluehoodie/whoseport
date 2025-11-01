package process

import (
	"fmt"

	"github.com/bluehoodie/whoseport/internal/model"
)

// ProcessRetriever orchestrates process information retrieval.
type ProcessRetriever struct {
	executor Executor
	parser   Parser
}

// NewProcessRetriever creates a new ProcessRetriever with the given executor and parser.
func NewProcessRetriever(executor Executor, parser Parser) *ProcessRetriever {
	return &ProcessRetriever{
		executor: executor,
		parser:   parser,
	}
}

// NewDefaultRetriever creates a ProcessRetriever with default lsof-based implementations.
func NewDefaultRetriever() *ProcessRetriever {
	return NewProcessRetriever(
		NewLsofExecutor(),
		NewLsofParser(),
	)
}

// GetProcessByPort retrieves process information for the given port.
func (r *ProcessRetriever) GetProcessByPort(port int) (*model.ProcessInfo, error) {
	output, err := r.executor.Execute(port)
	if err != nil {
		return nil, fmt.Errorf("failed to execute lsof: %w", err)
	}

	info, err := r.parser.Parse(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsof output: %w", err)
	}

	return info, nil
}
