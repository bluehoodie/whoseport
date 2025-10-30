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

// Color codes for terminal output (256-color palette for stunning visuals)
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorItalic = "\033[3m"

	// Basic colors
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"

	// Bright colors
	colorBrightRed     = "\033[91m"
	colorBrightGreen   = "\033[92m"
	colorBrightYellow  = "\033[93m"
	colorBrightBlue    = "\033[94m"
	colorBrightMagenta = "\033[95m"
	colorBrightCyan    = "\033[96m"
	colorBrightWhite   = "\033[97m"

	// 256-color palette for vibrant UI
	colorOrange   = "\033[38;5;208m"
	colorDeepPink = "\033[38;5;198m"
	colorViolet   = "\033[38;5;141m"
	colorSkyBlue  = "\033[38;5;117m"
	colorLime     = "\033[38;5;154m"
	colorGold     = "\033[38;5;220m"
	colorCoral    = "\033[38;5;210m"
	colorTeal     = "\033[38;5;80m"
	colorLavender = "\033[38;5;183m"
	colorMint     = "\033[38;5;121m"
	colorPeach    = "\033[38;5;216m"
	colorIndigo   = "\033[38;5;63m"

	// Background colors for boxes
	bgBlue   = "\033[48;5;24m"
	bgPurple = "\033[48;5;54m"
	bgGreen  = "\033[48;5;22m"
	bgOrange = "\033[48;5;130m"
)

var (
	killFlag      bool
	noInteractive bool
	jsonFlag      bool
)

// Subtle theme to avoid rainbow effect
const (
	themePrimary = colorCyan
)

