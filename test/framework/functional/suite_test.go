package functional

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestFunctional(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")

	tc := "[Framework][Functional] Suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-framework-functional.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
