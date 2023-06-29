package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("[internal][validations] Validate clusterlogging dependency", func() {
	var (
		clf    loggingv1.ClusterLogForwarder
		extras map[string]bool
	)

	BeforeEach(func() {
		extras = map[string]bool{}
		clf = loggingv1.ClusterLogForwarder{}
	})

	Context("with clusterlogforwarder named `instance`", func() {
		BeforeEach(func() {
			clf.Name = constants.SingletonName
		})
		It("should fail if clusterlogging not available", func() {
			extras[constants.ClusterLoggingAvailable] = false
			Expect(ValidateClusterLoggingDependency(clf, nil, extras)).To(Not(Succeed()))
		})
		It("should pass if clusterlogging is available", func() {
			extras[constants.ClusterLoggingAvailable] = true
			Expect(ValidateClusterLoggingDependency(clf, nil, extras)).To(Succeed())
		})
	})

	Context("with custom named clusterlogforwarder", func() {
		BeforeEach(func() {
			clf.Name = "custom-clf"
		})
		It("should pass if clusterlogging is available", func() {
			extras[constants.ClusterLoggingAvailable] = true
			Expect(ValidateClusterLoggingDependency(clf, nil, extras)).To(Succeed())
		})
		It("should pass if clusterlogging is not available", func() {
			extras[constants.ClusterLoggingAvailable] = false
			Expect(ValidateClusterLoggingDependency(clf, nil, extras)).To(Succeed())
		})
	})

})
