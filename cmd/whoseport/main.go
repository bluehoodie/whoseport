package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/bluehoodie/whoseport/internal/action"
	"github.com/bluehoodie/whoseport/internal/display/interactive"
	displayjson "github.com/bluehoodie/whoseport/internal/display/json"
	"github.com/bluehoodie/whoseport/internal/process"
	"github.com/bluehoodie/whoseport/internal/procfs"
	"github.com/bluehoodie/whoseport/internal/terminal"
)

var (
	killFlag      bool
	noInteractive bool
	jsonFlag      bool
)

func main() {
	flag.BoolVar(&killFlag, "kill", false, "Kill the process using the port")
	flag.BoolVar(&killFlag, "k", false, "Kill the process using the port (shorthand)")
	flag.BoolVar(&noInteractive, "no-interactive", false, "Disable interactive mode (show info only)")
	flag.BoolVar(&noInteractive, "n", false, "Disable interactive mode (shorthand)")
	flag.BoolVar(&jsonFlag, "json", false, "Output in JSON format")

	flag.Usage = func() {
		fmt.Printf("%s%sUsage of whoseport:%s\n", terminal.ColorBold, terminal.ColorCyan, terminal.ColorReset)
		fmt.Printf("  %swhoseport [options] {port}%s\n\n", terminal.ColorWhite, terminal.ColorReset)
		fmt.Printf("%sOptions:%s\n", terminal.ColorBold, terminal.ColorReset)
		fmt.Printf("  %s-k, --kill%s            Kill the process using the port\n", terminal.ColorYellow, terminal.ColorReset)
		fmt.Printf("  %s-n, --no-interactive%s  Disable interactive mode (show info only)\n", terminal.ColorYellow, terminal.ColorReset)
		fmt.Printf("  %s--json%s                Output in JSON format\n", terminal.ColorYellow, terminal.ColorReset)
		fmt.Printf("\n%sNote:%s Interactive mode is enabled by default\n", terminal.ColorBold, terminal.ColorReset)
		fmt.Printf("\n%sExample:%s\n", terminal.ColorBold, terminal.ColorReset)
		fmt.Printf("  whoseport 8080           # Show detailed info with interactive prompt\n")
		fmt.Printf("  whoseport -n 8080        # Show info only, no prompt\n")
		fmt.Printf("  whoseport -k 8080        # Kill without prompting\n")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Printf("%serror:%s missing port number\n", terminal.ColorRed, terminal.ColorReset)
		flag.Usage()
		os.Exit(1)
	}

	port, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Printf("%serror:%s port must be an integer\n", terminal.ColorRed, terminal.ColorReset)
		flag.Usage()
		os.Exit(1)
	}

	// Retrieve process information
	retriever := process.NewDefaultRetriever()
	processInfo, err := retriever.GetProcessByPort(port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", terminal.ColorRed, terminal.ColorReset, err)
		os.Exit(1)
	}

	// Enhance with detailed process information
	enhancer := procfs.NewProcessEnhancer()
	enhancer.Enhance(processInfo)

	// Display the process info
	if jsonFlag {
		displayer := displayjson.NewDisplayer()
		if err := displayer.Display(processInfo); err != nil {
			fmt.Fprintf(os.Stderr, "%serror:%s %v\n", terminal.ColorRed, terminal.ColorReset, err)
			os.Exit(1)
		}
	} else {
		displayer := interactive.NewDisplayer()
		displayer.Display(processInfo, port)
	}

	// Handle kill logic
	killer := action.NewKiller()
	if killFlag {
		// Direct kill without prompting (backward compatible: uses SIGTERM by default)
		if err := killer.Kill(processInfo.ID, syscall.SIGTERM); err != nil {
			fmt.Printf("%s✗ Failed to kill process:%s %v\n", terminal.ColorRed, terminal.ColorReset, err)
			os.Exit(1)
		}
		fmt.Printf("%s✓ Successfully killed process %d with SIGTERM%s\n", terminal.ColorGreen, processInfo.ID, terminal.ColorReset)
	} else if !noInteractive {
		// Interactive mode - prompt for action in a single step
		prompter := action.NewPrompter()
		signal, shouldKill := prompter.PromptKillAction(processInfo)
		if shouldKill {
			// Send the selected signal
			if err := killer.Kill(processInfo.ID, signal); err != nil {
				fmt.Printf("%s✗ Failed to kill process:%s %v\n", terminal.ColorRed, terminal.ColorReset, err)
				os.Exit(1)
			}
			fmt.Printf("%s✓ Successfully sent signal %v to process %d%s\n", terminal.ColorGreen, signal, processInfo.ID, terminal.ColorReset)
		}
	}
}
