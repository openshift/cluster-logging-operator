package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("[internal][validations] ClusterLogForwarder", func() {

	Context("#validateSingleton", func() {
		var (
			extras    = map[string]bool{}
			k8sClient client.Client
		)
		It("should fail validation when the name is anything other then 'instance'", func() {
			clf := runtime.NewClusterLogForwarder()
			clf.Name = "foo"
			Expect(validateSingleton(*clf, k8sClient, extras)).To(Not(Succeed()))
		})
		It("should pass validation when the name is 'instance'", func() {
			clf := runtime.NewClusterLogForwarder()
			Expect(validateSingleton(*clf, k8sClient, extras)).To(Succeed())
		})

	})

})
