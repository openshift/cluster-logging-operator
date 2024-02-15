// go:build vector
package schema

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/schema/otel"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	timestamp           = "2023-08-28T12:59:28.573159188+00:00"
	timestampNano int64 = 1693227568573159188
)

var _ = Describe("[Functional][Normalization][Schema] OTEL", func() {
	var (
		framework    *functional.CollectorFunctionalFramework
		appNamespace string
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(loggingv1.LogCollectionTypeVector)
		framework.Forwarder.Annotations = map[string]string{constants.AnnotationEnableSchema: constants.Enabled}
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should normalize application logs to OTEL format for HTTP sink", func() {
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(loggingv1.InputNameApplication).
			ToHttpOutputWithSchema(constants.OTELSchema)

		ExpectOK(framework.Deploy())

		appNamespace = framework.Pod.Namespace

		// Write message to namespace
		crioLine := functional.NewCRIOLogMessage(timestamp, "Format me to OTEL!", false)
		Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())
		// Read log
		raw, err := framework.ReadRawApplicationLogsFrom(loggingv1.OutputTypeHttp)
		Expect(err).To(BeNil(), "Expected no errors reading the logs for type")

		logs, err := otel.ParseLogs(utils.ToJsonLogs(raw))

		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		otelLog := logs[0].ContainerLog
		Expect(otelLog.TimeUnixNano).To(Equal(timestampNano), "Expect timestamp to be converted into unix nano")
		Expect(otelLog.SeverityText).ToNot(BeNil(), "Expect severityText to exist")
		Expect(otelLog.SeverityNumber).To(Equal(9), "Expect severityNumber to parse to 9")
		Expect(otelLog.Resources).ToNot(BeNil(), "Expect resources to exist")
		Expect(otelLog.Resources.K8s.Namespace.Name).To(Equal(appNamespace), "Expect namespace name to be nested under k8s.namespace")
	})

})
