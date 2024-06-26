package metrics

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

var _ = Describe("[Functional][Metrics]Function testing of collector metrics", func() {

	const (
		sampleMetric = `# HELP vector_adaptive_concurrency_in_flight adaptive_concurrency_in_flight
# TYPE vector_adaptive_concurrency_in_flight histogram`
	)

	var (
		framework *functional.CollectorFunctionalFramework
	)

	AfterEach(func() {
		framework.Cleanup()
	})

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToHttpOutput()
	})
	It("should return successfully when all outputs are up", func() {
		Expect(framework.Deploy()).To(BeNil())
		metrics, _ := framework.RunCommand(constants.CollectorName, "curl", "-ksv", fmt.Sprintf("https://%s.%s:24231/metrics", framework.Name, framework.Namespace))
		Expect(metrics).To(ContainSubstring(sampleMetric))
	})

	It("should return successfully even when the output is down", func() {
		Expect(framework.DeployWithVisitor(func(builder *runtime.PodBuilder) error { return nil })).To(BeNil())
		metrics, _ := framework.RunCommand(constants.CollectorName, "curl", "-ksv", fmt.Sprintf("https://%s.%s:24231/metrics", framework.Name, framework.Namespace))
		Expect(metrics).To(ContainSubstring(sampleMetric))
	})

})
