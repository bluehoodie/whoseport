package action

import (
	"bytes"
	"strings"
	"syscall"
	"testing"

	"github.com/bluehoodie/whoseport/internal/model"
)

// TestPromptKillActionSIGTERM tests selecting SIGTERM from the combined menu
func TestPromptKillActionSIGTERM(t *testing.T) {
	input := "1\n" // User selects option 1 (SIGTERM)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	signal, shouldKill := prompter.PromptKillAction(info)

	if !shouldKill {
		t.Error("PromptKillAction() should return true for SIGTERM selection")
	}
	if signal != syscall.SIGTERM {
		t.Errorf("Expected SIGTERM, got %v", signal)
	}

	// Verify menu was displayed with process info
	outputStr := output.String()
	if !strings.Contains(outputStr, "12345") {
		t.Error("Menu should display process ID")
	}
	if !strings.Contains(outputStr, "test-process") {
		t.Error("Menu should display process command")
	}
	if !strings.Contains(outputStr, "SIGTERM") {
		t.Error("Menu should contain SIGTERM option")
	}
}

// TestPromptKillActionSIGKILL tests selecting SIGKILL from the combined menu
func TestPromptKillActionSIGKILL(t *testing.T) {
	input := "2\n" // User selects option 2 (SIGKILL)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	signal, shouldKill := prompter.PromptKillAction(info)

	if !shouldKill {
		t.Error("PromptKillAction() should return true for SIGKILL selection")
	}
	if signal != syscall.SIGKILL {
		t.Errorf("Expected SIGKILL, got %v", signal)
	}
}

// TestPromptKillActionCancel tests selecting Cancel from the combined menu
func TestPromptKillActionCancel(t *testing.T) {
	input := "3\n" // User selects option 3 (Cancel)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	_, shouldKill := prompter.PromptKillAction(info)

	if shouldKill {
		t.Error("PromptKillAction() should return false when user cancels")
	}
}

// TestPromptKillActionDefaultCancel tests that empty input defaults to Cancel
func TestPromptKillActionDefaultCancel(t *testing.T) {
	input := "\n" // User presses Enter without selecting anything
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	_, shouldKill := prompter.PromptKillAction(info)

	if shouldKill {
		t.Error("PromptKillAction() should default to Cancel when user presses Enter")
	}
}

// TestPromptKillActionInvalidOption tests handling of invalid menu selection
func TestPromptKillActionInvalidOption(t *testing.T) {
	input := "99\n1\n" // Invalid option first, then valid option (SIGTERM)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	signal, shouldKill := prompter.PromptKillAction(info)

	if !shouldKill {
		t.Error("PromptKillAction() should return true after recovering from invalid input")
	}
	if signal != syscall.SIGTERM {
		t.Errorf("Expected SIGTERM after recovery, got %v", signal)
	}

	// Verify error message was shown
	outputStr := output.String()
	if !strings.Contains(outputStr, "Invalid") && !strings.Contains(outputStr, "invalid") {
		t.Error("Should display error message for invalid input")
	}
}

// TestPromptKillActionEOF tests handling of EOF (e.g., Ctrl+D)
func TestPromptKillActionEOF(t *testing.T) {
	input := "" // Empty input simulates EOF
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	_, shouldKill := prompter.PromptKillAction(info)

	if shouldKill {
		t.Error("PromptKillAction() should return false on EOF")
	}
}

// TestPromptSignalSIGTERM tests selecting SIGTERM from the menu
func TestPromptSignalSIGTERM(t *testing.T) {
	input := "1\n" // User selects option 1 (SIGTERM)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	signal, err := prompter.PromptSignal()

	if err != nil {
		t.Errorf("PromptSignal() failed: %v", err)
	}
	if signal != syscall.SIGTERM {
		t.Errorf("Expected SIGTERM, got %v", signal)
	}

	// Verify menu was displayed
	outputStr := output.String()
	if !strings.Contains(outputStr, "SIGTERM") {
		t.Error("Menu should contain SIGTERM option")
	}
}

// TestPromptSignalSIGKILL tests selecting SIGKILL from the menu
func TestPromptSignalSIGKILL(t *testing.T) {
	input := "2\n" // User selects option 2 (SIGKILL)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	signal, err := prompter.PromptSignal()

	if err != nil {
		t.Errorf("PromptSignal() failed: %v", err)
	}
	if signal != syscall.SIGKILL {
		t.Errorf("Expected SIGKILL, got %v", signal)
	}
}

// TestPromptSignalSIGHUP tests selecting SIGHUP from the menu
func TestPromptSignalSIGHUP(t *testing.T) {
	input := "3\n" // User selects option 3 (SIGHUP)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	signal, err := prompter.PromptSignal()

	if err != nil {
		t.Errorf("PromptSignal() failed: %v", err)
	}
	if signal != syscall.SIGHUP {
		t.Errorf("Expected SIGHUP, got %v", signal)
	}
}

