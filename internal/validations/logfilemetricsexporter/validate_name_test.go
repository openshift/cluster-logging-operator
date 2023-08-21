package logfilemetricsexporter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
)

var _ = Describe("[internal][validations] ClusterLogForwarder", func() {

	Context("#validateName", func() {

		// This conflicts with the legacy CLF
		It("should fail validation when the name is NOT 'instance' and NOT in the 'openshift-logging' namespace", func() {
			clf := loggingruntime.NewLogFileMetricExporter("some-ns", "not-instance")
			Expect(validateName(clf)).To(Not(Succeed()))
		})

		It("should fail validation when the name is NOT 'instance' and in any namespace other then 'openshift-logging'", func() {
			clf := loggingruntime.NewLogFileMetricExporter("some-ns", constants.SingletonName)
			Expect(validateName(clf)).To(Not(Succeed()))
		})

		It("should fail validation when the name is NOT 'instance' and in the 'openshift-logging' namespace", func() {
			clf := loggingruntime.NewLogFileMetricExporter(constants.OpenshiftNS, "my-instance")
			Expect(validateName(clf)).To(Not(Succeed()))
		})

		It("should pass validation when the name is 'instance' and in the 'openshift-logging' namespace", func() {
			clf := loggingruntime.NewLogFileMetricExporter(constants.OpenshiftNS, constants.SingletonName)
			Expect(validateName(clf)).To(Succeed())
		})

	})

})
