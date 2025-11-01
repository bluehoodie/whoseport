package process

import (
	"errors"
	"testing"

	"github.com/bluehoodie/whoseport/internal/testutil"
)

func TestProcessRetriever_GetProcessByPort(t *testing.T) {
	tests := []struct {
		name       string
		port       int
		execOutput string
		execErr    error
		wantCmd    string
		wantPID    int
		wantErr    bool
	}{
		{
			name:       "successful retrieval",
			port:       8080,
			execOutput: testutil.SampleLsofOutput(),
			execErr:    nil,
			wantCmd:    "node",
			wantPID:    12345,
			wantErr:    false,
		},
		{
			name:       "executor fails",
			port:       8080,
			execOutput: "",
			execErr:    errors.New("lsof failed"),
			wantErr:    true,
		},
		{
			name:       "parser fails - invalid output",
			port:       8080,
			execOutput: "invalid output",
			execErr:    nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock executor
			mockExec := &testutil.MockLsofExecutor{
				Output: tt.execOutput,
				Err:    tt.execErr,
			}

			retriever := NewProcessRetriever(mockExec, NewLsofParser())
			info, err := retriever.GetProcessByPort(tt.port)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProcessByPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if info.Command != tt.wantCmd {
				t.Errorf("Command = %v, want %v", info.Command, tt.wantCmd)
			}

			if info.ID != tt.wantPID {
				t.Errorf("ID = %v, want %v", info.ID, tt.wantPID)
			}
		})
	}
}

func TestNewDefaultRetriever(t *testing.T) {
	retriever := NewDefaultRetriever()

	if retriever == nil {
		t.Fatal("NewDefaultRetriever() returned nil")
	}

	if retriever.executor == nil {
		t.Error("executor is nil")
	}

	if retriever.parser == nil {
		t.Error("parser is nil")
	}
}
