package fluentlegacy

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogForwarder E2E Suite - Fluentd Forward Legacy"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-fluentd-legacy.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
