package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

var _ = Describe("[internal][validations] ClusterLogForwarder", func() {

	Context("#validateSingleton", func() {

		It("should fail validation when the name is anything other then 'instance'", func() {
			clf := runtime.NewClusterLogForwarder()
			clf.Name = "foo"
			Expect(validateSingleton(*clf)).To(Not(Succeed()))
		})
		It("should pass validation when the name is 'instance'", func() {
			clf := runtime.NewClusterLogForwarder()
			Expect(validateSingleton(*clf)).To(Succeed())
		})

	})

})
