package terminal

// Color codes for terminal output (256-color palette)
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorItalic = "\033[3m"

	// Basic colors
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"

	// Bright colors
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	// 256-color palette for vibrant UI
	ColorOrange   = "\033[38;5;208m"
	ColorDeepPink = "\033[38;5;198m"
	ColorViolet   = "\033[38;5;141m"
	ColorSkyBlue  = "\033[38;5;117m"
	ColorLime     = "\033[38;5;154m"
	ColorGold     = "\033[38;5;220m"
	ColorCoral    = "\033[38;5;210m"
	ColorTeal     = "\033[38;5;80m"
	ColorLavender = "\033[38;5;183m"
	ColorMint     = "\033[38;5;121m"
	ColorPeach    = "\033[38;5;216m"
	ColorIndigo   = "\033[38;5;63m"

	// Background colors
	BgBlue   = "\033[48;5;24m"
	BgPurple = "\033[48;5;54m"
	BgGreen  = "\033[48;5;22m"
	BgOrange = "\033[48;5;130m"
)

// Theme defines the color scheme for the UI
type Theme struct {
	Primary       string
	Label         string
	StateRunning  string
	StateSleeping string
	StateZombie   string
	StateStopped  string
	StateDefault  string
}

// DefaultTheme returns the default color theme
func DefaultTheme() Theme {
	return Theme{
		Primary:       ColorCyan,
		Label:         ColorYellow,
		StateRunning:  ColorGreen,
		StateSleeping: ColorCyan,
		StateZombie:   ColorRed,
		StateStopped:  ColorYellow,
		StateDefault:  ColorWhite,
	}
}
