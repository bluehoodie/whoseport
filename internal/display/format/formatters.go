package format

import (
	"fmt"
	"strings"
)

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
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

// FormatMemory converts KB to human-readable format
func FormatMemory(kb int64) string {
	if kb < 1024 {
		return fmt.Sprintf("%d KB", kb)
	} else if kb < 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(kb)/1024)
	} else {
		return fmt.Sprintf("%.2f GB", float64(kb)/1024/1024)
	}
}

// Truncate shortens a string to maxLen with ellipsis
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// VisualWidth calculates the visual width of a string accounting for emojis
func VisualWidth(s string) int {
	s = StripAnsiCodes(s)
	width := 0
	for _, r := range s {
		if r >= 0x1F300 && r <= 0x1F9FF {
			width += 2 // Emoji range
		} else if r >= 0x2E80 && r <= 0xA4CF {
			width += 2 // CJK range
		} else if r >= 0x3000 && r <= 0x303F {
			width += 2 // CJK symbols range
		} else if r < 32 || r == 127 {
			width += 0 // Control characters
		} else {
			width += 1 // Regular ASCII and most Unicode
		}
	}
	return width
}

// StripAnsiCodes removes ANSI escape sequences from a string
func StripAnsiCodes(s string) string {
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

// GetStateEmoji returns an emoji representing the process state
func GetStateEmoji(state string) string {
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

// GetStateColor returns the color for a process state
func GetStateColor(state string, theme interface{}) string {
	// Accept theme as interface{} to avoid circular dependency
	// Caller should pass theme colors directly
	if strings.Contains(state, "Running") {
		return "\033[32m" // Green
	} else if strings.Contains(state, "Sleeping") {
		return "\033[36m" // Cyan
	} else if strings.Contains(state, "Zombie") {
		return "\033[31m" // Red
	} else if strings.Contains(state, "Stopped") {
		return "\033[33m" // Yellow
	}
	return "\033[37m" // White
}
