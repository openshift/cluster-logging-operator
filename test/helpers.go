package test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/onsi/gomega/format"
	"sigs.k8s.io/yaml"
)

func init() {
	if os.Getenv("TEST_UNTRUNCATED_DIFF") != "" || os.Getenv("TEST_FULL_DIFF") != "" {
		format.TruncatedDiff = false
	}
}

//Debug is a convenient log mechnism to spit content to STDOUT
func Debug(value string, object interface{}) {
	if os.Getenv("TEST_DEBUG") != "" {
		fmt.Printf("%s\n%v\n", value, object)
	}
}

func marshalString(b []byte, err error) string {
	if err != nil {
		return err.Error()
	}
	return string(b)
}

//JSONString returns a JSON string of a value, or an error message.
func JSONString(v interface{}) string {
	return marshalString(json.MarshalIndent(v, "", "  "))
}

//YAMLString returns a YAML string of a value, using the sigs.k8s.io/yaml package,
//or an error message.
func YAMLString(v interface{}) string { return marshalString(yaml.Marshal(v)) }
