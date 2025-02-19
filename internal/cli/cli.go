package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"

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

func PrintBanner() {
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
	fmt.Printf("\n")
	fmt.Println("-----------------------------------------------------------")
	fmt.Printf("| %-27s %-9s %-9s %s |\n", "SUITE", "PASSED", "FAILED", "SKIPPED")
	fmt.Println("-----------------------------------------------------------")
	for groupName, groupResults := range results {
		fmt.Printf("| %-25s %8d %9d %10d |\n", groupName,
			groupResults[0],
			groupResults[1],
			groupResults[2])
		fmt.Println("-----------------------------------------------------------")
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

	fmt.Print(ClearLineCode + "[ " + CheckResultTagSkip + " ] " + checkName + "  (" + reason + ")\n")
}

func PrintCheckRunning(checkName string) {
	stopChan = make(chan bool)
	checkLoggerChan = make(chan string)

	line := "[ " + CheckResultTagRunning + " ] " + checkName
	if !isTTY() {
		line += "\n"
	}

	fmt.Print(line)

	go updateRunningCheckLine(checkName, stopChan)
}

func PrintCheckPassed(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagPass + " ] " + checkName + "\n")
}

func PrintCheckFailed(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagFail + " ] " + checkName + "\n")
}

func PrintCheckAborted(checkName, reason string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagAborted + " ] " + checkName + "  (" + reason + ")\n")
}

func PrintCheckErrored(checkName string) {
	stopCheckLineGoroutine()

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
