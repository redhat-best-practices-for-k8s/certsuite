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

// PrintBanner Displays a banner at startup
//
// This function writes the predefined banner string to standard output using
// the fmt package. It is invoked during application initialization to show
// branding or version information. No parameters are taken, and it does not
// return a value.
func PrintBanner() {
	fmt.Print(banner)
}

// cliCheckLogSniffer forwards terminal output to a logging channel
//
// This type implements an io.Writer that captures data written by the CLI when
// running in a TTY environment. When Write is called, it attempts to send the
// byte slice as a string over a dedicated channel; if the channel is not ready
// or closed, the data is silently dropped to avoid blocking execution. In
// non‑TTY scenarios, all writes are simply acknowledged without any side
// effects.
type cliCheckLogSniffer struct{}

// isTTY determines whether standard input is a terminal
//
// The function checks if the current process’s stdin corresponds to an
// interactive terminal device by converting its file descriptor to an integer
// and using the external library’s IsTerminal call. It returns true when
// output can be formatted for a tty, otherwise false. This value influences how
// log lines are printed or suppressed in the CLI.
func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// updateRunningCheckLine updates the running check status line with elapsed time and latest log
//
// This routine starts a ticker that triggers every to refresh the console
// output for a running test. It listens on a channel for new log messages,
// updating the displayed line accordingly, and stops when a stop signal is
// received. The function prints the check name, elapsed time, and optionally a
// cropped latest log if terminal width permits.
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

// getTerminalWidth Determines the current terminal width in columns
//
// It calls a system routine to query the size of the standard input device,
// returning the number of columns available for output. The value is used to
// format log lines so they fit within the terminal without wrapping or
// truncating unexpectedly.
func getTerminalWidth() int {
	width, _, _ := term.GetSize(int(os.Stdin.Fd()))
	return width
}

// cropLogLine Trims a log line to fit terminal width
//
// The function removes newline characters from the input string and then
// truncates it if its length exceeds the specified maximum width. It returns
// the processed string, which is safe to display in a single-line CLI output
// without breaking formatting.
func cropLogLine(line string, maxAvailableWidth int) string {
	// Remove line feeds to avoid the log line to break the cli output.
	filteredLine := strings.ReplaceAll(line, "\n", " ")
	// Print only the chars that fit in the available space.
	if len(filteredLine) > maxAvailableWidth {
		return filteredLine[:maxAvailableWidth]
	}
	return filteredLine
}

// printRunningCheckLine Displays the progress of a running check
//
// It prints a status line that includes the check name, elapsed time since
// start, and optionally a cropped log message when running in a terminal. If
// output is not a TTY it simply writes the line with a newline. The function
// clears the current terminal line before printing to keep the display updated.
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

// cliCheckLogSniffer.Write Writes log data to a channel when running in a terminal
//
// When the process is attached to a TTY, this method attempts to send the
// provided byte slice as a string into a dedicated logger channel without
// blocking; if the channel is not ready or closed, the data is silently
// dropped. In non‑TTY environments it simply returns the length of the input
// and no error, effectively discarding output. The function always reports the
// full number of bytes processed.
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

// PrintResultsTable Displays a formatted summary of test suite outcomes
//
// The function accepts a mapping from group names to integer slices that
// represent passed, failed, and skipped counts. It outputs a neatly aligned
// table with column headers and separators, iterating over each group to show
// its results. After listing all groups, it adds blank lines for readability.
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

// stopCheckLineGoroutine Signals the check line goroutine to stop
//
// This function checks whether a global channel used for signalling is set,
// sends a true value to that channel if it exists, then clears the reference so
// subsequent calls have no effect. It is called by various print functions when
// a check completes or is aborted, ensuring any ongoing line output goroutine
// terminates cleanly.
func stopCheckLineGoroutine() {
	if stopChan == nil {
		// This may happen for checks that were skipped if no compliant nor non-compliant objects found.
		return
	}

	stopChan <- true
	// Make this chnanel immediately unavailable.
	stopChan = nil
}

