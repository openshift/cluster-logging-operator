package test

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/format"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

func init() {
	if os.Getenv("TEST_UNTRUNCATED_DIFF") != "" || os.Getenv("TEST_FULL_DIFF") != "" {
		format.TruncatedDiff = false
	}
}

func marshalString(b []byte, err error) string {
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// JSONString returns a JSON string of a value, or an error message.
func JSONString(v interface{}) string {
	return marshalString(json.MarshalIndent(v, "", "  "))
}

// JSONLine returns a one-line JSON string, or an error message.
func JSONLine(v interface{}) string { return marshalString(json.Marshal(v)) }

// YAMLString returns a YAML string of a value, using the sigs.k8s.io/yaml package,
// or an error message.
func YAMLString(v interface{}) string { return marshalString(yaml.Marshal(v)) }

// Must panics if err is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustUnmarshal unmarshals JSON or YAML into a value, panic on error.
func MustUnmarshal(s string, v interface{}) {
	// Note sigs.k8s.io/yaml can parse JSON or YAML -  JSON is a subset of YAML.
	Must(yaml.Unmarshal([]byte(s), v))
}

var (
	uniqueReplace = regexp.MustCompile("[^a-z0-9]+")
	uniqueTrim    = regexp.MustCompile("(^[^a-z]+)|(-+$)")
)

// UniqueName generates a unique DNS label string starting with prefix.
// Illegal character sequences in prefix are replaced with "-".
// Suffix starts with a time-stamp so names sort by time of creation.
func UniqueName(prefix string) string {
	// Make prefix into a valid DNS Label that can be used to name resources.
	prefix = strings.ToLower(prefix)
	prefix = uniqueReplace.ReplaceAllLiteralString(prefix, "-")
	prefix = uniqueTrim.ReplaceAllLiteralString(prefix, "")

	// Timestamp for time ordering
	timeStamp := time.Now().Format("150405")

	// Random data for uniqueness, use crypto/rand for real randomness.
	random := [4]byte{}
	_, err := rand.Read(random[:])
	Must(err)

	// Don't exceed max label length, truncate prefix and keep time+random suffix.
	maxPrefix := validation.DNS1123LabelMaxLength - (len(timeStamp) + len(random)*2 + 1)
	if len(prefix) > maxPrefix {
		prefix = prefix[:maxPrefix]
	}
	return fmt.Sprintf("%s-%s%x", prefix, timeStamp, random[:])
}

// UniqueNameForTest generates a unique name prefixed with the current
// Ginkgo test name, or the string "test" if not in a Ginkgo test.
func UniqueNameForTest() string {
	prefix := ginkgo.CurrentGinkgoTestDescription().TestText
	if prefix == "" {
		return "test"
	}
	return UniqueName(prefix)
}

// LoggingNamespace returns env-var NAMESPACE or "openshift-logging".
func LoggingNamespace() string {
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		return ns
	}
	return OpenshiftLoggingNS
}
