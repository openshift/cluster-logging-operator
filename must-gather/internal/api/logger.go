package api

import (
	"fmt"
	"io"
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
	fmt.Fprintf(l.writer, "%s %s\n", timestamp, message)
}

// Logf is an alias for Log for convenience
func (l *DefaultLogger) Logf(format string, args ...interface{}) {
	l.Log(format, args...)
}
