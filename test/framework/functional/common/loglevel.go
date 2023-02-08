package common

import (
	"os"
	"strconv"

	log "github.com/ViaQ/logerr/v2/log/static"
)

func AdaptLogLevel() string {
	logLevel := "debug"
	if level, found := os.LookupEnv("LOG_LEVEL"); found {
		if i, err := strconv.Atoi(level); err == nil {
			switch i {
			case 0:
				logLevel = "error"
			case 1:
				logLevel = "info"
			case 2:
				logLevel = "debug"
			case 3 - 9:
				logLevel = "trace"
			default:
			}
		} else {
			log.V(1).Error(err, "Unable to set LOG_LEVEL from environment")
		}
	}
	return logLevel
}
