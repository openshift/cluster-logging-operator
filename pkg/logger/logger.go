package logger

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

//Debugf logs messages at level 2
func Debugf(format string, objects ...interface{}) {
	logrus.Debugf(format, objects...)
}

//IsDebugEnabled returns true if loglevel is 2
func IsDebugEnabled() bool {
	return logrus.GetLevel() == logrus.DebugLevel
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
