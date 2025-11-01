package interactive

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/bluehoodie/whoseport/internal/display/format"
)

func captureOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func TestPrintModernSectionAlignment(t *testing.T) {
	d := NewDisplayer()

	output := captureOutput(func() {
		d.printModernSection("⚙️  PROCESS IDENTITY")
	})

	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for section, got %d", len(lines))
	}

	topWidth := format.VisualWidth(lines[0])
	midWidth := format.VisualWidth(lines[1])
	bottomWidth := format.VisualWidth(lines[2])

	if topWidth != midWidth || midWidth != bottomWidth {
		t.Fatalf("section borders misaligned: top=%d mid=%d bottom=%d", topWidth, midWidth, bottomWidth)
	}
}
