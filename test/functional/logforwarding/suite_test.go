package logforwarding

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestLogForwarding(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogging Functional Suite - LogForwarding"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-functional-log-forwarding.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
