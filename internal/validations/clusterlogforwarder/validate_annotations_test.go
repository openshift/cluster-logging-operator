package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

var _ = Describe("[internal][validations] validate clusterlogforwarder annotations", func() {
	var (
		clf *loggingv1.ClusterLogForwarder
	)

	BeforeEach(func() {
		clf = runtime.NewClusterLogForwarder()
	})

	Context("#validateLogLevel", func() {
		It("should pass validation if no annotations are set", func() {
			Expect(validateAnnotations(*clf, nil, nil)).To(Succeed())
		})

		It("should fail validation if log level is not supported", func() {
			clf.Annotations = map[string]string{constants.AnnotationVectorLogLevel: "foo"}
			err, _ := validateAnnotations(*clf, nil, nil)
			Expect(err).ToNot(BeNil())
		})

		DescribeTable("valid log levels", func(level string) {
			clf.Annotations = map[string]string{constants.AnnotationVectorLogLevel: level}
			Expect(validateAnnotations(*clf, nil, nil)).To(Succeed())
		},
			Entry("should pass with level trace", "trace"),
			Entry("should pass with level debug", "debug"),
			Entry("should pass with level info", "info"),
			Entry("should pass with level warn", "warn"),
			Entry("should pass with level error", "error"),
			Entry("should pass with level off", "off"))
	})
})