// TestPromptSignalSIGINT tests selecting SIGINT from the menu
func TestPromptSignalSIGINT(t *testing.T) {
	input := "4\n" // User selects option 4 (SIGINT)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	signal, err := prompter.PromptSignal()

	if err != nil {
		t.Errorf("PromptSignal() failed: %v", err)
	}
	if signal != syscall.SIGINT {
		t.Errorf("Expected SIGINT, got %v", signal)
	}
}

// TestPromptSignalCustomNumeric tests entering a custom signal number
func TestPromptSignalCustomNumeric(t *testing.T) {
	input := "5\n10\n" // User selects option 5 (Custom), then enters 10 (SIGUSR1)
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	signal, err := prompter.PromptSignal()

	if err != nil {
		t.Errorf("PromptSignal() with custom number failed: %v", err)
	}
	if signal != syscall.Signal(10) {
		t.Errorf("Expected signal 10 (SIGUSR1), got %v", signal)
	}

	// Verify custom prompt was displayed
	outputStr := output.String()
	if !strings.Contains(outputStr, "Enter signal number") || !strings.Contains(outputStr, "1-31") {
		t.Error("Custom signal prompt should ask for signal number (1-31)")
	}
}

// TestPromptSignalInvalidOption tests handling of invalid menu selection
func TestPromptSignalInvalidOption(t *testing.T) {
	input := "99\n1\n" // Invalid option first, then valid option
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	signal, err := prompter.PromptSignal()

	if err != nil {
		t.Errorf("PromptSignal() should recover from invalid input: %v", err)
	}
	if signal != syscall.SIGTERM {
		t.Errorf("Expected SIGTERM after recovery, got %v", signal)
	}

	// Verify error message was shown
	outputStr := output.String()
	if !strings.Contains(outputStr, "Invalid") && !strings.Contains(outputStr, "invalid") {
		t.Error("Should display error message for invalid input")
	}
}

// TestPromptSignalInvalidCustomNumber tests validation of custom signal numbers
func TestPromptSignalInvalidCustomNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"zero", "5\n0\n15\n"},          // 0 is invalid, then enter valid 15 (SIGTERM)
		{"negative", "5\n-1\n15\n"},     // negative is invalid
		{"too high", "5\n99\n15\n"},     // >31 is invalid
		{"non-numeric", "5\nabc\n15\n"}, // non-number is invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			output := &bytes.Buffer{}

			prompter := NewPrompterWithIO(reader, output)
			signal, err := prompter.PromptSignal()

			if err != nil {
				t.Errorf("PromptSignal() should recover from invalid custom number: %v", err)
			}
			if signal != syscall.SIGTERM {
				t.Errorf("Expected SIGTERM (15) after recovery, got %v", signal)
			}

			outputStr := output.String()
			// Check for any error message (Invalid/invalid/Signal must)
			hasError := strings.Contains(outputStr, "Invalid") ||
				strings.Contains(outputStr, "invalid") ||
				strings.Contains(outputStr, "Signal must")
			if !hasError {
				t.Error("Should display error message for invalid custom number")
			}
		})
	}
}

// TestPromptSignalMenuDisplay tests that the menu shows all required options
func TestPromptSignalMenuDisplay(t *testing.T) {
	input := "1\n" // Just select something to complete
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	_, err := prompter.PromptSignal()

	if err != nil {
		t.Errorf("PromptSignal() failed: %v", err)
	}

	outputStr := output.String()
	requiredOptions := []string{
		"SIGTERM",
		"SIGKILL",
		"SIGHUP",
		"SIGINT",
		"Custom",
	}

	for _, option := range requiredOptions {
		if !strings.Contains(outputStr, option) {
			t.Errorf("Menu should contain option: %s", option)
		}
	}
}

// TestPromptSignalEOF tests handling of EOF (e.g., Ctrl+D)
func TestPromptSignalEOF(t *testing.T) {
	input := "" // Empty input simulates EOF
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)
	_, err := prompter.PromptSignal()

	if err == nil {
		t.Error("PromptSignal() should return error on EOF")
	}
}

// TestPromptKillYes tests the existing PromptKill with "yes" input
func TestPromptKillYes(t *testing.T) {
	input := "y\n"
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)

	// Create a mock ProcessInfo for testing
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	shouldKill := prompter.PromptKill(info)

	if !shouldKill {
		t.Error("PromptKill() should return true for 'y' input")
	}
}

// TestPromptKillNo tests the existing PromptKill with "no" input
func TestPromptKillNo(t *testing.T) {
	input := "n\n"
	reader := strings.NewReader(input)
	output := &bytes.Buffer{}

	prompter := NewPrompterWithIO(reader, output)

	// Create a mock ProcessInfo for testing
	info := &model.ProcessInfo{
		ID:      12345,
		Command: "test-process",
	}

	shouldKill := prompter.PromptKill(info)

	if shouldKill {
		t.Error("PromptKill() should return false for 'n' input")
	}
}
