package log

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	KeyError = "cause"
)

var (
	mtx sync.RWMutex

	// empty logger to prevent nil pointer panics before Init is called
	logger = zapr.NewLogger(zap.NewNop())

	defaultConfig = &zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     true,
		DisableStacktrace: true,
	}
)

// Init initializes the logger. This is required to use logging correctly
// component is the name of the component being used to log messages.
// Typically this is your application name.
//
// keyValuePairs are key/value pairs to be used with all logs in the future
func Init(component string, keyValuePairs ...interface{}) error {
	return InitWithOptions(component, nil, keyValuePairs...)
}

// MustInit calls Init and panics if it returns an error
func MustInit(component string, keyValuePairs ...interface{}) {
	if err := Init(component, keyValuePairs...); err != nil {
		panic(err)
	}
}

// InitWithOptions inits the logger with the provided opts
func InitWithOptions(component string, opts []Option, keyValuePairs ...interface{}) error {
	mtx.Lock()
	defer mtx.Unlock()

	defaultConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	for _, opt := range opts {
		opt(defaultConfig)
	}

	zl, err := defaultConfig.Build(zap.AddCallerSkip(2))
	if err != nil {
		return err
	}

	useLogger(zapr.NewLogger(zl).WithName(component))
	if len(keyValuePairs) > 0 {
		useLogger(logger.WithValues(keyValuePairs))
	}

	return nil
}

// MustInitWithOptions calls InitWithOptions and panics if an error is returned
func MustInitWithOptions(component string, opts []Option, keyValuePairs ...interface{}) {
	if err := InitWithOptions(component, opts, keyValuePairs...); err != nil {
		panic(err)
	}
}

// UseLogger bypasses the requirement for Init and sets the logger to l
func UseLogger(l logr.Logger) {
	mtx.Lock()
	defer mtx.Unlock()
	useLogger(l)
}

func useLogger(l logr.Logger) {
	logger = WrapLogger(l)
}

// Option is a configuration option
type Option func(*zap.Config)

// WithNoTimestamp removes the timestamp from the logged output
// this is primarily used for testing purposes
func WithNoTimestamp() Option {
	return func(c *zap.Config) {
		c.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("")
	}
}

// WithCaller logs the "caller" field
// Example: {"caller": "pkg/log/log.go:32"}
func WithCaller() Option {
	return func(c *zap.Config) {
		c.DisableCaller = false
	}
}

// WithStack logs the "stack" field with the stacktrace
func WithStack() Option {
	return func(c *zap.Config) {
		c.DisableStacktrace = false
	}
}

// WithVerbosity configures the verbosity output
func WithVerbosity(level uint8) Option {
	l := -1 * int(level)
	return func(c *zap.Config) {
		c.Level = zap.NewAtomicLevelAt(zapcore.Level(l))
	}
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func Info(msg string, keysAndValues ...interface{}) {
	mtx.RLock()
	defer mtx.RUnlock()
	logger.Info(msg, keysAndValues...)
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
func Error(err error, msg string, keysAndValues ...interface{}) {
	mtx.RLock()
	defer mtx.RUnlock()
	logger.Error(err, msg, keysAndValues...)
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func WithValues(keysAndValues ...interface{}) logr.Logger {
	mtx.RLock()
	defer mtx.RUnlock()
	return logger.WithValues(keysAndValues...)
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func WithName(name string) logr.Logger {
	mtx.RLock()
	defer mtx.RUnlock()
	return logger.WithName(name)
}

// V returns an Logger value for a specific verbosity level, relative to
// this Logger.  In other words, V values are additive.  V higher verbosity
// level means a log message is less important.
// V(level uint8) Logger
func V(level uint8) logr.Logger {
	return logger.V(int(level))
}
