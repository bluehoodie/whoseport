package json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bluehoodie/whoseport/internal/model"
)

// Displayer outputs ProcessInfo as JSON
type Displayer struct {
	output io.Writer
}

// NewDisplayer creates a new JSON displayer
func NewDisplayer() *Displayer {
	return &Displayer{output: os.Stdout}
}

// NewDisplayerWithWriter creates a displayer with custom writer
func NewDisplayerWithWriter(w io.Writer) *Displayer {
	return &Displayer{output: w}
}

// Display outputs the ProcessInfo as formatted JSON
func (d *Displayer) Display(info *model.ProcessInfo) error {
	j, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Fprintf(d.output, "%s\n", j)
	return nil
}
