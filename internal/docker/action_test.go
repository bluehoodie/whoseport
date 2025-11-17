package docker

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptAction(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedAction Action
	}{
		{
			name:           "choose stop",
			input:          "1\n",
			expectedAction: ActionStop,
		},
		{
			name:           "choose stop and remove",
			input:          "2\n",
			expectedAction: ActionStopAndRemove,
		},
		{
			name:           "choose force remove",
			input:          "3\n",
			expectedAction: ActionRemove,
		},
		{
			name:           "choose cancel",
			input:          "4\n",
			expectedAction: ActionCancel,
		},
		{
			name:           "default to cancel on empty input",
			input:          "\n",
			expectedAction: ActionCancel,
		},
		{
			name:           "invalid then valid choice",
			input:          "5\n1\n",
			expectedAction: ActionStop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			writer := &bytes.Buffer{}

			handler := NewActionHandlerWithIO(reader, writer)
			info := &ContainerInfo{
				Name:    "test-container",
				ShortID: "abc123def456",
			}

			action := handler.PromptAction(info)

			if action != tt.expectedAction {
				t.Errorf("PromptAction() = %v, want %v", action, tt.expectedAction)
			}

			// Verify that output contains the container name
			output := writer.String()
			if !strings.Contains(output, "test-container") {
				t.Errorf("Output should contain container name, got: %s", output)
			}
		})
	}
}
