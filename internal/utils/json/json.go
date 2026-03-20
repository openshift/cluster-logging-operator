package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	log "github.com/ViaQ/logerr/v2/log/static"
)

// JSONString returns a JSON string of a value, or an error message.
func MustMarshal(v interface{}) (value string) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.V(0).WithName("MustMarshal").Error(err, "unable to marshal object", "object", v)
		return ""
	}
	return string(out)
}

// SortedMap converts a map to a structure that marshals with sorted keys
type SortedMap map[string]interface{}

func (sm SortedMap) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(sm))
	for k := range sm {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(",")
		}
		// Marshal key
		keyBytes, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteString(":")
		// Marshal value
		valBytes, err := json.Marshal(sm[k])
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

// MarshalSorted marshals a map with sorted keys
func MarshalSorted(m map[string]interface{}) ([]byte, error) {
	return json.Marshal(SortedMap(m))
}

// MarshalToSortedMap marshals a value to JSON and back to a map
func MarshalToSortedMap(v interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}
	return result, nil
}
