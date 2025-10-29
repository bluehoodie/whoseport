package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	colorDim    = "\033[2m"
)

var (
	killFlag        bool
	noInteractive   bool
	jsonFlag        bool
)

func main() {
	flag.BoolVar(&killFlag, "kill", false, "Kill the process using the port")
	flag.BoolVar(&killFlag, "k", false, "Kill the process using the port (shorthand)")
	flag.BoolVar(&noInteractive, "no-interactive", false, "Disable interactive mode (show info only)")
	flag.BoolVar(&noInteractive, "n", false, "Disable interactive mode (shorthand)")
	flag.BoolVar(&jsonFlag, "json", false, "Output in JSON format")

	flag.Usage = func() {
		fmt.Printf("%s%sUsage of whoseport:%s\n", colorBold, colorCyan, colorReset)
		fmt.Printf("  %swhoseport [options] {port}%s\n\n", colorWhite, colorReset)
		fmt.Printf("%sOptions:%s\n", colorBold, colorReset)
		fmt.Printf("  %s-k, --kill%s            Kill the process using the port\n", colorYellow, colorReset)
		fmt.Printf("  %s-n, --no-interactive%s  Disable interactive mode (show info only)\n", colorYellow, colorReset)
		fmt.Printf("  %s--json%s                Output in JSON format\n", colorYellow, colorReset)
		fmt.Printf("\n%sNote:%s Interactive mode is enabled by default\n", colorBold, colorReset)
		fmt.Printf("\n%sExample:%s\n", colorBold, colorReset)
		fmt.Printf("  whoseport 8080           # Show detailed info with interactive prompt\n")
		fmt.Printf("  whoseport -n 8080        # Show info only, no prompt\n")
		fmt.Printf("  whoseport -k 8080        # Kill without prompting\n")
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

	// Enhance with detailed process information
	enhanceProcessInfo(processInfo)

	// Display the process info
	if jsonFlag {
		printJSON(processInfo)
	} else {
		printInteractive(processInfo, port)
	}

	// Handle kill logic - interactive by default unless disabled
	if killFlag {
		// Direct kill without prompting
		if err := killProcess(processInfo.ID); err != nil {
			fmt.Printf("%s✗ Failed to kill process:%s %v\n", colorRed, colorReset, err)
			os.Exit(1)
		}
		fmt.Printf("%s✓ Successfully killed process %d%s\n", colorGreen, processInfo.ID, colorReset)
	} else if !noInteractive {
		// Interactive mode is default
		if promptKill(processInfo) {
			if err := killProcess(processInfo.ID); err != nil {
				fmt.Printf("%s✗ Failed to kill process:%s %v\n", colorRed, colorReset, err)
				os.Exit(1)
			}
			fmt.Printf("%s✓ Successfully killed process %d%s\n", colorGreen, processInfo.ID, colorReset)
		}
	}
}

type ProcessInfo struct {
	// Basic info from lsof
	Command    string `json:"command"`
	ID         int    `json:"id"`
	User       string `json:"user"`
	FD         string `json:"fd"`
	Type       string `json:"type"`
	Device     string `json:"device"`
	SizeOffset string `json:"size_offset"`
	Node       string `json:"node"`
	Name       string `json:"name"`

	// Enhanced details from /proc
	FullCommand    string   `json:"full_command"`
	PPid           int      `json:"ppid"`
	ParentCommand  string   `json:"parent_command"`
	State          string   `json:"state"`
	Threads        int      `json:"threads"`
	WorkingDir     string   `json:"working_dir"`
	MemoryRSS      int64    `json:"memory_rss_kb"`
	MemoryVMS      int64    `json:"memory_vms_kb"`
	CPUTime        float64  `json:"cpu_time_seconds"`
	StartTime      string   `json:"start_time"`
	Uptime         string   `json:"uptime"`
	OpenFDs        int      `json:"open_fds"`
	MaxFDs         int      `json:"max_fds"`
	UID            int      `json:"uid"`
	GID            int      `json:"gid"`
	Groups         string   `json:"groups"`
	NetworkConns   int      `json:"network_connections"`
	TCPConns       []string `json:"tcp_connections"`
	UDPConns       []string `json:"udp_connections"`
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

func enhanceProcessInfo(info *ProcessInfo) {
	pid := info.ID

	// Read /proc/[pid]/cmdline for full command
	if cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil {
		info.FullCommand = strings.ReplaceAll(string(bytes.TrimSpace(cmdline)), "\x00", " ")
		if info.FullCommand == "" {
			info.FullCommand = info.Command
		}
	}

	// Read /proc/[pid]/status for detailed information
	if status, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid)); err == nil {
		parseStatus(string(status), info)
	}

	// Read /proc/[pid]/stat for timing information
	if stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
		parseStat(string(stat), info)
	}

	// Get working directory
	if cwd, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid)); err == nil {
		info.WorkingDir = cwd
	}

	// Count open file descriptors
	if fds, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid)); err == nil {
		info.OpenFDs = len(fds)
	}

	// Get parent process info
	if info.PPid > 0 {
		if cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", info.PPid)); err == nil {
			parentCmd := strings.ReplaceAll(string(bytes.TrimSpace(cmdline)), "\x00", " ")
			if parentCmd != "" {
				parts := strings.Fields(parentCmd)
				if len(parts) > 0 {
					info.ParentCommand = filepath.Base(parts[0])
				}
			}
		}
		if info.ParentCommand == "" {
			if comm, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", info.PPid)); err == nil {
				info.ParentCommand = strings.TrimSpace(string(comm))
			}
		}
	}

	// Get network connections for this PID
	info.TCPConns = getNetworkConnections(pid, "tcp")
	info.UDPConns = getNetworkConnections(pid, "udp")
	info.NetworkConns = len(info.TCPConns) + len(info.UDPConns)
}

