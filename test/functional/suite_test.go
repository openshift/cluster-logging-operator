package functional

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestMatchers(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "Functional framework suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-matchers.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
