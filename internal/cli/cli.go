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

// PrintBanner displays an ASCII banner and returns a cleanup function.
//
// It prints a pre-defined banner using the standard Print function.
// The returned function can be invoked to stop any background activity
// associated with the banner, such as ticker updates or log sniffing.
func PrintBanner() {
	fmt.Print(banner)
}

// cliCheckLogSniffer is an internal helper that captures log output from the
// check framework and forwards it to the command‑line interface.
//
// It implements io.Writer, allowing slog handlers to write directly into the
// CLI's output stream. The Write method checks whether the destination is a
// terminal, formats the incoming bytes appropriately, and writes them while
// reporting the number of bytes written or any errors that occur.
type cliCheckLogSniffer struct{}

// isTTY reports whether the current output stream is a terminal.
//
// It returns true if stdout is attached to a TTY device, allowing
// colorized or interactive output; otherwise it returns false.
// The function checks the file descriptor of os.Stdout using IsTerminal.
func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// updateRunningCheckLine creates a goroutine that periodically updates the terminal line showing the status of a running check until signaled to stop.
//
// It takes a string identifier and a channel that signals when the check has finished.
// The function checks if the output is a TTY, then starts a ticker based on tickerPeriodSeconds.
// On each tick it prints the current running check line. When a value is received on the
// stop channel, the ticker stops and the goroutine exits. This allows dynamic progress
// updates in the CLI while a check runs.
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

// getTerminalWidth returns the current terminal width in columns.
//
// It calls golang.org/x/crypto/ssh/terminal.GetSize on standard output
// to determine the number of columns available. If an error occurs,
// it defaults to 80 columns. The function returns the width as an int.
func getTerminalWidth() int {
	width, _, _ := term.GetSize(int(os.Stdin.Fd()))
	return width
}

// cropLogLine truncates a log line to a maximum length, preserving its tail.
//
// It takes a string line and an integer maxLen. If the line exceeds maxLen,
// it returns a shortened version that begins with "... " followed by the last
// (maxLen - 4) characters of the original line. If the line is shorter than or
// equal to maxLen, the original line is returned unchanged. This function is
// used to keep log output concise while still showing the most recent part of
// long messages.
func cropLogLine(line string, maxAvailableWidth int) string {
	// Remove line feeds to avoid the log line to break the cli output.
	filteredLine := strings.ReplaceAll(line, "\n", " ")
	// Print only the chars that fit in the available space.
	if len(filteredLine) > maxAvailableWidth {
		return filteredLine[:maxAvailableWidth]
	}
	return filteredLine
}

// printRunningCheckLine displays a live status line for a running check, updating with elapsed time and the current log message. It takes the check name, start time, and the latest log text, then returns a function that when called will refresh the terminal output. The returned function writes a single line showing the check name, how long it has been running, and a cropped version of the most recent log entry, ensuring the output fits within the current terminal width.
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

// Write writes log output to the terminal and forwards it to the check logger channel.
//
// It implements io.Writer for cliCheckLogSniffer, converting the byte slice to a string,
// trimming trailing newlines, and sending the result on checkLoggerChan if the
// output is not empty. The method returns the number of bytes written and any error
// encountered during processing. If the receiver is attached to a TTY, it also
// clears the current line before writing.
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

// PrintResultsTable prints a formatted table of test results.
//
// It takes a map where the key is a string label and the value is a slice
// of integers representing counts of each result type. The function writes
// the table to standard output, showing totals for each category (pass,
// fail, skip, etc.) in a human‑readable format. No values are returned.
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

// stopCheckLineGoroutine creates and returns a function that signals the
// check‑line goroutine to terminate.
//
// The returned function closes the global stop channel used by the goroutine
// responsible for sniffing log lines from CliCheckLogSniffer. When invoked,
// it causes the goroutine to exit cleanly, allowing the caller to wait for
// shutdown before proceeding. This helper centralises the shutdown logic so
// callers need not manipulate channels directly.
func stopCheckLineGoroutine() {
	if stopChan == nil {
		// This may happen for checks that were skipped if no compliant nor non-compliant objects found.
		return
	}

	stopChan <- true
	// Make this chnanel immediately unavailable.
	stopChan = nil
}

