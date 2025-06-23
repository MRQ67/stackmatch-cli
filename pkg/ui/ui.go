package ui

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

// isTerminal checks if the output is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// colorize adds ANSI color codes if output is a terminal
func colorize(text, colorCode string) string {
	if !isTerminal() {
		return text
	}
	return colorCode + text + colorReset
}

// Success returns a green success message
func Success(format string, a ...interface{}) string {
	return colorize(fmt.Sprintf(format, a...), colorGreen)
}

// Error returns a red error message
func Error(format string, a ...interface{}) string {
	return colorize(fmt.Sprintf(format, a...), colorRed)
}

// Warning returns a yellow warning message
func Warning(format string, a ...interface{}) string {
	return colorize(fmt.Sprintf(format, a...), colorYellow)
}

// Info returns a blue info message
func Info(format string, a ...interface{}) string {
	return colorize(fmt.Sprintf(format, a...), colorBlue)
}

// Confirm asks the user for confirmation
func Confirm(prompt string, defaultYes bool) (bool, error) {
	var options string
	if defaultYes {
		options = " [Y/n] "
	} else {
		options = " [y/N] "
	}

	fmt.Print(Info("❔ ") + prompt + options)

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil && err.Error() != "unexpected newline" {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "" {
		return defaultYes, nil
	}

	return response == "y" || response == "yes", nil
}

// ProgressBar is a simple progress bar implementation
type ProgressBar struct {
	total   int
	current int
	desc    string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(max int, description string) *ProgressBar {
	return &ProgressBar{
		total:   max,
		current: 0,
		desc:    description,
	}
}

// Add updates the progress
func (p *ProgressBar) Add(n int) error {
	p.current += n
	p.Render()
	return nil
}

// Render renders the progress bar
func (p *ProgressBar) Render() {
	if !isTerminal() {
		return
	}

	width := 30
	percent := float64(p.current) / float64(p.total)
	filled := int(float64(width) * percent)

	bar := "[" + strings.Repeat("=", filled) +
		strings.Repeat(" ", width-filled) + "] " +
		fmt.Sprintf("%d/%d", p.current, p.total)

	fmt.Fprintf(os.Stderr, "\r%s %s %s", "⌛", p.desc, bar)
}

// Close finishes the progress bar
func (p *ProgressBar) Close() {
	if isTerminal() {
		fmt.Fprintln(os.Stderr)
	}
}

// PrintSuccess prints a success message
func PrintSuccess(format string, a ...interface{}) {
	fmt.Fprintln(os.Stdout, Success("✓")+" "+fmt.Sprintf(format, a...))
}

// PrintError prints an error message
func PrintError(err error, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if err != nil {
		msg = fmt.Sprintf("%s: %v", msg, err)
	}
	fmt.Fprintln(os.Stderr, Error("✗")+" "+msg)
}

// PrintWarning prints a warning message
func PrintWarning(format string, a ...interface{}) {
	fmt.Fprintln(os.Stdout, Warning("!")+" "+fmt.Sprintf(format, a...))
}

// PrintInfo prints an info message
func PrintInfo(format string, a ...interface{}) {
	fmt.Fprintln(os.Stdout, Info("ℹ")+" "+fmt.Sprintf(format, a...))
}

// Spinner is a simple spinner implementation
type Spinner struct {
	msg string
}

// NewSpinner creates a new spinner
func NewSpinner(msg string) *Spinner {
	fmt.Fprintf(os.Stderr, "⌛ %s... ", msg)
	return &Spinner{msg: msg}
}

// Close finishes the spinner
func (s *Spinner) Close() {
	if isTerminal() {
		fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", len(s.msg)+5)+"\r")
	}
}
