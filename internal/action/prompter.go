package action

import (
	"fmt"
	"strings"

	"github.com/bluehoodie/whoseport/internal/model"
)

// Prompter handles user prompts for process actions
type Prompter struct {
	colorBold   string
	colorYellow string
	colorReset  string
}

// NewPrompter creates a new Prompter
func NewPrompter() *Prompter {
	return &Prompter{
		colorBold:   "\033[1m",
		colorYellow: "\033[33m",
		colorReset:  "\033[0m",
	}
}

// PromptKill asks the user if they want to kill the process
func (p *Prompter) PromptKill(info *model.ProcessInfo) bool {
	fmt.Printf("%s%s⚠️  Do you want to kill process %d (%s)?%s [y/N]: ",
		p.colorBold, p.colorYellow, info.ID, info.Command, p.colorReset)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