func parseStatus(status string, info *ProcessInfo) {
	lines := strings.Split(status, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		switch key {
		case "PPid":
			info.PPid, _ = strconv.Atoi(fields[1])
		case "State":
			info.State = expandState(fields[1])
		case "Threads":
			info.Threads, _ = strconv.Atoi(fields[1])
		case "VmRSS":
			info.MemoryRSS, _ = strconv.ParseInt(fields[1], 10, 64)
		case "VmSize":
			info.MemoryVMS, _ = strconv.ParseInt(fields[1], 10, 64)
		case "Uid":
			if len(fields) > 1 {
				info.UID, _ = strconv.Atoi(fields[1])
			}
		case "Gid":
			if len(fields) > 1 {
				info.GID, _ = strconv.Atoi(fields[1])
			}
		case "Groups":
			if len(fields) > 1 {
				info.Groups = strings.Join(fields[1:], " ")
			}
		case "FDSize":
			info.MaxFDs, _ = strconv.Atoi(fields[1])
		}
	}
}

func parseStat(stat string, info *ProcessInfo) {
	fields := strings.Fields(stat)
	if len(fields) < 22 {
		return
	}

	// CPU time (user + system) in clock ticks
	utime, _ := strconv.ParseInt(fields[13], 10, 64)
	stime, _ := strconv.ParseInt(fields[14], 10, 64)
	clockTicks := float64(utime + stime)
	info.CPUTime = clockTicks / 100.0 // Convert to seconds (assuming 100 Hz)

	// Start time
	starttime, _ := strconv.ParseInt(fields[21], 10, 64)
	bootTime := getBootTime()
	startTimeSec := bootTime + (starttime / 100) // Convert clock ticks to seconds
	startTimeObj := time.Unix(startTimeSec, 0)
	info.StartTime = startTimeObj.Format("2006-01-02 15:04:05")

	// Calculate uptime
	uptime := time.Since(startTimeObj)
	info.Uptime = formatDuration(uptime)
}

