package logforwarding

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogForwarder E2E Suite - One-time conversion LF to CLF"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-lf-to-clf-upgrade.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