func testAlignment() {
	width := 80
	text := "🔍 PROCESS DETAILS FOR PORT 8889"

	fmt.Println("Testing box alignment:")
	fmt.Println()

	// Top border (no colors for clarity in test)
	topLine := fmt.Sprintf("╔%s╗", strings.Repeat("═", width-2))
	fmt.Println(topLine)

	// Header line - using same logic as printBoxLine
	textLen := visualWidth(text)
	interiorWidth := width - 2
	padding := (interiorWidth - textLen) / 2
	leftPad := padding
	rightPad := interiorWidth - textLen - leftPad

	// Build header line for testing (without ANSI codes for this part)
	// Build it exactly as the real function would
	headerLine := "║" + strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad) + "║"

	// This mimics what printBoxLine does with colors (for visual output)
	fmt.Printf("%s%s║%s%s%s%s%s%s║%s\n",
		colorBold, colorCyan,
		strings.Repeat(" ", leftPad),
		colorBold, colorCyan, text, colorReset,
		strings.Repeat(" ", rightPad),
		colorReset)

	// Bottom border (no colors for clarity in test)
	bottomLine := fmt.Sprintf("╚%s╝", strings.Repeat("═", width-2))
	fmt.Println(bottomLine)

	fmt.Println()
	fmt.Println("=== ALIGNMENT CHECK ===")

	// Strip colors for analysis
	topClean := stripAnsiCodes(topLine)
	bottomClean := stripAnsiCodes(bottomLine)
	headerClean := stripAnsiCodes(headerLine)

	fmt.Printf("Top line:    '%s'\n", topClean)
	fmt.Printf("Header line: '%s'\n", headerClean)
	fmt.Printf("Bottom line: '%s'\n", bottomClean)
	fmt.Println()

	fmt.Printf("Top length:    %d\n", len(topClean))
	fmt.Printf("Header length: %d\n", len(headerClean))
	fmt.Printf("Bottom length: %d\n", len(bottomClean))
	fmt.Println()

	topRightPos := strings.LastIndex(topClean, "╗")
	headerRightPos := strings.LastIndex(headerClean, "║")
	bottomRightPos := strings.LastIndex(bottomClean, "╝")

	fmt.Printf("Top ╗ at position:     %d\n", topRightPos)
	fmt.Printf("Header ║ at position:  %d\n", headerRightPos)
	fmt.Printf("Bottom ╝ at position:  %d\n", bottomRightPos)
	fmt.Println()

	if topRightPos == headerRightPos && headerRightPos == bottomRightPos {
		fmt.Println("✓ PERFECTLY ALIGNED!")
	} else {
		fmt.Printf("✗ MISALIGNED:\n")
		fmt.Printf("  Header offset from top: %+d\n", headerRightPos-topRightPos)
		fmt.Printf("  Bottom offset from top: %+d\n", bottomRightPos-topRightPos)
	}
}

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
	FullCommand   string   `json:"full_command"`
	PPid          int      `json:"ppid"`
	ParentCommand string   `json:"parent_command"`
	State         string   `json:"state"`
	Threads       int      `json:"threads"`
	WorkingDir    string   `json:"working_dir"`
	MemoryRSS     int64    `json:"memory_rss_kb"`
	MemoryVMS     int64    `json:"memory_vms_kb"`
	CPUTime       float64  `json:"cpu_time_seconds"`
	StartTime     string   `json:"start_time"`
	Uptime        string   `json:"uptime"`
	OpenFDs       int      `json:"open_fds"`
	MaxFDs        int      `json:"max_fds"`
	UID           int      `json:"uid"`
	GID           int      `json:"gid"`
	Groups        string   `json:"groups"`
	NetworkConns  int      `json:"network_connections"`
	TCPConns      []string `json:"tcp_connections"`
	UDPConns      []string `json:"udp_connections"`

	// Additional enhanced info
	ExePath         string  `json:"exe_path"`
	ExeSize         int64   `json:"exe_size_bytes"`
	NiceValue       int     `json:"nice_value"`
	Priority        int     `json:"priority"`
	EnvCount        int     `json:"env_count"`
	ChildCount      int     `json:"child_count"`
	IOReadBytes     int64   `json:"io_read_bytes"`
	IOWriteBytes    int64   `json:"io_write_bytes"`
	IOReadSyscalls  int64   `json:"io_read_syscalls"`
	IOWriteSyscalls int64   `json:"io_write_syscalls"`
	MemoryLimit     int64   `json:"memory_limit_kb"`
	CPUPercent      float64 `json:"cpu_percent"`
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

	// Get executable path and size
	if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
		info.ExePath = exe
		if stat, err := os.Stat(exe); err == nil {
			info.ExeSize = stat.Size()
		}
	}

	// Count open file descriptors
	if fds, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid)); err == nil {
		info.OpenFDs = len(fds)
	}

	// Count environment variables
	if environ, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", pid)); err == nil {
		envVars := bytes.Split(environ, []byte{0})
		info.EnvCount = len(envVars) - 1 // Subtract 1 for trailing null
		if info.EnvCount < 0 {
			info.EnvCount = 0
		}
	}

	// Count child processes
	if tasks, err := os.ReadDir("/proc"); err == nil {
		for _, task := range tasks {
			if !task.IsDir() {
				continue
			}
			childPid, err := strconv.Atoi(task.Name())
			if err != nil {
				continue
			}
			// Read the child's status to check PPid
			if status, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", childPid)); err == nil {
				lines := strings.Split(string(status), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "PPid:") {
						fields := strings.Fields(line)
						if len(fields) >= 2 {
							ppid, _ := strconv.Atoi(fields[1])
							if ppid == pid {
								info.ChildCount++
							}
						}
						break
					}
				}
			}
		}
	}

	// Get IO statistics
	if io, err := os.ReadFile(fmt.Sprintf("/proc/%d/io", pid)); err == nil {
		parseIO(string(io), info)
	}

	// Get limits
	if limits, err := os.ReadFile(fmt.Sprintf("/proc/%d/limits", pid)); err == nil {
		parseLimits(string(limits), info)
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

	// Priority and nice value
	if len(fields) > 17 {
		info.Priority, _ = strconv.Atoi(fields[17])
	}
	if len(fields) > 18 {
		info.NiceValue, _ = strconv.Atoi(fields[18])
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

	// Calculate uptime and CPU percentage
	uptime := time.Since(startTimeObj)
	info.Uptime = formatDuration(uptime)

	// Calculate CPU percentage (CPU time / uptime * 100)
	if uptime.Seconds() > 0 {
		info.CPUPercent = (info.CPUTime / uptime.Seconds()) * 100.0
	}
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

func parseIO(io string, info *ProcessInfo) {
	lines := strings.Split(io, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, _ := strconv.ParseInt(fields[1], 10, 64)

		switch key {
		case "read_bytes":
			info.IOReadBytes = value
		case "write_bytes":
			info.IOWriteBytes = value
		case "syscr":
			info.IOReadSyscalls = value
		case "syscw":
			info.IOWriteSyscalls = value
		}
	}
}

func parseLimits(limits string, info *ProcessInfo) {
	lines := strings.Split(limits, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Max address space") || strings.Contains(line, "Max data size") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				limit := fields[3]
				if limit != "unlimited" {
					limitVal, _ := strconv.ParseInt(limit, 10, 64)
					if info.MemoryLimit == 0 || limitVal < info.MemoryLimit {
						info.MemoryLimit = limitVal / 1024 // Convert to KB
					}
				}
			}
		}
	}
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
	width := 90

	// Subtle header banner
	printGradientBanner(width, fmt.Sprintf("PORT %d ANALYSIS", port))

	// Section 1: Process Identity with modern box design
	printModernSection("⚙️  PROCESS IDENTITY", themePrimary, width)
	printEnhancedField("Command", info.Command, colorBrightGreen, "")
	printEnhancedField("Full Command", truncate(info.FullCommand, 60), colorLime, "")
	printEnhancedField("Process ID", fmt.Sprintf("%d", info.ID), colorOrange, "PID")
	printEnhancedField("Parent PID", fmt.Sprintf("%d", info.PPid), colorPeach, "")
	if info.ParentCommand != "" {
		printEnhancedField("Parent Process", info.ParentCommand, colorLavender, "")
	}
	printEnhancedField("User", info.User, colorGold, "")
	printEnhancedField("UID / GID", fmt.Sprintf("%d / %d", info.UID, info.GID), colorDim, "")
	if info.ChildCount > 0 {
		printEnhancedField("Child Processes", fmt.Sprintf("%d", info.ChildCount), colorMint, "")
	}

	// Section 2: Binary Information
	if info.ExePath != "" {
		printModernSection("📦 BINARY INFORMATION", themePrimary, width)
		printEnhancedField("Executable Path", info.ExePath, colorCyan, "")
		if info.ExeSize > 0 {
			printEnhancedField("Binary Size", formatBytes(info.ExeSize), colorTeal, "")
		}
		printEnhancedField("Environment Vars", fmt.Sprintf("%d", info.EnvCount), colorLavender, "")
		if info.WorkingDir != "" {
			printEnhancedField("Working Directory", info.WorkingDir, colorSkyBlue, "")
		}
	}

	// Section 3: Process State with status indicator
	printModernSection("📊 PROCESS STATE", themePrimary, width)
	stateEmoji := getStateEmoji(info.State)
	printEnhancedField("State", fmt.Sprintf("%s %s", stateEmoji, info.State), getStateColor(info.State), "")
	printEnhancedField("Threads", fmt.Sprintf("%d", info.Threads), colorViolet, "")
	if info.NiceValue != 0 || info.Priority != 0 {
		printEnhancedField("Nice / Priority", fmt.Sprintf("%d / %d", info.NiceValue, info.Priority), colorPeach, "")
	}
	if info.StartTime != "" {
		printEnhancedField("Started", info.StartTime, colorCyan, "")
	}
	if info.Uptime != "" {
		printEnhancedField("Uptime", info.Uptime, colorLime, "⏱")
	}
	if info.CPUTime > 0 {
		cpuStr := fmt.Sprintf("%.2fs", info.CPUTime)
		if info.CPUPercent > 0 {
			cpuStr += fmt.Sprintf(" (%.2f%%)", info.CPUPercent)
		}
		printEnhancedField("CPU Time", cpuStr, colorOrange, "")
	}

	// Section 4: Memory Usage with visual bars
	printModernSection("💾 MEMORY USAGE", themePrimary, width)
	if info.MemoryRSS > 0 {
		rssBar := createMemoryBar(info.MemoryRSS, info.MemoryVMS, 30)
		printEnhancedField("Resident Set (RSS)", formatMemory(info.MemoryRSS), colorLime, "")
		fmt.Printf("  %s%s%s\n", colorLime, rssBar, colorReset)
	}
	if info.MemoryVMS > 0 {
		vmsBar := createMemoryBar(info.MemoryVMS, info.MemoryVMS*2, 30)
		printEnhancedField("Virtual Memory", formatMemory(info.MemoryVMS), colorSkyBlue, "")
		fmt.Printf("  %s%s%s\n", colorSkyBlue, vmsBar, colorReset)
	}
	if info.MemoryLimit > 0 {
		printEnhancedField("Memory Limit", formatMemory(info.MemoryLimit), colorDim, "")
	}

	// Section 5: I/O Statistics
	if info.IOReadBytes > 0 || info.IOWriteBytes > 0 {
		printModernSection("💿 I/O STATISTICS", themePrimary, width)
		if info.IOReadBytes > 0 {
			printEnhancedField("Read", formatBytes(info.IOReadBytes), colorBrightCyan, "📖")
			if info.IOReadSyscalls > 0 {
				printEnhancedField("  Read Syscalls", fmt.Sprintf("%d", info.IOReadSyscalls), colorDim, "")
			}
		}
		if info.IOWriteBytes > 0 {
			printEnhancedField("Write", formatBytes(info.IOWriteBytes), colorCoral, "📝")
			if info.IOWriteSyscalls > 0 {
				printEnhancedField("  Write Syscalls", fmt.Sprintf("%d", info.IOWriteSyscalls), colorDim, "")
			}
		}
	}

	// Section 6: File Descriptors
	printModernSection("📁 FILE DESCRIPTORS", themePrimary, width)
	if info.MaxFDs > 0 {
		fdPercent := float64(info.OpenFDs) / float64(info.MaxFDs) * 100
		fdBar := createProgressBar(fdPercent, 30)
		printEnhancedField("Open FDs", fmt.Sprintf("%d / %d (%.1f%%)", info.OpenFDs, info.MaxFDs, fdPercent), colorYellow, "")
		fmt.Printf("  %s\n", fdBar)
	} else {
		printEnhancedField("Open FDs", fmt.Sprintf("%d / N/A", info.OpenFDs), colorYellow, "")
		fmt.Printf("  %s\n", "N/A")
	}

	// Section 7: Network Information with enhanced visuals
	printModernSection("🌐 NETWORK", themePrimary, width)
	printEnhancedField("Protocol", strings.ToUpper(info.Type), colorBrightCyan, "")
	printEnhancedField("Listening On", info.Name, colorBrightGreen, "🎧")
	printEnhancedField("Node Type", info.Node, colorLavender, "")
	printEnhancedField("File Descriptor", info.FD, colorDim, "")
	printEnhancedField("Total Connections", fmt.Sprintf("%d", info.NetworkConns), colorOrange, "")

	if len(info.TCPConns) > 0 {
		fmt.Printf("  %s%s%s▸ TCP Connections%s\n", colorBold, colorBrightCyan, colorReset, colorReset)
		for i, conn := range info.TCPConns {
			if i < 8 {
				fmt.Printf("    %s┃%s %s%s%s\n", colorBrightBlue, colorReset, colorMint, conn, colorReset)
			}
		}
		if len(info.TCPConns) > 8 {
			fmt.Printf("    %s┗━ and %d more...%s\n", colorDim, len(info.TCPConns)-8, colorReset)
		}
	}

	if len(info.UDPConns) > 0 {
		fmt.Printf("  %s%s%s▸ UDP Connections%s\n", colorBold, colorBrightCyan, colorReset, colorReset)
		for i, conn := range info.UDPConns {
			if i < 5 {
				fmt.Printf("    %s┃%s %s%s%s\n", colorBrightBlue, colorReset, colorLavender, conn, colorReset)
			}
		}
		if len(info.UDPConns) > 5 {
			fmt.Printf("    %s┗━ and %d more...%s\n", colorDim, len(info.UDPConns)-5, colorReset)
		}
	}
	printGradientDivider(width)
}

