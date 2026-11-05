// Copyright (C) 2023-2026 Red Hat, Inc.
package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"

	ClearLineCode = "\r\033[2K\r"
)

// ASCII art generated on http://www.patorjk.com/software/taag/ with
// the font Standard by Glenn Chappell & Ian Chai 3/93.
const banner = `

   ____  _____  ____  _____  ____   _   _  ___  _____  _____
  / ___|| ____||  _ \|_   _|/ ___| | | | ||_ _||_   _|| ____|
 | |    |  _|  | |_) | | |  \___ \ | | | | | |   | |  |  _|
 | |___ | |___ |  _ <  | |   ___) || |_| | | |   | |  | |___
  \____||_____||_| \_\ |_|  |____/  \___/ |___|  |_|  |_____|



`

const (
	CheckResultTagPass    = Green + "PASS" + Reset
	CheckResultTagFail    = Red + "FAIL" + Reset
	CheckResultTagSkip    = Yellow + "SKIP" + Reset
	CheckResultTagRunning = Cyan + "RUNNING" + Reset
	CheckResultTagAborted = Red + "ABORTED" + Reset
	CheckResultTagError   = Red + "ERROR" + Reset

	tickerPeriodSeconds = 10
	lineLength          = 5
)

var CliCheckLogSniffer = &cliCheckLogSniffer{}

var (
	checkLoggerChan chan string
	stopChan        chan bool
)

// quietMode silences the console progress output when the suite runs with
// --quiet. The log file output is not affected.
var quietMode bool

// SetQuiet enables or disables quiet mode for the console output functions.
func SetQuiet(quiet bool) {
	quietMode = quiet
}

// IsQuiet reports whether quiet mode is enabled.
func IsQuiet() bool {
	return quietMode
}

func PrintBanner() {
	if quietMode {
		return
	}
	fmt.Print(banner)
}

type cliCheckLogSniffer struct{}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func updateRunningCheckLine(checkName string, stopChan <-chan bool) {
	startTime := time.Now()

	// Local string var to save the last received log line from the running check.
	lastCheckLogLine := ""

	tickerPeriod := 1 * time.Second
	if !isTTY() {
		// Increase it to avoid flooding the text output.
		tickerPeriod = tickerPeriodSeconds * time.Second
	}

	timer := time.NewTicker(tickerPeriod)
	for {
		select {
		case <-timer.C:
			printRunningCheckLine(checkName, startTime, lastCheckLogLine)
		case newLogLine := <-checkLoggerChan:
			lastCheckLogLine = newLogLine
			printRunningCheckLine(checkName, startTime, lastCheckLogLine)
		case <-stopChan:
			timer.Stop()
			return
		}
	}
}

func getTerminalWidth() int {
	width, _, _ := term.GetSize(int(os.Stdin.Fd()))
	return width
}

func cropLogLine(line string, maxAvailableWidth int) string {
	// Remove line feeds to avoid the log line to break the cli output.
	filteredLine := strings.ReplaceAll(line, "\n", " ")
	// Print only the chars that fit in the available space.
	if len(filteredLine) > maxAvailableWidth {
		return filteredLine[:maxAvailableWidth]
	}
	return filteredLine
}

func printRunningCheckLine(checkName string, startTime time.Time, logLine string) {
	if quietMode {
		return
	}

	// Minimum space on the right needed to show the current last log line.
	const minColsNeededForLogLine = 40

	elapsedTime := time.Since(startTime).Round(time.Second)
	line := "[ " + CheckResultTagRunning + " ] " + checkName + " (" + elapsedTime.String() + ")"
	if !isTTY() {
		fmt.Print(line + "\n")
		return
	}

	// Add check's last log line only if the program is running in a tty/ptty.
	maxAvailableWidth := getTerminalWidth() - len(line) - lineLength
	if logLine != "" && maxAvailableWidth > minColsNeededForLogLine {
		// Append a cropped log line only if it makes sense due to the available space on the right.
		line += "   " + cropLogLine(logLine, maxAvailableWidth)
	}

	fmt.Print(ClearLineCode + line)
}

