package logger

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

//Debug logs messages at level 2
func Debug(message string) {
	logrus.Debug(message)
}

//Debugf logs messages at level 2
func Debugf(format string, objects ...interface{}) {
	logrus.Debugf(format, objects...)
}

//Warnf logs messages at level 2
func Warnf(format string, objects ...interface{}) {
	logrus.Warnf(format, objects...)
}

//Errorf logs messages at level error
func Errorf(format string, objects ...interface{}) {
	logrus.Errorf(format, objects...)
}

//Error logs messages at level error
func Error(message string) {
	logrus.Error(message)
}

//Infof logs messages at level info
func Infof(format string, objects ...interface{}) {
	logrus.Infof(format, objects...)
}

//IsDebugEnabled returns true if loglevel is 2
func IsDebugEnabled() bool {
	return logrus.GetLevel() == logrus.DebugLevel
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

//DebugObject pretty prints the given object
func DebugObject(sprintfMessage string, object interface{}) {
	if IsDebugEnabled() && object != nil {
		pretty, err := json.MarshalIndent(object, "", "  ")
		if err != nil {
			logrus.Debugf("Error marshalling object %v for debug log: %v", object, err)
		}
		logrus.Debugf(sprintfMessage, string(pretty))
	}
}
func init() {
	level := os.Getenv("LOG_LEVEL")
	if strings.TrimSpace(level) == "" {
		return
	}
	parsed, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.Warnf("Unable to parse LOG_LEVEL %q", level)
		return
	}
	logrus.SetLevel(parsed)
}
