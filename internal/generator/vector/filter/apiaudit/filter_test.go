package apiaudit

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

var _ = Describe("KubeAPIAudit filter VRL generation", func() {
	Context("when OmitManagedFields is set", func() {
		It("should generate VRL code to delete managedFields when set to true", func() {
			omitTrue := true
			spec := &obs.KubeAPIAudit{
				Rules: []auditv1.PolicyRule{
					{
						Level:             auditv1.LevelRequestResponse,
						OmitManagedFields: &omitTrue,
					},
				},
			}
			vrl, err := NewFilter(spec).VRL()
			Expect(err).NotTo(HaveOccurred())
			Expect(vrl).To(ContainSubstring("omit_managed = true"))
			Expect(vrl).To(ContainSubstring("del(.requestObject.metadata.managedFields)"))
			Expect(vrl).To(ContainSubstring("del(.responseObject.metadata.managedFields)"))
		})

		It("should generate VRL code to not delete managedFields when set to false", func() {
			omitFalse := false
			spec := &obs.KubeAPIAudit{
				Rules: []auditv1.PolicyRule{
					{
						Level:             auditv1.LevelRequestResponse,
						OmitManagedFields: &omitFalse,
					},
				},
			}
			vrl, err := NewFilter(spec).VRL()
			Expect(err).NotTo(HaveOccurred())
			Expect(vrl).To(ContainSubstring("omit_managed = false"))
			Expect(vrl).To(ContainSubstring("del(.requestObject.metadata.managedFields)"))
			Expect(vrl).To(ContainSubstring("del(.responseObject.metadata.managedFields)"))
		})

		It("should initialize omit_managed to false when not configured", func() {
			spec := &obs.KubeAPIAudit{
				Rules: []auditv1.PolicyRule{
					{
						Level: auditv1.LevelRequestResponse,
						// OmitManagedFields not set (nil)
					},
				},
			}
			vrl, err := NewFilter(spec).VRL()
			Expect(err).NotTo(HaveOccurred())
			// Should initialize to false
			Expect(vrl).To(ContainSubstring("omit_managed = false"))
			// Should not set it to true in any rule
			lines := strings.Split(vrl, "\n")
			falseCount := 0
			for _, line := range lines {
				if strings.Contains(line, "omit_managed = false") {
					falseCount++
				}
				if strings.Contains(line, "omit_managed = true") {
					Fail("Should not set omit_managed = true when not configured")
				}
			}
			// Should only have one initialization
			Expect(falseCount).To(Equal(1))
			// Should still have the conditional deletion code
			Expect(vrl).To(ContainSubstring("if omit_managed"))
		})
	})
})
