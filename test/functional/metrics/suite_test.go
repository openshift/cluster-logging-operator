package metrics

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFunctionMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterLogging Functional Suite - SampleCollector")
}
