package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("[internal][validations] ClusterLogForwarder", func() {

	Context("#validateName", func() {
		var (
			extras    = map[string]bool{}
			k8sClient client.Client
		)

		// This conflicts with the legacy CLF
		It("should fail validation when the name is 'collector' and in the 'openshift-logging' namespace", func() {
			clf := runtime.NewClusterLogForwarder()
			clf.Name = "collector"
			Expect(validateName(*clf, k8sClient, extras)).To(Not(Succeed()))
		})

		It("should pass validation when the name is 'collector' and in any namespace other then 'openshift-logging'", func() {
			clf := runtime.NewClusterLogForwarder()
			clf.Namespace = "foobar"
			clf.Name = "collector"
			Expect(validateName(*clf, k8sClient, extras)).To(Succeed())
		})

		It("should fail validation when the name results in an object that will fail name validation (e.g. service)", func() {
			clf := runtime.NewClusterLogForwarder()
			clf.Namespace = "foobar"
			clf.Name = "65409debug-3y8sw019"
			Expect(validateName(*clf, k8sClient, extras)).To(Not(Succeed()))
		})

	})

})
