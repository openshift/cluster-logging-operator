package fluent

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogForwarder E2E Suite - Fluentd Forward"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-fluentd-forward.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
