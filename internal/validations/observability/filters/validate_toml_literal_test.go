package filters

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

// The collector config is serialized as YAML (quoted/escaped or block scalar),
// so filter values containing the former TOML literal multiline terminator or
// newlines can no longer break out of the generated VRL. Admission must accept
// them; safety holds by construction. See LOG-10361.
var _ = Describe("Filter values that formerly tripped the TOML literal guards (LOG-10361)", func() {

	Context("#validateFieldPath", func() {
		It("should accept a valid quoted segment", func() {
			Expect(validateFieldPath(`.kubernetes."test-foo"`)).To(BeEmpty())
		})
		It("should accept a quoted segment containing the former TOML literal terminator", func() {
			Expect(validateFieldPath(`.kubernetes."x'''y"`)).To(BeEmpty())
		})
	})

	Context("#KubeAPIAudit", func() {
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
		It("should accept a rule value containing the former terminator and a newline", func() {
			spec := newSpec(auditv1.PolicyRule{Level: "Metadata", Namespaces: []string{"x'''y\nz"}})
			Expect(ValidateFilter(spec).Status).To(Equal(metav1.ConditionTrue))
		})
	})

	Context("#OpenshiftLabels", func() {
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
		It("should accept a label key and value containing the former terminator", func() {
			spec := newSpec(map[string]string{"x'''y": "v'''w"})
			Expect(ValidateFilter(spec).Status).To(Equal(metav1.ConditionTrue))
		})
	})
})
