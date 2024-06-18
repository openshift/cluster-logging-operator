package observability

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validations", func() {

	Context("#MustUndeployCollector", func() {

		Context("should be false", func() {
			It("when the conditions are empty", func() {
				Expect(MustUndeployCollector([]metav1.Condition{})).To(BeFalse())
			})
			It("when there are no unauthorized conditions", func() {
				Expect(MustUndeployCollector([]metav1.Condition{
					internalobs.NewCondition(obs.ConditionTypeAuthorized, obs.ConditionTrue, "", "some message"),
				})).To(BeFalse())
			})
		})
		Context("should be true", func() {
			It("when not authorized to collect", func() {
				cond := []metav1.Condition{
					internalobs.NewCondition(obs.ConditionTypeAuthorized, obs.ConditionFalse, obs.ReasonClusterRoleMissing, "some message"),
				}
				Expect(MustUndeployCollector(cond)).To(BeTrue())
			})
		})

	})

})
