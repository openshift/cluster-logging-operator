package metrics

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("[Functional][Metrics]Function testing of fluentd metrics", func() {

	const (
		sampleMetric = "# HELP fluentd_output_status_buffer_total_bytes"
	)

	var (
		framework *functional.CollectorFunctionalFramework
	)

	AfterEach(func() {
		framework.Cleanup()
	})

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
	})
	It("when using a service address should return successfully", func() {
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
