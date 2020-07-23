package multiple_outputs

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestLogForwarding(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogForwarder E2E Suite - Ensure Forward to Multiple Outputs"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-multiple-outputs.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
