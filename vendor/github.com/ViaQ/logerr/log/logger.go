package log

import (
	"github.com/ViaQ/logerr/kverrors"
	"github.com/go-logr/logr"
)

// WrapLogger wraps the logger with the internal Logger
func WrapLogger(l logr.Logger) *Logger {
	if lg, ok := l.(*Logger); ok {
		return lg
	}
	return &Logger{l}
}

// Logger wraps zapr.Logger and fixes the Error method to log errors
// as structured. zaprs Logger enforces zap.Error(err) which will break
// our structured errors that implement zapcore.MarshalLogObject.
// Ref: https://github.com/go-logr/zapr/blob/49ca6b4dc551f8fdf9fe385fbd7a60ee3b846a21/zapr.go#L133
type Logger struct {
	base logr.Logger
}

// Enabled tests whether this Logger is enabled.  For example, commandline
// flags might be used to set the logging verbosity and disable some info
// logs.
func (l *Logger) Enabled() bool {
	return l.base.Enabled()
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.base.Info(msg, keysAndValues...)
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	if err == nil {
		l.base.Error(nil, msg, keysAndValues...)
		return
	}
	e := err
	if _, ok := err.(*kverrors.KVError); !ok {
		// If err is not structured then convert to a KVError so that it is structured for consistency
		e = kverrors.New(err.Error())
	}
	// this uses a nil err because the base zapr.Logger.Error implementation enforces zap.Error(err)
	// which converts the provided err to a standard string. Since we are using a complex err
	// which could be a pkg/errors.KVError we want to pass err as a complex object which zap
	// can then serialize according to KVError.MarshalLogObject()
	l.base.Error(nil, msg, append(keysAndValues, []interface{}{KeyError, e}...)...)
}

// V returns an Logger value for a specific verbosity level, relative to
// this Logger.  In other words, V values are additive.  V higher verbosity
// level means a log message is less important.  It's illegal to pass a log
// level less than zero.
func (l *Logger) V(level int) logr.Logger {
	return WrapLogger(l.base.V(level))
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (l *Logger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return WrapLogger(l.base.WithValues(keysAndValues...))
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (l *Logger) WithName(name string) logr.Logger {
	return WrapLogger(l.base.WithName(name))
}
