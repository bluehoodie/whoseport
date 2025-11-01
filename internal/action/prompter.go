package action

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/bluehoodie/whoseport/internal/model"
)

// Prompter handles user prompts for process actions
// Following Dependency Inversion Principle: depends on io interfaces
type Prompter struct {
	reader      io.Reader
	writer      io.Writer
	colorBold   string
	colorYellow string
	colorCyan   string
	colorReset  string
}

// NewPrompter creates a new Prompter with stdin/stdout
func NewPrompter() *Prompter {
	return &Prompter{
		reader:      os.Stdin,
		writer:      os.Stdout,
		colorBold:   "\033[1m",
		colorYellow: "\033[33m",
		colorCyan:   "\033[36m",
		colorReset:  "\033[0m",
	}
}

// NewPrompterWithIO creates a new Prompter with custom reader/writer (for testing)
// Following Open/Closed Principle: extensible via constructor injection
func NewPrompterWithIO(reader io.Reader, writer io.Writer) *Prompter {
	return &Prompter{
		reader:      reader,
		writer:      writer,
		colorBold:   "\033[1m",
		colorYellow: "\033[33m",
		colorCyan:   "\033[36m",
		colorReset:  "\033[0m",
	}
}

// PromptKillAction prompts the user to select an action for the process
// Returns the selected signal and true if user wants to kill, or nil and false if cancelled
func (p *Prompter) PromptKillAction(info *model.ProcessInfo) (syscall.Signal, bool) {
	fmt.Fprintf(p.writer, "%s%s⚠️  Process %d (%s) - Select action:%s\n",
		p.colorBold, p.colorYellow, info.ID, info.Command, p.colorReset)
	fmt.Fprintf(p.writer, "  [1] SIGTERM (15) - Graceful termination\n")
	fmt.Fprintf(p.writer, "  [2] SIGKILL (9)  - Force kill (cannot be caught)\n")
	fmt.Fprintf(p.writer, "  [3] Cancel\n")
	fmt.Fprintf(p.writer, "%sChoice [3]:%s ", p.colorBold, p.colorReset)

	scanner := bufio.NewScanner(p.reader)

	for {
		if !scanner.Scan() {
			return 0, false
		}

		choice := strings.TrimSpace(scanner.Text())

		// Default to Cancel if empty
		if choice == "" {
			choice = "3"
		}

		switch choice {
		case "1":
			return syscall.SIGTERM, true
		case "2":
			return syscall.SIGKILL, true
		case "3":
			return 0, false
		default:
			fmt.Fprintf(p.writer, "%sInvalid choice. Please enter 1-3.%s\n", p.colorYellow, p.colorReset)
			fmt.Fprintf(p.writer, "%sChoice [3]:%s ", p.colorBold, p.colorReset)
		}
	}
}

// PromptKill asks the user if they want to kill the process
// Deprecated: Use PromptKillAction instead for a single-step prompt
func (p *Prompter) PromptKill(info *model.ProcessInfo) bool {
	fmt.Fprintf(p.writer, "%s%s⚠️  Do you want to kill process %d (%s)?%s [y/N]: ",
		p.colorBold, p.colorYellow, info.ID, info.Command, p.colorReset)

	scanner := bufio.NewScanner(p.reader)
	if !scanner.Scan() {
		return false
	}

	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "y" || response == "yes"
}

// PromptSignal prompts the user to select a signal to send
// Supports common signals (SIGTERM, SIGKILL, SIGHUP, SIGINT) and custom numeric input
// Following Single Responsibility Principle: only handles signal selection
func (p *Prompter) PromptSignal() (syscall.Signal, error) {
	// Display signal menu
	fmt.Fprintf(p.writer, "\n%s%sSelect signal to send:%s\n", p.colorBold, p.colorCyan, p.colorReset)
	fmt.Fprintf(p.writer, "  [1] SIGTERM (15) - Graceful termination\n")
	fmt.Fprintf(p.writer, "  [2] SIGKILL (9)  - Force kill (cannot be caught)\n")
	fmt.Fprintf(p.writer, "  [3] SIGHUP (1)   - Hangup (reload config)\n")
	fmt.Fprintf(p.writer, "  [4] SIGINT (2)   - Interrupt (Ctrl+C)\n")
	fmt.Fprintf(p.writer, "  [5] Custom       - Enter signal number (1-31)\n")
	fmt.Fprintf(p.writer, "%sChoice [1]:%s ", p.colorBold, p.colorReset)

	scanner := bufio.NewScanner(p.reader)

	for {
		if !scanner.Scan() {
			return 0, fmt.Errorf("failed to read input: %w", scanner.Err())
		}

		choice := strings.TrimSpace(scanner.Text())

		// Default to SIGTERM if empty
		if choice == "" {
			choice = "1"
		}

		switch choice {
		case "1":
			return syscall.SIGTERM, nil
		case "2":
			return syscall.SIGKILL, nil
		case "3":
			return syscall.SIGHUP, nil
		case "4":
			return syscall.SIGINT, nil
		case "5":
			return p.promptCustomSignal(scanner)
		default:
			fmt.Fprintf(p.writer, "%sInvalid choice. Please enter 1-5.%s\n", p.colorYellow, p.colorReset)
			fmt.Fprintf(p.writer, "%sChoice [1]:%s ", p.colorBold, p.colorReset)
		}
	}
}

// promptCustomSignal prompts for a custom signal number
// Validates range 1-31 (standard Unix signal range)
func (p *Prompter) promptCustomSignal(scanner *bufio.Scanner) (syscall.Signal, error) {
	fmt.Fprintf(p.writer, "%sEnter signal number (1-31):%s ", p.colorBold, p.colorReset)

	for {
		if !scanner.Scan() {
			return 0, fmt.Errorf("failed to read signal number: %w", scanner.Err())
		}

		input := strings.TrimSpace(scanner.Text())
		signalNum, err := strconv.Atoi(input)

		if err != nil {
			fmt.Fprintf(p.writer, "%sInvalid number. Please enter a number between 1-31:%s ", p.colorYellow, p.colorReset)
			continue
		}

		if signalNum < 1 || signalNum > 31 {
			fmt.Fprintf(p.writer, "%sSignal must be between 1-31. Please try again:%s ", p.colorYellow, p.colorReset)
			continue
		}

		return syscall.Signal(signalNum), nil
	}
}