// Modern UI helper functions

func printGradientBanner(width int, text string) {
	textLen := visualWidth(text)
	emojiWidth := 2 // 🔍 is typically 2 cells wide
	spaceWidth := 1 // one space after emoji
	totalTextLen := textLen + emojiWidth + spaceWidth

	// Inner width is the space between the two vertical borders
	innerWidth := width - 2
	if innerWidth < totalTextLen {
		innerWidth = totalTextLen
	}
	leftPadding := (innerWidth - totalTextLen) / 2
	rightPadding := innerWidth - totalTextLen - leftPadding

	// Top border
	fmt.Printf("%s%s╔", colorBold, themePrimary)
	for i := 0; i < innerWidth; i++ {
		fmt.Printf("═")
	}
	fmt.Printf("╗%s\n", colorReset)

	// Middle line (keep color active until after closing border)
	fmt.Printf("%s%s║%s%s🔍 %s%s%s%s║%s\n",
		colorBold, themePrimary,
		strings.Repeat(" ", leftPadding),
		colorBold, colorBrightCyan, text, colorReset,
		strings.Repeat(" ", rightPadding),
		colorReset)

	// Bottom border
	fmt.Printf("%s%s╚", colorBold, themePrimary)
	for i := 0; i < innerWidth; i++ {
		fmt.Printf("═")
	}
	fmt.Printf("╝%s\n", colorReset)
}

