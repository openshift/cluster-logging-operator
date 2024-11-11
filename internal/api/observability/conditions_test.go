package observability_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[internal][api][observability]", func() {

	Context("Helper Conditions", func() {

		It("should prune conditions", func() {
			conditions := []metav1.Condition{
				NewConditionFromPrefix(v1.ConditionTypeValidPipelinePrefix, "bar", true, "", ""),
				NewConditionFromPrefix(v1.ConditionTypeValidPipelinePrefix, "foo-bar", true, "", ""),
				NewConditionFromPrefix(v1.ConditionTypeValidPipelinePrefix, "bar-baz", true, "", ""),
			}
			pipelines := Pipelines{
				{Name: "bar"},
				{Name: "foo-bar"},
			}
			Expect(len(conditions)).To(Equal(3))
			PruneConditions(&conditions, pipelines, v1.ConditionTypeValidPipelinePrefix)
			Expect(len(conditions)).To(Equal(2))
		})
	})
})
