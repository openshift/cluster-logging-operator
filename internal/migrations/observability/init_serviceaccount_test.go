package observability

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#InitServiceAccount", func() {
	It("should initialize the audiance to the default when blank", func() {
		spec, cond := InitServiceAccount(obs.ClusterLogForwarderSpec{})
		Expect(spec.ServiceAccount.Audience).To(Equal(defaultAudience))
		Expect(cond).To(BeEmpty())
	})
})
