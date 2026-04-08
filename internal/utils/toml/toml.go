package toml

import (
	"bytes"
	"fmt"

	"github.com/pelletier/go-toml"
)

func MustMarshal(v interface{}) string {
	out, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	return out
}

func Marshal(v interface{}) (string, error) {
	out := new(bytes.Buffer)
	encoder := toml.NewEncoder(out).Indentation("").Order(toml.OrderPreserve)
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	return out.String(), nil
}

func Unmarshal(s string, v interface{}) error {
	return toml.Unmarshal([]byte(s), v)
}

func MustUnmarshal(s string, v interface{}) {
	if err := Unmarshal(s, v); err != nil {
		panic(fmt.Sprintf("Error unmarshalling toml: %v\n%s\n", err, s))
	}
}

// SetValue modifies a value at the specified path in TOML config and returns the modified config.
// Path is specified as a slice of keys, e.g. []string{"sinks", "output_s3", "batch", "timeout_secs"}
// Creates intermediate tables as needed.
func SetValue(config string, path []string, value interface{}) (string, error) {
	tree, err := toml.LoadBytes([]byte(config))
	if err != nil {
		return "", fmt.Errorf("failed to parse TOML: %w", err)
	}

	tree.SetPath(path, value)

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(tree); err != nil {
		return "", fmt.Errorf("failed to encode TOML: %w", err)
	}

	return buf.String(), nil
}
