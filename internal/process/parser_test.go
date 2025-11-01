package process

import (
	"testing"

	"github.com/bluehoodie/whoseport/internal/testutil"
)

func TestLsofParser_Parse(t *testing.T) {
	parser := NewLsofParser()

	tests := []struct {
		name    string
		input   string
		wantCmd string
		wantPID int
		wantErr bool
	}{
		{
			name:    "valid lsof output",
			input:   testutil.SampleLsofOutput(),
			wantCmd: "node",
			wantPID: 12345,
			wantErr: false,
		},
		{
			name:    "valid lsof output with different process",
			input:   testutil.SampleLsofOutputMultipleFields(),
			wantCmd: "python3",
			wantPID: 9876,
			wantErr: false,
		},
		{
			name:    "invalid output - too few fields",
			input:   "node 12345",
			wantErr: true,
		},
		{
			name:    "invalid output - non-numeric PID",
			input:   "node abcd testuser 21u IPv4 0x123 0t0 TCP *:8080",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parser.Parse([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
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
