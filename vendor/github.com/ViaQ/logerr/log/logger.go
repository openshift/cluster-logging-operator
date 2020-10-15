package log

import (
	"github.com/ViaQ/logerr/kverrors"
	"github.com/go-logr/logr"
)

// WrapLogger wraps the logger with the internal Logger
func WrapLogger(l logr.Logger) *Logger {
	switch l.(type) {
	case *Logger:
		return l.(*Logger)
	default:
		return &Logger{l}
	}
}

// Logger wraps zapr.Logger and fixes the Error method to log errors
// as structured. zaprs Logger enforces zap.Error(err) which will break
// our structured errors that implement zapcore.MarshalLogObject.
// Ref: https://github.com/go-logr/zapr/blob/49ca6b4dc551f8fdf9fe385fbd7a60ee3b846a21/zapr.go#L133
type Logger struct {
	base logr.Logger
}

func (l *Logger) Enabled() bool {
	return l.base.Enabled()
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.base.Info(msg, keysAndValues...)
}

func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	var e error
	if ee, ok := err.(kverrors.Error); ok {
		e = ee
	} else {
		// If err is not structured then convert to a KVError so that it is structured for consistency
		e = kverrors.New(err.Error())
	}
	// this uses a nil err because the base zapr.Logger.Error implementation enforces zap.Error(err)
	// which converts the provided err to a standard string. Since we are using a complex err
	// which could be a pkg/errors.KVError we want to pass err as a complex object which zap
	// can then serialize according to KVError.MarshalLogObject()
	l.base.Error(nil, msg, append(keysAndValues, []interface{}{KeyError, e}...)...)
}

func (l *Logger) V(level int) logr.Logger {
	return WrapLogger(l.base.V(level))
}

func (l *Logger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return WrapLogger(l.base.WithValues(keysAndValues...))
}

func (l *Logger) WithName(name string) logr.Logger {
	return WrapLogger(l.base.WithName(name))
}

