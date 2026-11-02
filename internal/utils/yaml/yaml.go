package yaml

import (
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"gopkg.in/yaml.v3"
)

// MustMarshal returns a YAML string of a value, or an empty string on error.
// Uses gopkg.in/yaml.v3 which respects `yaml:` struct tags.
// Recovers from panics during marshaling and logs the error.
func MustMarshal(v interface{}) (value string) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic during marshal: %v", r)
			log.V(0).WithName("MustMarshal").Error(err, "unable to marshal object to YAML", "object", v)
			value = ""
		}
	}()

	out, err := yaml.Marshal(v)
	if err != nil {
		log.V(0).WithName("MustMarshal").Error(err, "unable to marshal object to YAML", "object", v)
		return ""
	}
	return string(out)
}

// Marshal marshals a value to YAML bytes using gopkg.in/yaml.v3.
// Returns error if marshaling fails.
func Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// MarshalIndent marshals a value to YAML bytes with custom indentation.
// indent specifies the number of spaces to use for indentation.
func MarshalIndent(v interface{}, indent int) ([]byte, error) {
	encoder := yaml.NewEncoder(nil)
	encoder.SetIndent(indent)
	return yaml.Marshal(v)
}

// Unmarshal unmarshals YAML data into a value using gopkg.in/yaml.v3.
func Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// UnmarshalStrict unmarshals YAML data into a value with strict mode enabled.
// Strict mode returns an error if the YAML contains duplicate keys or unknown fields.
func UnmarshalStrict(data []byte, v interface{}) error {
	decoder := yaml.NewDecoder(nil)
	decoder.KnownFields(true)
	return yaml.Unmarshal(data, v)
}

// MarshalToMap marshals a value to YAML and back to a map[string]interface{}.
// This is useful for converting structs to maps while preserving YAML tag behavior.
func MarshalToMap(v interface{}) (map[string]interface{}, error) {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := yaml.Unmarshal(bytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}
	return result, nil
}
