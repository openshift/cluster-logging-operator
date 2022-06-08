package test

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unsafe"

	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/format"
	"golang.org/x/net/html"
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

// Unmarshal JSON or YAML string into a value according to k8s rules.
// Uses sigs.k8s.io/yaml.
func Unmarshal(s string, v interface{}) error { return yaml.Unmarshal([]byte(s), v) }

// MustUnmarshal unmarshals JSON or YAML into a value, panic on error.
func MustUnmarshal(s string, v interface{}) { Must(Unmarshal(s, v)) }

var (
	uniqueReplace = regexp.MustCompile("[^a-z0-9]+")
)

// UniqueName generates a unique DNS label string starting with prefix.
// Illegal character sequences in prefix are replaced with "-".
// Suffix sorts by time of creation if < 12h apart.
func UniqueName(prefix string) string {
	pre := strings.ToLower(prefix)
	pre = uniqueReplace.ReplaceAllLiteralString(pre, "-")
	pre = strings.Trim(pre, "-")

	// NOTE: Short names are important. A namespace name may be
	// concatenated with service names in route host names, if they
	// exceed the limit of 63 chars things will fail.

	if len(pre) > 16 {
		pre = pre[:16]
	}

	secs := uint16(time.Now().Unix() % (12 * 60 * 60)) // Keep 12 hours == 43200 secs, 2 bytes.
	var unique [unsafe.Sizeof(secs) + 3]byte
	binary.BigEndian.PutUint16(unique[0:], secs)
	_, err := rand.Read(unique[unsafe.Sizeof(secs):])
	Must(err)
	uniqueStr := strings.ToLower(base32.StdEncoding.EncodeToString(unique[:]))
	name := fmt.Sprintf("%s-%s", pre, uniqueStr)
	if len(validation.IsDNS1123Label(name)) != 0 {
		panic(fmt.Errorf("invalid DNS label %q cannot use prefix %q", name, pre))
	}
	return name
}

// GinkgoCurrentTest tries to get the current Ginkgo test description.
// Returns true if successful, false if not in a ginkgo test.
func GinkgoCurrentTest() (g ginkgo.GinkgoTestDescription, ok bool) {
	defer func() { _ = recover() }()
	g = ginkgo.CurrentGinkgoTestDescription() // May panic if not in a ginkgo test.
	ok = true
	return
}

// UniqueNameForTest generates a unique name for a test.
func UniqueNameForTest() string { return UniqueName("test") }

// LoggingNamespace returns env-var NAMESPACE or "openshift-logging".
func LoggingNamespace() string {
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		return ns
	}
	return constants.OpenshiftNS
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

// HTTPError returns an error constructed from resp.Body if resp.Status is not 2xx.
// Returns nil otherwise.
func HTTPError(resp *http.Response) error {
	if resp.Status[0] == '2' {
		return nil
	}
	msg, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if txt, err := HTMLBodyText(bytes.NewReader(msg)); err == nil {
		msg = txt
	}
	return fmt.Errorf("%v: %v", resp.Status, string(msg))
}

// HTMLBodyText extracts text from the <body> of a HTML document.
// Useful for test error messages, not for much else.
func HTMLBodyText(r io.Reader) ([]byte, error) {
	var txt []byte
	z := html.NewTokenizer(r)
	inBody := false
	isBody := func(tag []byte, _ bool) bool { return bytes.Equal(tag, []byte("body")) }
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return nil, z.Err()
		case html.TextToken:
			if inBody {
				txt = append(txt, z.Text()...)
			}
		case html.StartTagToken:
			if isBody(z.TagName()) {
				inBody = true
			}
		case html.EndTagToken:
			if isBody(z.TagName()) {
				return bytes.TrimSpace(regexpNNN.ReplaceAllLiteral(txt, []byte{'\n', '\n'})), nil
			}
		}
	}
}

var regexpNNN = regexp.MustCompile("\n\n(\n*)")

// MapIndices indexes into a multi-level map[string]interface{} unmarshalled from JSON.
// Result is like `m[index1][index2]..`
// Returns nil if an index is not found.
func MapIndices(m interface{}, index ...interface{}) (value interface{}) {
	defer func() {
		// Convert reflect panics to errors.
		if r := recover(); r != nil {
			value = fmt.Errorf("%v", r)
		}
	}()
	v := reflect.ValueOf(m)
	for _, i := range index {
		v = v.MapIndex(reflect.ValueOf(i))
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		if !v.IsValid() {
			return nil
		}
		if v.Kind() != reflect.Map {
			return v.Interface()
		}
	}
	return v.Interface()
}

// TrimLines trims leading and trailing whitespace from every line in lines.
// Useful for comparing configuration snippets with varied indenting.
func TrimLines(lines string) string {
	b := &strings.Builder{}
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			fmt.Fprintln(b, line)
		}
	}
	return b.String()
}
