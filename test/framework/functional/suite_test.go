package functional

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "[Framework][Functional] Suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-framework-functional.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
