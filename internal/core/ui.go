package core

import "io"

// UI defines the interface for user interaction and output.
type UI interface {
	// Section prints a section header.
	Section(title string)
	// Title prints a main title.
	Title(title string)
	// Success prints a success message.
	Success(msg string)
	// Info prints an informational message.
	Info(msg string)
	// Debug prints a debug message.
	Debug(msg string)
	// Warning prints a warning message.
	Warning(msg string)
	// Error prints an error message.
	Error(msg string)
	// Printf prints a formatted message to standard output.
	Printf(format string, args ...interface{})
	// Println prints a line to standard output.
	Println(args ...interface{})
	// WithWriter returns a new UI instance writing to the specified writer.
	WithWriter(w io.Writer) UI
}