// Implements the io.Write for the checks' custom handler for slog.
func (c *cliCheckLogSniffer) Write(p []byte) (n int, err error) {
	if !isTTY() {
		return len(p), nil
	}
	// Send to channel, or ignore it in case the channel is not ready or is closed.
	// This way we avoid blocking the whole program.
	select {
	case checkLoggerChan <- string(p):
	default:
	}

	return len(p), nil
}

func PrintResultsTable(results map[string][]int) {
	if quietMode {
		return
	}
	fmt.Printf("\n")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Printf("| %-27s %-9s %-9s %-9s %s |\n", "SUITE", "PASSED", "FAILED", "SKIPPED", "ERROR")
	fmt.Println("------------------------------------------------------------------------")
	for groupName, groupResults := range results {
		fmt.Printf("| %-25s %8d %9d %10d %7d |\n", groupName,
			groupResults[0],
			groupResults[1],
			groupResults[2],
			groupResults[3])
		fmt.Println("------------------------------------------------------------------------")
	}
	fmt.Printf("\n")
}

func stopCheckLineGoroutine() {
	if stopChan == nil {
		// This may happen for checks that were skipped if no compliant nor non-compliant objects found.
		return
	}

	stopChan <- true
	// Make this chnanel immediately unavailable.
	stopChan = nil
}

func PrintCheckSkipped(checkName, reason string) {
	// It shouldn't happen too often, but some checks might be set as skipped inside the checkFn
	// if neither compliant objects nor non-compliant objects were found.
	stopCheckLineGoroutine()

	if quietMode {
		return
	}
	fmt.Print(ClearLineCode + "[ " + CheckResultTagSkip + " ] " + checkName + "  (" + reason + ")\n")
}

func PrintCheckRunning(checkName string) {
	stopChan = make(chan bool)
	checkLoggerChan = make(chan string)

	// The goroutine below keeps running in quiet mode to drain checkLoggerChan.
	if !quietMode {
		line := "[ " + CheckResultTagRunning + " ] " + checkName
		if !isTTY() {
			line += "\n"
		}

		fmt.Print(line)
	}

	go updateRunningCheckLine(checkName, stopChan)
}

func PrintCheckPassed(checkName string) {
	stopCheckLineGoroutine()

	if quietMode {
		return
	}
	fmt.Print(ClearLineCode + "[ " + CheckResultTagPass + " ] " + checkName + "\n")
}

func PrintCheckFailed(checkName string) {
	stopCheckLineGoroutine()

	if quietMode {
		return
	}
	fmt.Print(ClearLineCode + "[ " + CheckResultTagFail + " ] " + checkName + "\n")
}

func PrintCheckAborted(checkName, reason string) {
	stopCheckLineGoroutine()

	if quietMode {
		return
	}
	fmt.Print(ClearLineCode + "[ " + CheckResultTagAborted + " ] " + checkName + "  (" + reason + ")\n")
}

func PrintCheckErrored(checkName string) {
	stopCheckLineGoroutine()

	if quietMode {
		return
	}
	fmt.Print(ClearLineCode + "[ " + CheckResultTagError + " ] " + checkName + "\n")
}

func WrapLines(text string, maxWidth int) []string {
	lines := strings.Split(text, "\n")
	wrappedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) <= maxWidth {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Break lines longer than maxWidth
		words := strings.Fields(line)
		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+len(word)+1 <= maxWidth {
				currentLine += " " + word
			} else {
				wrappedLines = append(wrappedLines, currentLine)
				currentLine = word
			}
		}

		wrappedLines = append(wrappedLines, currentLine)
	}

	return wrappedLines
}

func LineAlignLeft(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, s)
}

func LineAlignCenter(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, fmt.Sprintf("%[1]*s", (w+len(s))/2, s)) //nolint:mnd // magic number
}

func LineColor(s, color string) string {
	return color + s + Reset
}

// AddCommand adds a child command to a parent command if the child is not nil.
func AddCommand(parent, child *cobra.Command) {
	if child != nil {
		parent.AddCommand(child)
	}
}
