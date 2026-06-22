package api

import (
	"fmt"
	"io"
	"os"
	"time"
)

// DefaultLogger provides timestamped logging similar to the bash script
type DefaultLogger struct {
	writer io.Writer
}

// NewLogger creates a new logger that writes to the given writer
func NewLogger(w io.Writer) Logger {
	return &DefaultLogger{writer: w}
}

// Log writes a timestamped log message
func (l *DefaultLogger) Log(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	if _, err := fmt.Fprintf(l.writer, "%s %s\n", timestamp, message); err != nil {
		// Fallback to stderr if primary writer fails
		fmt.Fprintf(os.Stderr, "LOGGER ERROR: failed to write log: %v\nOriginal message: %s %s\n", err, timestamp, message)
	}
}

// Begin logs a BEGIN message and returns a function that logs the corresponding END message
func (l *DefaultLogger) Begin(format string, args ...interface{}) func() {
	l.Log("BEGIN "+format, args...)
	return func() {
		l.Log("END "+format, args...)
	}
}

// Warn logs a warning message with WARN prefix
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	l.Log("WARN: "+format, args...)
}

// Info logs an informational message with INFO prefix
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.Log("INFO: "+format, args...)
}

// Logf is an alias for Log for convenience
func (l *DefaultLogger) Logf(format string, args ...interface{}) {
	l.Log(format, args...)
}
