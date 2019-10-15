package test

import (
	"fmt"
	"os"
	"strings"

	expectations "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func StringsContain(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

//Debug is a convenient log mechnism to spit content to STDOUT
func Debug(value string, object interface{}) {
	if os.Getenv("TEST_UNTRUNCATED_DIFF") != "" {
		format.TruncatedDiff = false
	}
	if os.Getenv("TEST_DEBUG") != "" {
		fmt.Printf("%s\n%v\n", value, object)
	}
}

func normalize(in string) string {
	out := []string{}
	for _, line := range strings.Split(in, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return strings.Join(out, "\n")
}

//TestExpectation is a helper struct to allow chaining expectations
type TestExpectation struct {
	act string
}

func Expect(act string) *TestExpectation {
	return &TestExpectation{act}
}

func (t *TestExpectation) ToEqual(exp string) {
	Debug("actual:", t.act)
	Debug("expected:", exp)
	expectations.Expect(normalize(t.act)).To(expectations.Equal(normalize(exp)))
}
