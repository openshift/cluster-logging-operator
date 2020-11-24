package outputs

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestFunctionalOutputs(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogging Functional Output Suite"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-functional-output.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