func expandState(state string) string {
	states := map[string]string{
		"R": "Running",
		"S": "Sleeping (interruptible)",
		"D": "Waiting (uninterruptible)",
		"Z": "Zombie",
		"T": "Stopped",
		"t": "Tracing stop",
		"X": "Dead",
		"I": "Idle",
	}

	if len(state) > 0 {
		mainState := string(state[0])
		expanded := states[mainState]
		if expanded == "" {
			expanded = state
		}

		// Add modifiers
		if len(state) > 1 {
			modifiers := []string{}
			for _, c := range state[1:] {
				switch c {
				case '<':
					modifiers = append(modifiers, "high-priority")
				case 'N':
					modifiers = append(modifiers, "low-priority")
				case 'L':
					modifiers = append(modifiers, "pages locked")
				case 's':
					modifiers = append(modifiers, "session leader")
				case 'l':
					modifiers = append(modifiers, "multi-threaded")
				case '+':
					modifiers = append(modifiers, "foreground")
				}
			}
			if len(modifiers) > 0 {
				expanded += " (" + strings.Join(modifiers, ", ") + ")"
			}
		}
		return expanded
	}
	return state
}

func getBootTime() int64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Now().Unix()
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				btime, _ := strconv.ParseInt(fields[1], 10, 64)
				return btime
			}
		}
	}
	return time.Now().Unix()
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 || days > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 || hours > 0 || days > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	parts = append(parts, fmt.Sprintf("%ds", seconds))

	return strings.Join(parts, " ")
}

func getNetworkConnections(pid int, protocol string) []string {
	var connections []string

	// Map of inode to socket info
	inodes := getProcessInodes(pid)
	if len(inodes) == 0 {
		return connections
	}

	// Read /proc/net/tcp or /proc/net/udp
	file := fmt.Sprintf("/proc/net/%s", protocol)
	data, err := os.ReadFile(file)
	if err != nil {
		return connections
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		inode := fields[9]
		if _, exists := inodes[inode]; exists {
			localAddr := parseAddress(fields[1])
			remoteAddr := parseAddress(fields[2])
			state := ""
			if protocol == "tcp" && len(fields) > 3 {
				state = parseTCPState(fields[3])
			}

			connStr := fmt.Sprintf("%s -> %s", localAddr, remoteAddr)
			if state != "" {
				connStr += fmt.Sprintf(" [%s]", state)
			}
			connections = append(connections, connStr)
		}
	}

	return connections
}

func getProcessInodes(pid int) map[string]bool {
	inodes := make(map[string]bool)

	fdPath := fmt.Sprintf("/proc/%d/fd", pid)
	fds, err := os.ReadDir(fdPath)
	if err != nil {
		return inodes
	}

	for _, fd := range fds {
		link, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
		if err != nil {
			continue
		}

		if strings.HasPrefix(link, "socket:[") {
			inode := strings.TrimPrefix(link, "socket:[")
			inode = strings.TrimSuffix(inode, "]")
			inodes[inode] = true
		}
	}

	return inodes
}

func parseAddress(hex string) string {
	parts := strings.Split(hex, ":")
	if len(parts) != 2 {
		return hex
	}

	ip := hexToIP(parts[0])
	port, _ := strconv.ParseInt(parts[1], 16, 64)

	if ip == "0.0.0.0" || ip == "0:0:0:0:0:0:0:0" {
		return fmt.Sprintf("*:%d", port)
	}

	return fmt.Sprintf("%s:%d", ip, port)
}

func hexToIP(hex string) string {
	if len(hex) == 8 {
		// IPv4
		var parts []string
		for i := 6; i >= 0; i -= 2 {
			val, _ := strconv.ParseInt(hex[i:i+2], 16, 64)
			parts = append(parts, fmt.Sprintf("%d", val))
		}
		return strings.Join(parts, ".")
	}
	// For IPv6, return simplified
	return "IPv6"
}

