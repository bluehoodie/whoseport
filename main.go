package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

var (
	killFlag       bool
	interactiveFlag bool
	jsonFlag       bool
)

func main() {
	flag.BoolVar(&killFlag, "kill", false, "Kill the process using the port")
	flag.BoolVar(&killFlag, "k", false, "Kill the process using the port (shorthand)")
	flag.BoolVar(&interactiveFlag, "interactive", false, "Interactive mode - prompt before killing")
	flag.BoolVar(&interactiveFlag, "i", false, "Interactive mode - prompt before killing (shorthand)")
	flag.BoolVar(&jsonFlag, "json", false, "Output in JSON format")

	flag.Usage = func() {
		fmt.Printf("%s%sUsage of whoseport:%s\n", colorBold, colorCyan, colorReset)
		fmt.Printf("  %swhoseport [options] {port}%s\n\n", colorWhite, colorReset)
		fmt.Printf("%sOptions:%s\n", colorBold, colorReset)
		fmt.Printf("  %s-k, --kill%s         Kill the process using the port\n", colorYellow, colorReset)
		fmt.Printf("  %s-i, --interactive%s  Prompt before killing the process\n", colorYellow, colorReset)
		fmt.Printf("  %s--json%s             Output in JSON format\n", colorYellow, colorReset)
		fmt.Printf("\n%sExample:%s\n", colorBold, colorReset)
		fmt.Printf("  whoseport 8080\n")
		fmt.Printf("  whoseport -i 8080\n")
		fmt.Printf("  whoseport -k 8080\n")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Printf("%serror:%s missing port number\n", colorRed, colorReset)
		flag.Usage()
		os.Exit(1)
	}

	port, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Printf("%serror:%s port must be an integer\n", colorRed, colorReset)
		flag.Usage()
		os.Exit(1)
	}

	processInfo, err := lsof(port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}

	// Display the process info
	if jsonFlag {
		printJSON(processInfo)
	} else {
		printInteractive(processInfo, port)
	}

	// Handle kill logic
	if killFlag || interactiveFlag {
		shouldKill := killFlag

		if interactiveFlag && !killFlag {
			shouldKill = promptKill(processInfo)
		}

		if shouldKill {
			if err := killProcess(processInfo.ID); err != nil {
				fmt.Printf("%s✗ Failed to kill process:%s %v\n", colorRed, colorReset, err)
				os.Exit(1)
			}
			fmt.Printf("%s✓ Successfully killed process %d%s\n", colorGreen, processInfo.ID, colorReset)
		}
	}
}

type ProcessInfo struct {
	Command    string `json:"command"`
	ID         int    `json:"id"`
	User       string `json:"user"`
	FD         string `json:"fd"`
	Type       string `json:"type"`
	Device     string `json:"device"`
	SizeOffset string `json:"size_offset"`
	Node       string `json:"node"`
	Name       string `json:"name"`
}

func lsof(port int) (*ProcessInfo, error) {
	c1 := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	c2 := exec.Command("grep", "LISTEN")

	r, w := io.Pipe()
	defer r.Close()

	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	c1.Start()
	c2.Start()

	c1.Wait()
	w.Close()
	c2.Wait()

	d := &data{&b2}
	return d.toProcessInfo()
}

func printJSON(info *ProcessInfo) {
	j, err := json.MarshalIndent(info, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s failed to marshal JSON: %v\n", colorRed, colorReset, err)
		return
	}
	fmt.Fprintf(os.Stdout, "%s\n", j)
}

func printInteractive(info *ProcessInfo, port int) {
	fmt.Println()
	fmt.Printf("%s%s╔════════════════════════════════════════════════════════════════╗%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s║          Process Information for Port %-24d ║%s\n", colorBold, colorCyan, port, colorReset)
	fmt.Printf("%s%s╚════════════════════════════════════════════════════════════════╝%s\n", colorBold, colorCyan, colorReset)
	fmt.Println()
	fmt.Printf("  %s%-12s%s %s%s%s\n", colorYellow, "Command:", colorReset, colorBold, info.Command, colorReset)
	fmt.Printf("  %s%-12s%s %s%d%s\n", colorYellow, "PID:", colorReset, colorGreen, info.ID, colorReset)
	fmt.Printf("  %s%-12s%s %s\n", colorYellow, "User:", colorReset, info.User)
	fmt.Printf("  %s%-12s%s %s\n", colorYellow, "Type:", colorReset, info.Type)
	fmt.Printf("  %s%-12s%s %s\n", colorYellow, "Node:", colorReset, info.Node)
	fmt.Printf("  %s%-12s%s %s\n", colorYellow, "Name:", colorReset, info.Name)
	fmt.Println()
}

func promptKill(info *ProcessInfo) bool {
	fmt.Printf("%s⚠ Do you want to kill process %d (%s)?%s [y/N]: ", colorYellow, info.ID, info.Command, colorReset)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	return nil
}

type data struct {
	values *bytes.Buffer
}

func (i *data) toProcessInfo() (*ProcessInfo, error) {
	var values []string

	spl := bytes.Split(i.values.Bytes(), []byte(" "))
	for i, v := range spl {
		if len(v) <= 0 {
			continue
		}

		if len(values) == 8 {
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
		return nil, fmt.Errorf("could not convert process id to int: %v", err)
	}

	return &ProcessInfo{
		Command:    values[0],
		ID:         pid,
		User:       values[2],
		FD:         values[3],
		Type:       values[4],
		Device:     values[5],
		SizeOffset: values[6],
		Node:       values[7],
		Name:       values[8],
	}, nil
}
