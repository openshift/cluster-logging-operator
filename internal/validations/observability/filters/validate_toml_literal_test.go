package filters

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("TOML literal injection guards (LOG-9752)", func() {

	Context("#validateFieldPath", func() {
		It("should accept a valid quoted segment", func() {
			Expect(validateFieldPath(`.kubernetes."test-foo"`)).To(BeEmpty())
		})
		It("should reject the literal multiline terminator inside a quoted segment", func() {
			Expect(validateFieldPath(`.kubernetes."x'''y"`)).To(ContainSubstring("must not contain the sequence"))
		})
	})

	Context("#validateKubeAPIAuditFilter", func() {
		newSpec := func(rules ...auditv1.PolicyRule) obs.FilterSpec {
			return obs.FilterSpec{
				Name:         "my-audit",
				Type:         obs.FilterTypeKubeAPIAudit,
				KubeAPIAudit: &obs.KubeAPIAudit{Rules: rules},
			}
		}
		It("should accept benign rules", func() {
			spec := newSpec(auditv1.PolicyRule{Level: "Metadata", Namespaces: []string{"openshift-*"}, Verbs: []string{"get"}})
			Expect(ValidateFilter(spec).Status).To(Equal(metav1.ConditionTrue))
		})
		It("should reject a namespace wildcard containing the terminator", func() {
			spec := newSpec(auditv1.PolicyRule{Level: "Metadata", Namespaces: []string{"x'''y"}})
			cond := ValidateFilter(spec)
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Message).To(ContainSubstring("must not contain the sequence"))
		})
		It("should reject a resourceName containing the terminator", func() {
			spec := newSpec(auditv1.PolicyRule{Level: "Metadata", Resources: []auditv1.GroupResources{{Resources: []string{"pods"}, ResourceNames: []string{"a'''b"}}}})
			cond := ValidateFilter(spec)
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Message).To(ContainSubstring("must not contain the sequence"))
		})
		It("should reject a rule level containing the terminator", func() {
			spec := newSpec(auditv1.PolicyRule{Level: "x'''y"})
			cond := ValidateFilter(spec)
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Message).To(ContainSubstring("must not contain the sequence"))
		})
		It("should reject a rule omitStage containing the terminator", func() {
			spec := newSpec(auditv1.PolicyRule{Level: "Metadata", OmitStages: []auditv1.Stage{"x'''y"}})
			cond := ValidateFilter(spec)
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Message).To(ContainSubstring("must not contain the sequence"))
		})
		It("should reject a filter-level omitStage containing the terminator", func() {
			spec := obs.FilterSpec{
				Name:         "my-audit",
				Type:         obs.FilterTypeKubeAPIAudit,
				KubeAPIAudit: &obs.KubeAPIAudit{OmitStages: []auditv1.Stage{"x'''y"}},
			}
			cond := ValidateFilter(spec)
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Message).To(ContainSubstring("must not contain the sequence"))
		})
	})

	Context("#validateOpenshiftLabelsFilter", func() {
		newSpec := func(labels map[string]string) obs.FilterSpec {
			return obs.FilterSpec{
				Name:            "my-labels",
				Type:            obs.FilterTypeOpenshiftLabels,
				OpenshiftLabels: labels,
			}
		}
		It("should accept benign labels", func() {
			spec := newSpec(map[string]string{"environment": "production"})
			Expect(ValidateFilter(spec).Status).To(Equal(metav1.ConditionTrue))
		})
		It("should reject a label value containing the terminator", func() {
			spec := newSpec(map[string]string{"environment": "x'''y"})
			cond := ValidateFilter(spec)
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Message).To(ContainSubstring("must not contain the sequence"))
		})
		It("should reject a label key containing the terminator", func() {
			spec := newSpec(map[string]string{"x'''y": "value"})
			cond := ValidateFilter(spec)
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Message).To(ContainSubstring("must not contain the sequence"))
		})
	})
})
