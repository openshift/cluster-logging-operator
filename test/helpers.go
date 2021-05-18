package test

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"
	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/format"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

func init() {
	if os.Getenv("TEST_UNTRUNCATED_DIFF") != "" || os.Getenv("TEST_FULL_DIFF") != "" {
		format.TruncatedDiff = false
	}
	// Set up logging for tests.
	log.MustInit("test")
	if level, err := strconv.Atoi(os.Getenv("LOG_LEVEL")); err == nil {
		log.SetLogLevel(level)
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

// Unmarshal JSON or YAML string into a value according to k8s rules.
// Uses sigs.k8s.io/yaml.
func Unmarshal(s string, v interface{}) error { return yaml.Unmarshal([]byte(s), v) }

// MustUnmarshal unmarshals JSON or YAML into a value, panic on error.
func MustUnmarshal(s string, v interface{}) { Must(Unmarshal(s, v)) }

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
	maxPrefix := validation.DNS1123LabelMaxLength - (len(timeStamp) + len(random)*2 + 2)
	if len(prefix) > maxPrefix {
		prefix = prefix[:maxPrefix]
	}
	return fmt.Sprintf("%s-%s-%x", prefix, timeStamp, random[:])
}

// GinkgoCurrentTest tries to get the current Ginkgo test description.
// Returns true if successful, false if not in a ginkgo test.
func GinkgoCurrentTest() (g ginkgo.GinkgoTestDescription, ok bool) {
	defer func() { _ = recover() }()
	g = ginkgo.CurrentGinkgoTestDescription() // May panic if not in a ginkgo test.
	ok = true
	return
}

// UniqueNameForTest generates a unique name prefixed with the current
// Ginkgo test name, or the string "test" if not in a Ginkgo test.
func UniqueNameForTest() (name string) {
	prefix := "test"
	if ct, ok := GinkgoCurrentTest(); ok {
		prefix = ct.TestText
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

// LogBeginEnd logs an  l.V(3) begin message, returns func to log an lV(2) end message.
// End message includes elapsed time and error errp and *errp are non nil.
// Use it to log the time spent in a function like this:
//     func blah() (err error) {
//         defer LogBeginEnd(log, "eating", &err)()
//         ...
// Note the trailing () - this calls LogBeginEnd() and defers calling the func it returns.
func LogBeginEnd(l logr.Logger, msg string, errp *error, kv ...interface{}) func() {
	l.V(3).Info("begin: "+msg, kv...)
	start := time.Now()
	return func() {
		kv := append(kv, "elapsed", time.Since(start).String())
		if errp != nil && *errp != nil {
			l.V(2).Error(*errp, "error: "+msg, kv...)
		} else {
			l.V(2).Info("end  : "+msg, kv...)
		}
	}
}

func Escapelines(logline string) string {
	logline = strings.ReplaceAll(logline, "\\", "\\\\")
	logline = strings.ReplaceAll(logline, "\"", "\\\"")
	return logline
}

// GitRoot joins paths to the root of the git repository.
// Panics if current directory is not inside a git repository.
func GitRoot(paths ...string) string {
	wd, err := os.Getwd()
	Must(err)
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return filepath.Join(append([]string{dir}, paths...)...)
		}
		dir = filepath.Dir(dir)
		if dir == "/" {
			panic(fmt.Errorf("not in a git repository: %v", wd))
		}
	}
}