func printModernSection(title string, color string, width int) {
	// Section header with consistent color to the closing border
	titleLen := visualWidth(title)
	fmt.Printf("\n  %s%s┏━%s━┓%s\n", colorBold, color, strings.Repeat("━", titleLen), colorReset)
	fmt.Printf("  %s%s┃ %s ┃%s\n", colorBold, color, title, colorReset)
	fmt.Printf("  %s%s┗━%s━┛%s\n", colorBold, color, strings.Repeat("━", titleLen), colorReset)
}

func printEnhancedField(label string, value string, valueColor string, emoji string) {
	labelColor := colorYellow
	if emoji != "" {
		fmt.Printf("  %s%s%-20s%s %s%s %s%s%s\n",
			colorBold, labelColor, label+":", colorReset,
			emoji, valueColor, value, colorReset, colorReset)
	} else {
		fmt.Printf("  %s%s%-20s%s %s%s%s\n",
			colorBold, labelColor, label+":", colorReset,
			valueColor, value, colorReset)
	}
}

func createProgressBar(percent float64, width int) string {
	filled := int(float64(width) * percent / 100.0)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	empty := width - filled

	// Color based on percentage
	barColor := colorLime
	if percent > 75 {
		barColor = colorOrange
	}
	if percent > 90 {
		barColor = colorRed
	}

	bar := fmt.Sprintf("%s[%s%s%s]%s",
		colorDim,
		barColor, strings.Repeat("█", filled)+strings.Repeat("░", empty),
		colorDim,
		colorReset)

	return bar
}

