package process

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/bluehoodie/whoseport/internal/model"
)

// LsofParser parses lsof command output into ProcessInfo.
type LsofParser struct{}

// NewLsofParser creates a new LsofParser.
func NewLsofParser() *LsofParser {
	return &LsofParser{}
}

// Parse parses lsof output and returns a ProcessInfo.
// Expected format: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
func (p *LsofParser) Parse(output []byte) (*model.ProcessInfo, error) {
	var values []string

	spl := bytes.Split(output, []byte(" "))
	for i, v := range spl {
		if len(v) <= 0 {
			continue
		}

		if len(values) == 8 {
			// Remaining parts are the NAME field (e.g., "*:8080 (LISTEN)")
			values = append(values, strings.TrimSpace(string(bytes.Join(spl[i:], []byte(" ")))))
			break
		}

		values = append(values, strings.TrimSpace(string(v)))
	}

	if len(values) != 9 {
		return nil, fmt.Errorf("no service found on this port")
	}

	pid, err := strconv.Atoi(values[1])
	if err != nil {
		return nil, fmt.Errorf("could not convert process id to int: %w", err)
	}

	return model.New(
		values[0], // command
		pid,       // id
		values[2], // user
		values[3], // fd
		values[4], // type
		values[5], // device
		values[6], // size_offset
		values[7], // node
		values[8], // name
	), nil
}
