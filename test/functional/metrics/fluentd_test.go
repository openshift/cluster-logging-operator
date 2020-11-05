package metrics

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
)

var _ = Describe("[Metrics] Function testing of fluentd metrics", func() {

	const (
		sampleMetric = "# HELP fluentd_output_status_buffer_total_bytes"
	)

	var (
		framework *functional.FluentdFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when using a service address", func() {
		It("should return successfully", func() {
			cmd := fmt.Sprintf("curl -ksv https://%s.%s:24231/metrics", framework.Name, framework.Namespace)
			metrics, _ := framework.RunCommand(cmd)
			Expect(metrics).To(ContainSubstring(sampleMetric))
		})
	})

})
