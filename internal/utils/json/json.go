package json

import (
	"encoding/json"
	log "github.com/ViaQ/logerr/v2/log/static"
)

func marshalString(b []byte, err error) string {
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// JSONString returns a JSON string of a value, or an error message.
func MustMarshal(v interface{}) (value string) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.V(0).WithName("MustMarshal").Error(err, "unable to marshal object", "object", v)
		return ""
	}
	return string(out)
}