// PrintCheckSkipped Logs a skipped check with its reason
//
// This function stops the ongoing check line goroutine, then prints a formatted
// message indicating that the specified check was skipped along with the
// provided reason. The output includes control codes to clear the current
// terminal line and displays the skip tag followed by the check name and
// explanation. No value is returned.
func PrintCheckSkipped(checkName, reason string) {
	// It shouldn't happen too often, but some checks might be set as skipped inside the checkFn
	// if neither compliant objects nor non-compliant objects were found.
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagSkip + " ] " + checkName + "  (" + reason + ")\n")
}

// PrintCheckRunning Displays a running check status message
//
// The function prints an initial line indicating that a specific check is in
// progress, appending a newline when output is not a terminal to keep the
// display clean. It then starts a background goroutine that updates this line
// every second with elapsed time and any new log messages until the check
// completes and signals the stop channel.
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

// PrintCheckPassed Shows a passed check with formatted output
//
// The function stops any active line‑printing goroutine, then writes a clear
// line indicator followed by a pass tag and the provided check name to standard
// output. It uses predefined constants for formatting and ensures the display
// is updated correctly before returning.
func PrintCheckPassed(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagPass + " ] " + checkName + "\n")
}

// PrintCheckFailed Displays a failed check status line
//
// The function stops the running goroutine that updates the check progress,
// then prints a formatted message indicating failure for the given check name.
// It writes the output directly to standard output with escape codes to clear
// the previous line and show a red "FAIL" tag followed by the check identifier.
func PrintCheckFailed(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagFail + " ] " + checkName + "\n")
}

// PrintCheckAborted Notifies the user that a check has been aborted
//
// This routine stops any ongoing line‑printing goroutine, then outputs a
// formatted message indicating the check’s name and the reason for abortion.
// The output includes special control codes to clear the current terminal line
// before displaying the status tag. No value is returned.
func PrintCheckAborted(checkName, reason string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagAborted + " ] " + checkName + "  (" + reason + ")\n")
}

// PrintCheckErrored Stops the progress display and shows an error line
//
// This routine halts any ongoing check‑line goroutine, clears the current
// terminal line, and prints a formatted message indicating that the specified
// check has failed with an error. The output includes a clear line code, an
// error tag, and the check identifier.
func PrintCheckErrored(checkName string) {
	stopCheckLineGoroutine()

	fmt.Print(ClearLineCode + "[ " + CheckResultTagError + " ] " + checkName + "\n")
}

// WrapLines Breaks a string into lines that fit within a maximum width
//
// The function splits the input text on newline characters, then examines each
// line to see if it exceeds the specified width. Lines longer than the limit
// are broken into words and reassembled so no resulting line surpasses the
// maximum length. The wrapped lines are returned as a slice of strings.
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

// LineAlignLeft left‑justifies a string to a given column width
//
// The function takes an input string and a desired width, returning the string
// padded with spaces on the right so that its total length equals the specified
// width. It uses formatted printing with a negative field width to achieve left
// alignment. If the original string exceeds the requested width, it is returned
// unchanged without truncation.
func LineAlignLeft(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, s)
}

// LineAlignCenter Centers a string within a specified width
//
// The function takes an input string and a target width, then returns the
// string padded with spaces so it appears centered when printed. It calculates
// padding by determining how many leading spaces are needed to shift the
// original text toward the middle of the given width. The resulting string is
// always exactly the specified length.
func LineAlignCenter(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, fmt.Sprintf("%[1]*s", (w+len(s))/2, s)) //nolint:mnd // magic number
}

// LineColor Adds ANSI color codes around text
//
// This function takes a plain string and a color code, prefixes the string with
// the color escape sequence, appends the reset code, and returns the resulting
// colored string. It is used to display terminal output in different colors
// without altering the original content.
func LineColor(s, color string) string {
	return color + s + Reset
}
