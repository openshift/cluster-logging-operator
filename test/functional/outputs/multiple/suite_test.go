package multiple

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputs(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "[Functional][Outputs][Multiple] Test Suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-functional-outputs-multiple.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
