package kv

import (
	"encoding/json"
	"fmt"
)

// ToMap converts keysAndValues to a map
func ToMap(keysAndValues ...interface{}) map[string]interface{} {
	kve := map[string]interface{}{}

	for i, kv := range keysAndValues {
		if i%2 == 1 {
			continue
		}
		if len(keysAndValues) <= i+1 {
			continue
		}
		kve[fmt.Sprintf("%s", kv)] = keysAndValues[i+1]
	}
	return kve
}

// AppendMap appends one map to another
func AppendMap(a, b map[string]interface{}) map[string]interface{} {
	for k, v := range b {
		a[k] = v
	}
	return a
}

// FromMap converts a map to a key/value slice
func FromMap(m map[string]interface{}) []interface{} {
	res := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		res = append(res, k, v)
	}
	return res
}

// ToJSON marshals keysAndValues to json
func ToJSON(keysAndValues ...interface{}) ([]byte, error) {
	return json.Marshal(ToMap(keysAndValues))
}
