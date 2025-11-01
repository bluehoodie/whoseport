// Package model defines core data structures for process information.
package model

// ProcessInfo represents comprehensive information about a process listening on a port.
// It combines data from multiple sources: lsof output and the /proc filesystem.
type ProcessInfo struct {
	// Basic info from lsof (9 fields from lsof output)
	Command    string `json:"command"`      // Process command name
	ID         int    `json:"id"`           // Process ID (PID)
	User       string `json:"user"`         // Username running the process
	FD         string `json:"fd"`           // File descriptor
	Type       string `json:"type"`         // Network type (IPv4/IPv6)
	Device     string `json:"device"`       // Device number
	SizeOffset string `json:"size_offset"`  // Size/offset
	Node       string `json:"node"`         // Protocol (TCP/UDP)
	Name       string `json:"name"`         // Connection name (e.g., "*:8080 (LISTEN)")

	// Enhanced details from /proc filesystem (18 fields)
	FullCommand   string   `json:"full_command"`         // Full command line with arguments
	PPid          int      `json:"ppid"`                 // Parent process ID
	ParentCommand string   `json:"parent_command"`       // Parent process command
	State         string   `json:"state"`                // Process state (R, S, D, Z, T, etc.)
	Threads       int      `json:"threads"`              // Number of threads
	WorkingDir    string   `json:"working_dir"`          // Current working directory
	MemoryRSS     int64    `json:"memory_rss_kb"`        // Resident set size in KB
	MemoryVMS     int64    `json:"memory_vms_kb"`        // Virtual memory size in KB
	CPUTime       float64  `json:"cpu_time_seconds"`     // Total CPU time in seconds
	StartTime     string   `json:"start_time"`           // Process start time
	Uptime        string   `json:"uptime"`               // Process uptime duration
	OpenFDs       int      `json:"open_fds"`             // Number of open file descriptors
	MaxFDs        int      `json:"max_fds"`              // Maximum file descriptors
	UID           int      `json:"uid"`                  // User ID
	GID           int      `json:"gid"`                  // Group ID
	Groups        string   `json:"groups"`               // Supplementary groups
	NetworkConns  int      `json:"network_connections"`  // Number of network connections
	TCPConns      []string `json:"tcp_connections"`      // TCP connection details
	UDPConns      []string `json:"udp_connections"`      // UDP connection details

	// Additional enhanced info (11 fields)
	ExePath         string  `json:"exe_path"`           // Path to executable
	ExeSize         int64   `json:"exe_size_bytes"`     // Executable file size in bytes
	NiceValue       int     `json:"nice_value"`         // Process nice value
	Priority        int     `json:"priority"`           // Process priority
	EnvCount        int     `json:"env_count"`          // Number of environment variables
	ChildCount      int     `json:"child_count"`        // Number of child processes
	IOReadBytes     int64   `json:"io_read_bytes"`      // Bytes read from storage
	IOWriteBytes    int64   `json:"io_write_bytes"`     // Bytes written to storage
	IOReadSyscalls  int64   `json:"io_read_syscalls"`   // Number of read syscalls
	IOWriteSyscalls int64   `json:"io_write_syscalls"`  // Number of write syscalls
	MemoryLimit     int64   `json:"memory_limit_kb"`    // Memory limit in KB (-1 if unlimited)
	CPUPercent      float64 `json:"cpu_percent"`        // CPU usage percentage
}

// New creates a new ProcessInfo with basic lsof data.
// Enhanced fields are populated separately via the enhancer package.
func New(command string, id int, user, fd, typ, device, sizeOffset, node, name string) *ProcessInfo {
	return &ProcessInfo{
		Command:    command,
		ID:         id,
		User:       user,
		FD:         fd,
		Type:       typ,
		Device:     device,
		SizeOffset: sizeOffset,
		Node:       node,
		Name:       name,
	}
}
