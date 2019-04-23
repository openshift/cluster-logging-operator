package test

import (
	"fmt"
	"os"
	"strings"

	expectations "github.com/onsi/gomega"
)

//Debug is a convenient log mechnism to spit content to STDOUT
func Debug(object interface{}) {
	if os.Getenv("TEST_DEBUG") != "" {
		fmt.Println(object)
	}
}

func normalize(in string) string {
	out := []string{}
	for _, line := range strings.Split(in, "\n") {
		trimmed := strings.Trim(line, " \t")
		if trimmed != "" {
			out = append(out, strings.Trim(line, " \t"))
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
	Debug(t.act)
	Debug(exp)
	expectations.Expect(normalize(t.act)).To(expectations.Equal(normalize(exp)))
}
