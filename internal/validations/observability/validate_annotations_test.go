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
			validateLogLevelAnnotation(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		})

		It("should fail validation if log level is not supported", func() {
			clf.Annotations = map[string]string{constants.AnnotationVectorLogLevel: "foo"}
			validateLogLevelAnnotation(context)
			Expect(clf.Status.Conditions).To(HaveCondition(obs.ConditionTypeLogLevel, false, obs.ReasonLogLevelSupported, ".*must be one of.*"))
		})

		DescribeTable("valid log levels", func(level string) {
			clf.Annotations = map[string]string{constants.AnnotationVectorLogLevel: level}
			validateLogLevelAnnotation(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		},
			Entry("should pass with level trace", "trace"),
			Entry("should pass with level debug", "debug"),
			Entry("should pass with level info", "info"),
			Entry("should pass with level warn", "warn"),
			Entry("should pass with level error", "error"),
			Entry("should pass with level off", "off"))
	})

	Context("#validateMaxUnavailable", func() {
		It("should pass validation if no annotations are set", func() {
			validateMaxUnavailableAnnotation(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		})

		DescribeTable("invalid maxUnavailable values", func(value string) {
			clf.Annotations = map[string]string{constants.AnnotationMaxUnavailable: value}
			validateMaxUnavailableAnnotation(context)
			Expect(clf.Status.Conditions).To(HaveCondition(obs.ConditionTypeMaxUnavailable, false, obs.ReasonMaxUnavailableSupported, ".*must be an absolute number or a valid percentage.*"))
		},
			Entry("should fail with empty value", ""),
			Entry("should fail with value 0", "0"),
			Entry("should fail with value 01", "01"),
			Entry("should fail with value 0%", "0%"),
			Entry("should fail with value 101%", "101%"),
			Entry("should fail with value '-1'", "-1"),
			Entry("should fail with value '5.5'", "5.5"),
			Entry("should fail with value 'foo'", "foo"))

		DescribeTable("valid maxUnavailable values", func(value string) {
			clf.Annotations = map[string]string{constants.AnnotationMaxUnavailable: value}
			validateMaxUnavailableAnnotation(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		},
			Entry("should pass with value 1", "1"),
			Entry("should pass with value 1%", "1%"),
			Entry("should pass with value 100", "100"),
			Entry("should pass with value 100%", "100%"))
	})

	Context("#validateUseKubeCacheAnnotation", func() {
		It("should pass validation if no annotations are set", func() {
			validateUseKubeCacheAnnotation(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		})

		DescribeTable("invalid kube-cache values", func(value string) {
			clf.Annotations = map[string]string{constants.AnnotationKubeCache: value}
			validateUseKubeCacheAnnotation(context)
			Expect(clf.Status.Conditions).To(HaveCondition(obs.ConditionTypeUseKubeCache, false, obs.ReasonKubeCacheSupported, ".*must be one of.*"))
		},
			Entry("should fail with empty value", ""),
			Entry("should fail with value 0", "0"),
			Entry("should fail with value false", "false"),
			Entry("should fail with value disabled", "disabled"),
			Entry("should fail with value 'foo'", "foo"))

		DescribeTable("valid kube-cache values", func(value string) {
			clf.Annotations = map[string]string{constants.AnnotationKubeCache: value}
			validateUseKubeCacheAnnotation(context)
			Expect(clf.Status.Conditions).To(BeEmpty())
		},
			Entry("should pass with value true", "true"),
			Entry("should pass with value True", "True"),
			Entry("should pass with value enabled", "enabled"),
			Entry("should pass with value Enabled", "Enabled"))
	})
})
