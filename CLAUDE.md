# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**whoseport** is a lightweight CLI utility written in Go that identifies which process is using a specific network port on Unix systems. It provides both interactive (colorized UI) and JSON output modes, with the ability to kill processes directly.

- **Language**: Go 1.24.4
- **Module**: github.com/bluehoodie/whoseport
- **Type**: Standalone single-binary CLI application
- **Architecture**: Monolithic, procedural design (~871 lines in a single main.go file)
- **Dependencies**: Zero external dependencies (stdlib only)
- **Platform**: Unix/Linux systems with `lsof` and `/proc` filesystem

## Common Commands

```bash
# Build the binary locally
make build

# Install to $GOBIN (or $GOPATH/bin or $HOME/go/bin)
make install

# Run tests
make test

# Clean build artifacts
make clean

# Local testing after build
./whoseport 8080
```

## Code Architecture

All code is in a single `main.go` file organized into functional layers:

### 1. **Data Structures** (lines 179-211)
- `ProcessInfo`: Rich struct with JSON serialization holding process details
- Fields: Command, PID, User, FD, Type, Device, SizeOffset, Node, Name, and enhanced fields (State, VmRSS, Uptime, etc.)

### 2. **Core Processing Flow**
- **Retrieval** (`lsof()`, `toProcessInfo()`): Parse lsof output into ProcessInfo
- **Enhancement** (`enhanceProcessInfo()`): Augment with /proc filesystem data
  - `parseStatus()`: Reads `/proc/[pid]/status` for state, memory usage
  - `parseStat()`: Reads `/proc/[pid]/stat` for uptime, CPU info
  - `getNetworkConnections()`: Reads `/proc/net/tcp` and `/proc/net/udp`
  - `getBootTime()`: Reads `/proc/stat` for system boot time
  - Helper parsers: `expandState()`, `parseAddress()`, `hexToIP()`, `parseTCPState()`
- **Action** (`killProcess()`, `promptKill()`): Process termination
- **Display** (`printJSON()`, `printInteractive()`): Output formatting

### 3. **Display System**
- Interactive mode: Box-drawn UI with ANSI colors and emoji (🔍⚙️📊💾)
- JSON mode: Structured output for scripting
- Box drawing utilities: `printBoxTop()`, `printBoxLine()`, `printBoxBottom()`
- Field formatting: `printSectionHeader()`, `printField()`, `printDivider()`
- Special handling: `visualWidth()` for emoji/wide characters, `stripAnsiCodes()` for ANSI cleanup

### 4. **CLI Interface**
- Entry point: `main()` function (line 108)
- Flag parsing: Uses Go standard `flag` package
- Flags: `-k/--kill`, `-i/--interactive`, `-n/--no-interactive`, `--json`
- Required argument: Port number (integer)

## Key Implementation Details

### External Dependencies
- **`lsof` command**: Used via subprocess to find processes listening on ports
- **`/proc` filesystem**: Linux-specific, provides process and network state
- Graceful fallback when /proc data unavailable

### Data Flow
```
Port Number
    ↓
lsof + grep (LISTEN filter)
    ↓
Parse lsof output → ProcessInfo struct
    ↓
Enhance with /proc filesystem data
    ↓
Display (JSON or Interactive UI)
    ↓
Optional: Kill process with SIGTERM
```

### Color & Terminal Output
- 15 predefined ANSI color constants
- UTF-8 box-drawing characters for UI (╔╗╚╝═║─)
- Emoji support with proper width calculation for alignment
- Handles terminal width constraints gracefully

## Testing

Currently configured but not implemented:
- Run with: `make test` or `go test ./...`
- No test files (`*_test.go`) exist in the repository

## Code Style Notes

- Procedural, functional decomposition (20+ helper functions)
- No OOP patterns or interfaces
- Minimal error handling (mostly direct error returns and os.Exit)
- Direct system interactions (file reads, subprocess execution, signal sending)
- Global functions organized by concern (retrieval, enhancement, display, actions)

## Development Workflow

1. **Make changes** to `main.go`
2. **Test locally**: `make build && ./whoseport [port]`
3. **Verify all modes**: Test with `--json`, interactive, and `-k` flags
4. **Consider cross-platform**: Code runs on Unix systems only; test assumptions about `/proc` and `lsof`
5. **Test edge cases**: Non-existent ports, processes without /proc data, permission issues