// PrintCheckSkipped outputs a message indicating that a check has been skipped.
//
// It takes the name of the check and a reason string, stops any running
// progress animation goroutine, then prints the formatted skip notification
// to the console using the Print helper function. The function returns no
// value.
func PrintCheckSkipped(checkName, reason string) {
	// It shouldn't happen too often, but some checks might be set as skipped inside the checkFn
	// if neither compliant objects nor non-compliant objects were found.
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagSkip + " ] " + checkName + "  (" + reason + ")\n")
}

// PrintCheckRunning displays a dynamic running check status line and returns a function to stop it.
//
// It starts a background goroutine that periodically updates the terminal line with
// progress information received from a channel. The returned function signals the
// goroutine to terminate, ensuring graceful cleanup. The input string parameter is
// used as an initial message displayed before the periodic updates begin.
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

// PrintCheckPassed prints a success message for a check and stops the progress goroutine.
//
// It receives a string containing the name of the check that has passed,
// stops any running progress indicator, and outputs a formatted success
// line using the Print helper function. The message is prefixed with a green
// status tag to indicate success.
func PrintCheckPassed(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagPass + " ] " + checkName + "\n")
}

// PrintCheckFailed stops the check line goroutine and logs a failure message.
//
// It takes a string argument containing the reason for failure, stops any
// ongoing progress output, and prints a formatted error using the package's
// Print helper. The returned function performs no additional actions but is
// intended to be deferred by callers so that cleanup occurs automatically.
func PrintCheckFailed(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagFail + " ] " + checkName + "\n")
}

// PrintCheckAborted reports that a check has been aborted and stops the progress line.
//
// It takes two string arguments: the name of the check and an optional message.
// The function stops any running check line goroutine, prints a formatted abort
// notification using the provided message, and then returns.
func PrintCheckAborted(checkName, reason string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagAborted + " ] " + checkName + "  (" + reason + ")\n")
}

// PrintCheckErrored creates a deferred cleanup routine for a failed check.
//
// It accepts an error message string, stops the ongoing check line
// goroutine, outputs the formatted error using Print, and returns a
// zero-argument function that can be invoked to perform this cleanup.
// The returned function is intended to be used in a defer statement
// when a check encounters an error.
func PrintCheckErrored(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagError + " ] " + checkName + "\n")
}

// WrapLines splits a string into lines no longer than the specified width.
//
// It breaks the input on newline characters, then further splits long
// lines into chunks that do not exceed maxWidth. Words are preserved
// as whole units; if a single word is longer than maxWidth it will
// appear on its own line even if it exceeds the limit.
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

// LineAlignLeft left‑justifies a string to a given width.
//
// It returns the input text padded with spaces on the right so that its total
// length equals the specified width. If the text is longer than the width,
// it is returned unchanged. The function uses Sprintf internally to format
// the result.
func LineAlignLeft(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, s)
}

// LineAlignCenter centers a line within a given width, padding with spaces.
//
// It takes an input string and a desired total length. The function
// calculates the number of spaces needed on each side to center the
// text, then returns a new string that is padded accordingly.
// If the requested length is less than or equal to the input's
// length, the original string is returned unchanged.
func LineAlignCenter(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, fmt.Sprintf("%[1]*s", (w+len(s))/2, s)) //nolint:mnd // magic number
}

// LineColor formats a line with ANSI color codes based on the provided tag and message.
//
// It takes two string arguments: a tag indicating the status (such as "PASS" or "FAIL") and
// the corresponding text to display. The function returns a single formatted string that
// includes the appropriate escape sequences for foreground colors, allowing terminal
// output to be colorized according to the tag's meaning. The returned string can be
// printed directly to standard output.
func LineColor(s, color string) string {
	return color + s + Reset
}
