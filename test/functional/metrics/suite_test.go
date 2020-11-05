package metrics

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestFunctionMetrics(t *testing.T) {
	RegisterFailHandler(Fail)

	tc := "ClusterLogging Functional Suite - Metrics"
	jr := reporters.NewJUnitReporter("/tmp/artifacts/junit/junit-functional-metrics.xml")
	RunSpecsWithDefaultAndCustomReporters(t, tc, []Reporter{jr})
}