func parseTCPState(hex string) string {
	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTEN",
		"0B": "CLOSING",
	}

	state, _ := strconv.ParseInt(hex, 16, 64)
	hexStr := fmt.Sprintf("%02X", state)
	if s, ok := states[hexStr]; ok {
		return s
	}
	return hex
}

func printJSON(info *ProcessInfo) {
	j, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s failed to marshal JSON: %v\n", colorRed, colorReset, err)
		return
	}
	fmt.Fprintf(os.Stdout, "%s\n", j)
}

func printInteractive(info *ProcessInfo, port int) {
	width := 80

	// Header
	fmt.Println()
	printBoxTop(width)
	printBoxLine(width, fmt.Sprintf("🔍 PROCESS DETAILS FOR PORT %d", port), colorCyan, true)
	printBoxBottom(width)
	fmt.Println()

	// Section 1: Process Identity
	printSectionHeader("⚙️  PROCESS IDENTITY", width)
	printField("Command", info.Command, colorGreen)
	printField("Full Command", truncate(info.FullCommand, 65), colorWhite)
	printField("Process ID (PID)", fmt.Sprintf("%d", info.ID), colorGreen)
	printField("Parent PID", fmt.Sprintf("%d", info.PPid), colorWhite)
	if info.ParentCommand != "" {
		printField("Parent Process", info.ParentCommand, colorWhite)
	}
	printField("User", info.User, colorYellow)
	printField("UID / GID", fmt.Sprintf("%d / %d", info.UID, info.GID), colorWhite)
	if info.Groups != "" {
		printField("Groups", info.Groups, colorDim)
	}
	fmt.Println()

	// Section 2: Process State
	printSectionHeader("📊 PROCESS STATE", width)
	printField("State", info.State, getStateColor(info.State))
	printField("Threads", fmt.Sprintf("%d", info.Threads), colorWhite)
	if info.StartTime != "" {
		printField("Started", info.StartTime, colorCyan)
	}
	if info.Uptime != "" {
		printField("Uptime", info.Uptime, colorGreen)
	}
	if info.CPUTime > 0 {
		printField("CPU Time", fmt.Sprintf("%.2f seconds", info.CPUTime), colorYellow)
	}
	fmt.Println()

	// Section 3: Memory Usage
	printSectionHeader("💾 MEMORY USAGE", width)
	if info.MemoryRSS > 0 {
		printField("Resident Set Size", formatMemory(info.MemoryRSS), colorGreen)
	}
	if info.MemoryVMS > 0 {
		printField("Virtual Memory", formatMemory(info.MemoryVMS), colorCyan)
	}
	fmt.Println()

	// Section 4: File System
	printSectionHeader("📁 FILE SYSTEM", width)
	if info.WorkingDir != "" {
		printField("Working Directory", info.WorkingDir, colorCyan)
	}
	printField("Open File Descriptors", fmt.Sprintf("%d / %d", info.OpenFDs, info.MaxFDs), colorYellow)
	fmt.Println()

	// Section 5: Network Information
	printSectionHeader("🌐 NETWORK", width)
	printField("Protocol", strings.ToUpper(info.Type), colorCyan)
	printField("Listening On", info.Name, colorGreen)
	printField("Node Type", info.Node, colorWhite)
	printField("File Descriptor", info.FD, colorDim)
	printField("Total Connections", fmt.Sprintf("%d", info.NetworkConns), colorYellow)

	if len(info.TCPConns) > 0 {
		fmt.Println()
		fmt.Printf("  %s%sTCP Connections:%s\n", colorBold, colorCyan, colorReset)
		for i, conn := range info.TCPConns {
			if i < 10 { // Limit display
				fmt.Printf("    %s•%s %s\n", colorGreen, colorReset, conn)
			}
		}
		if len(info.TCPConns) > 10 {
			fmt.Printf("    %s... and %d more%s\n", colorDim, len(info.TCPConns)-10, colorReset)
		}
	}

	if len(info.UDPConns) > 0 {
		fmt.Println()
		fmt.Printf("  %s%sUDP Connections:%s\n", colorBold, colorCyan, colorReset)
		for i, conn := range info.UDPConns {
			if i < 5 { // Limit display
				fmt.Printf("    %s•%s %s\n", colorGreen, colorReset, conn)
			}
		}
		if len(info.UDPConns) > 5 {
			fmt.Printf("    %s... and %d more%s\n", colorDim, len(info.UDPConns)-5, colorReset)
		}
	}

	fmt.Println()
	printDivider(width)
	fmt.Println()
}

