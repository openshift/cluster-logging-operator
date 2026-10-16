package toml

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml"
)

// literalMultilineTerminator is the delimiter (three single-quote characters) go-toml v1 uses
// for a struct field tagged `multiline:"true" literal:"true"`. The encoder emits such a field
// verbatim between these delimiters with NO escaping (see tomltree_write.go), so a value that
// itself contains the terminator breaks out of the TOML string and any following bytes are
// parsed as raw TOML. When the value is derived from an untrusted ClusterLogForwarder spec
// (VRL generated for outputs, filters, prune paths, kubeAPIAudit wildcards, ...) this is a
// config-injection sink. Reject such values before marshalling so a malformed collector config
// is never rendered. See LOG-9752 (and LOG-9704 for the originally reported drop-filter sink).
const literalMultilineTerminator = "'''"

func MustMarshal(v interface{}) string {
	out, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	return out
}

func Marshal(v interface{}) (string, error) {
	if err := validateLiteralMultiline(reflect.ValueOf(v), ""); err != nil {
		return "", err
	}
	out := new(bytes.Buffer)
	encoder := toml.NewEncoder(out).Indentation("").Order(toml.OrderPreserve)
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	return out.String(), nil
}

// validateLiteralMultiline walks v the same way the TOML encoder does and rejects any string
// field tagged `multiline:"true" literal:"true"` whose value contains the literal multiline
// terminator. Because go-toml v1 emits those fields with no escaping, an embedded terminator
// would allow arbitrary TOML injection into the generated collector configuration.
func validateLiteralMultiline(v reflect.Value, path string) error {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return validateLiteralMultiline(v.Elem(), path)
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" { // skip unexported fields
				continue
			}
			fv := v.Field(i)
			fieldPath := field.Name
			if path != "" {
				fieldPath = path + "." + field.Name
			}
			if fv.Kind() == reflect.String &&
				field.Tag.Get("literal") == "true" && field.Tag.Get("multiline") == "true" &&
				strings.Contains(fv.String(), literalMultilineTerminator) {
				return fmt.Errorf("field %q contains the TOML literal multiline terminator %q, which would allow config injection", fieldPath, literalMultilineTerminator)
			}
			if err := validateLiteralMultiline(fv, fieldPath); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := validateLiteralMultiline(v.Index(i), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if err := validateLiteralMultiline(v.MapIndex(k), fmt.Sprintf("%s[%v]", path, k)); err != nil {
				return err
			}
		}
	}
	return nil
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