func createMemoryBar(used int64, total int64, width int) string {
	if total == 0 {
		total = used
	}
	percent := float64(used) / float64(total) * 100.0
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * percent / 100.0)
	if filled > width {
		filled = width
	}

	empty := width - filled

	bar := fmt.Sprintf("[%s%s%s]",
		strings.Repeat("▓", filled),
		strings.Repeat("░", empty),
		colorReset)

	return bar
}

func printGradientDivider(width int) {
	// Use a single subtle color instead of a rainbow gradient
	fmt.Printf("%s%s%s%s\n", colorBold, themePrimary, strings.Repeat("─", width), colorReset)
}

func getStateEmoji(state string) string {
	if strings.Contains(state, "Running") {
		return "🟢"
	} else if strings.Contains(state, "Sleeping") {
		return "🔵"
	} else if strings.Contains(state, "Zombie") {
		return "💀"
	} else if strings.Contains(state, "Stopped") {
		return "🟡"
	} else if strings.Contains(state, "Waiting") {
		return "⏸"
	}
	return "⚪"
}

func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
	}
}

func printBoxTop(width int) {
	fmt.Printf("%s%s╔%s╗%s\n", colorBold, colorCyan, strings.Repeat("═", width-2), colorReset)
}

func printBoxBottom(width int) {
	fmt.Printf("%s%s╚%s╝%s\n", colorBold, colorCyan, strings.Repeat("═", width-2), colorReset)
}

func printBoxLine(width int, text string, color string, center bool) {
	// Calculate visual width of text (accounts for emojis and wide characters)
	textLen := visualWidth(text)
	// The interior width is: total width - 2 (for the two border characters ║)
	interiorWidth := width - 2

	if center {
		// Calculate padding for centering
		padding := (interiorWidth - textLen) / 2
		leftPad := padding
		rightPad := interiorWidth - textLen - leftPad

		// Build the interior content with exact visible character count
		fmt.Printf("%s%s║%s%s%s%s%s%s║%s\n",
			colorBold, colorCyan,
			strings.Repeat(" ", leftPad),
			colorBold, color, text, colorReset,
			strings.Repeat(" ", rightPad),
			colorReset)
	} else {
		rightPad := interiorWidth - textLen
		fmt.Printf("%s%s║ %s%s%s%s%s ║%s\n",
			colorBold, colorCyan,
			colorBold, color, text, colorReset,
			strings.Repeat(" ", rightPad-2),
			colorCyan, colorReset)
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

func visualWidth(s string) int {
	// Calculate the visual width of a string accounting for emojis and wide characters
	s = stripAnsiCodes(s) // Remove ANSI codes first
	width := 0
	for _, r := range s {
		// Emoji and CJK characters are typically 2 cells wide in terminals
		if r >= 0x1F300 && r <= 0x1F9FF {
			// Emoji range
			width += 2
		} else if r >= 0x2E80 && r <= 0xA4CF {
			// CJK range
			width += 2
		} else if r >= 0x3000 && r <= 0x303F {
			// CJK symbols range
			width += 2
		} else if r < 32 || r == 127 {
			// Control characters don't display
			width += 0
		} else {
			// Regular ASCII and most Unicode
			width += 1
		}
	}
	return width
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