func printBoxTop(width int) {
	fmt.Printf("%s%s╔%s╗%s\n", colorBold, colorCyan, strings.Repeat("═", width-2), colorReset)
}

func printBoxBottom(width int) {
	fmt.Printf("%s%s╚%s╝%s\n", colorBold, colorCyan, strings.Repeat("═", width-2), colorReset)
}

func printBoxLine(width int, text string, color string, center bool) {
	// Remove color codes from text for length calculation
	cleanText := stripAnsiCodes(text)
	textLen := len(cleanText)

	if center {
		padding := (width - 2 - textLen) / 2
		leftPad := padding
		rightPad := width - 2 - textLen - leftPad
		line := fmt.Sprintf("%s%s%s%s%s%s",
			strings.Repeat(" ", leftPad),
			colorBold, color, text, colorReset,
			strings.Repeat(" ", rightPad))
		fmt.Printf("%s%s║%s║%s\n", colorBold, colorCyan, line, colorReset)
	} else {
		rightPad := width - 2 - textLen
		line := fmt.Sprintf("%s%s%s%s%s",
			colorBold, color, text, colorReset,
			strings.Repeat(" ", rightPad-2))
		fmt.Printf("%s%s║ %s %s%s║%s\n", colorBold, colorCyan, line, colorBold, colorCyan, colorReset)
	}
}

func stripAnsiCodes(s string) string {
	// Simple ANSI code stripper
	result := ""
	inEscape := false
	for _, c := range s {
		if c == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if c == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(c)
	}
	return result
}

func printSectionHeader(title string, width int) {
	fmt.Printf("  %s%s%s%s\n", colorBold, colorCyan, title, colorReset)
	fmt.Printf("  %s%s%s\n", colorCyan, strings.Repeat("─", width-4), colorReset)
}

func printField(label string, value string, valueColor string) {
	fmt.Printf("  %s%-22s%s %s%s%s\n", colorYellow, label+":", colorReset, valueColor, value, colorReset)
}

func printDivider(width int) {
	fmt.Printf("%s%s%s%s\n", colorDim, strings.Repeat("─", width), colorReset, colorReset)
}

func getStateColor(state string) string {
	if strings.Contains(state, "Running") {
		return colorGreen
	} else if strings.Contains(state, "Sleeping") {
		return colorCyan
	} else if strings.Contains(state, "Zombie") {
		return colorRed
	} else if strings.Contains(state, "Stopped") {
		return colorYellow
	}
	return colorWhite
}

func formatMemory(kb int64) string {
	if kb < 1024 {
		return fmt.Sprintf("%d KB", kb)
	} else if kb < 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(kb)/1024)
	} else {
		return fmt.Sprintf("%.2f GB", float64(kb)/1024/1024)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func promptKill(info *ProcessInfo) bool {
	fmt.Printf("%s%s⚠️  Do you want to kill process %d (%s)?%s [y/N]: ",
		colorBold, colorYellow, info.ID, info.Command, colorReset)
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

	// Check if the process exists by sending signal 0
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return fmt.Errorf("process with PID %d does not exist or is not accessible: %w", pid, err)
	}

	// Attempt to send SIGTERM
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
