package observability

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[internal][validations] validate clusterlogforwarder annotations", func() {
	var (
		clf     *obs.ClusterLogForwarder
		context internalcontext.ForwarderContext
	)

	BeforeEach(func() {
		clf = obsruntime.NewClusterLogForwarder("foo", "bar", runtime.Initialize)
		context = internalcontext.ForwarderContext{
			Forwarder: clf,
		}
	})

	Context("#validateLogLevel", func() {
		It("should pass validation if no annotations are set", func() {
			validateAnnotations(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		})

		It("should fail validation if log level is not supported", func() {
			clf.Annotations = map[string]string{constants.AnnotationVectorLogLevel: "foo"}
			validateAnnotations(context)
			Expect(clf.Status.Conditions).To(HaveCondition(obs.ConditionTypeLogLevel, false, obs.ReasonLogLevelSupported, ".*must be one of.*"))
		})

		DescribeTable("valid log levels", func(level string) {
			clf.Annotations = map[string]string{constants.AnnotationVectorLogLevel: level}
			validateAnnotations(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		},
			Entry("should pass with level trace", "trace"),
			Entry("should pass with level debug", "debug"),
			Entry("should pass with level info", "info"),
			Entry("should pass with level warn", "warn"),
			Entry("should pass with level error", "error"),
			Entry("should pass with level off", "off"))
	})
})
