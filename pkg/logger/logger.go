package logger

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func Trace(args ...interface{}) {
	logrus.Trace(args...)
}

func Tracef(format string, objects ...interface{}) {
	logrus.Tracef(format, objects...)
}

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

//Debugf logs messages at level 2
func Debugf(format string, objects ...interface{}) {
	logrus.Debugf(format, objects...)
}

func Warnf(format string, objects ...interface{}) {
	logrus.Warnf(format, objects...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(format string, objects ...interface{}) {
	logrus.Infof(format, objects...)
}

//IsDebugEnabled returns true if loglevel is 2
func IsDebugEnabled() bool {
	return logrus.GetLevel() >= logrus.DebugLevel
}

func init() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		return
	}
	parsed, err := logrus.ParseLevel(level)
	if err != nil {
		parsed = logrus.InfoLevel
		logrus.Warnf("Unable to parse loglevel %q", level)
	}
	logrus.SetLevel(parsed)
}

// JSONString returns a JSON string representation for logging.
func JSONString(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshalling %v for logging: %v", v, err)
	} else {
		return string(b)
	}
}

//DebugObject pretty prints the given object
func DebugObject(sprintfMessage string, object interface{}) {
	if IsDebugEnabled() && object != nil {
		logrus.Debugf(sprintfMessage, JSONString(object))
	}
}
