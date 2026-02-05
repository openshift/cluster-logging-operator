package errors

import (
	log "github.com/ViaQ/logerr/v2/log/static"
)

func LogIfError(err error, messages ...string) {
	if err != nil {
		msg := ""
		if len(messages) > 0 {
			msg = messages[0]
		}
		if len(messages) > 1 {
			messages = messages[1:]
		}
		log.Error(err, msg, messages)
	}
}
