package log

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ViaQ/logerr/internal/kv"
	"github.com/ViaQ/logerr/kverrors"
	"github.com/go-logr/logr"
)

const (
	ErrorKey     = "cause"
	MessageKey   = "message"
	TimeStampKey = "ts"
	ComponentKey = "component"
	VerbosityKey = "v"
)

// Verbosity is a level of verbosity to log between 0 and math.MaxInt32
// However it is recommended to keep the numbers between 0 and 3
type Verbosity int

// TimestampFunc returns a string formatted version of the current time.
// This should probably only be used with tests or if you want to change
// the default time formatting of the output logs.
var TimestampFunc = func() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// Logger writes logs to a specified output
type Logger struct {
	mtx       sync.RWMutex
	verbosity Verbosity
	output    io.Writer
	context   map[string]interface{}
	encoder   Encoder
	name      string
}

func NewLogger(name string, w io.Writer, v Verbosity, e Encoder, keysAndValues ...interface{}) *Logger {
	return &Logger{
		name:      name,
		verbosity: v,
		output:    w,
		context:   kv.ToMap(keysAndValues...),
		encoder:   e,
	}
}

// withValues clones the logger and appends keysAndValues
// but returns a struct instead of the logr.Logger interface
func (l *Logger) withValues(keysAndValues ...interface{}) *Logger {
	ll := NewLogger(l.name, l.output, l.verbosity, l.encoder)
	ll.context = kv.AppendMap(kv.ToMap(keysAndValues...), l.context)
	return ll
}

// WithValues clones the logger and appends keysAndValues
func (l *Logger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return l.withValues(keysAndValues...)
}

// SetOutput sets the writer that JSON is written to
func (l *Logger) SetOutput(w io.Writer) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.output = w
}

// Enabled tests whether this Logger is enabled.  For example, commandline
// flags might be used to set the logging verbosity and disable some info
// logs.
func (l *Logger) Enabled() bool {
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	return l.verbosity <= Verbosity(logLevel)
}

// log will log the message. It DOES NOT check Enabled() first so that should
// be checked by it's callers
func (l *Logger) log(msg string, context map[string]interface{}) {
	m := kv.AppendMap(context, map[string]interface{}{
		MessageKey:   msg,
		TimeStampKey: TimestampFunc(),
		ComponentKey: l.name,
		// VerbosityKey: l.verbosity,
	})
	err := l.encoder.Encode(l.output, m)
	if err != nil {
		// expand first so we can quote later
		orig := fmt.Sprintf("%#v", context)
		_, _ = fmt.Fprintf(l.output, `{"message","failed to encode message", "encoder":"%T","log":%q,"cause":%q}`, l.encoder, orig, err)
	}
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	if !l.Enabled() {
		return
	}
	l.log(msg, kv.AppendMap(l.context, kv.ToMap(keysAndValues...)))
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
	if !l.Enabled() {
		return
	}

	if err == nil {
		l.Info(msg, keysAndValues)
		return
	}

	switch err.(type) {
	case *kverrors.KVError:
		// nothing to be done
	default:
		err = kverrors.New(err.Error())
	}

	l.Info(msg, append(keysAndValues, ErrorKey, err)...)
}

// V returns an Logger value for a specific verbosity level, relative to
// this Logger.  In other words, V values are additive.  V higher verbosity
// level means a log message is less important.  It's illegal to pass a log
// level less than zero.
func (l *Logger) V(v int) logr.Logger {
	return NewLogger(l.name, l.output, Verbosity(v)+l.verbosity, l.encoder, l.context)
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (l *Logger) WithName(name string) logr.Logger {
	newname := name

	if l.name != "" {
		newname = fmt.Sprintf("%s_%s", l.name, name)
	}

	return NewLogger(
		newname,
		l.output,
		l.verbosity,
		l.encoder,
		l.context,
	)
}

// KeysAndValues returns the keysAndValues assigned to this logger
func (l *Logger) KeysAndValues() map[string]interface{} {
	return l.context
}
