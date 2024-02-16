package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/status"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = Describe("Validate Splunk:", func() {

	Context("when validating Output Type", func() {

		It("should be valid", func() {
			splunk := &loggingv1.Splunk{
				IndexKey: "kubernetes.labels.test_logging",
			}
			valid, st := VerifySplunk("splunk", splunk)
			Expect(valid).To(BeTrue())
			Expect(map[string]status.Condition{"splunk": st}).To(HaveCondition("Ready", true, "", ""))
		})

		It("should fail if no IndexKey and ", func() {
			splunk := &loggingv1.Splunk{
				IndexKey:  "kubernetes.labels.test_logging",
				IndexName: "application",
			}
			valid, st := VerifySplunk("splunk", splunk)
			Expect(valid).To(BeFalse())
			Expect(map[string]status.Condition{"azureMonitor": st}).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"splunk\": Only one of indexKey or indexName can be set, not both."))
		})

		It("should fail if IndexKey bad format: not allowed symbols", func() {
			splunk := &loggingv1.Splunk{
				IndexKey: "kubernetes.labels.test:logging",
			}
			valid, st := VerifySplunk("splunk", splunk)
			Expect(valid).To(BeFalse())
			Expect(map[string]status.Condition{"splunk": st}).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"splunk\": IndexKey can only contain letters, numbers, and underscores \\(a-zA-Z0-9_\\). Segments that contain characters outside of this range must be quoted."))
		})

		It("should pass with condition Degraded if Fields is included in spec", func() {
			splunk := &loggingv1.Splunk{
				Fields: []string{"foo"},
			}
			valid, st := VerifySplunk("splunk", splunk)
			Expect(valid).To(BeTrue())
			Expect(map[string]status.Condition{"splunk": st}).To(HaveCondition(loggingv1.ConditionDegraded, true, loggingv1.ReasonUnused, "Warning: Support for 'fields' is not implemented and deprecated for output \"splunk\""))
		})
	})

	Context("when validating secret for Splunk", func() {

		It("should fail for secret that are missing hecToken", func() {
			secret := runtime.NewSecret("namespace", "mysec", map[string][]byte{
				"foo": {'b', 'a', 'r'},
			})
			Expect(VerifySecretKeysForSplunk(&loggingv1.OutputSpec{}, loggingv1.NamedConditions{}, secret)).To(BeFalse())
		})

		It("should fail for secret that has empty hecToken", func() {
			secret := runtime.NewSecret("namespace", "mysec", map[string][]byte{
				constants.SplunkHECTokenKey: nil,
			})
			Expect(VerifySecretKeysForSplunk(&loggingv1.OutputSpec{}, loggingv1.NamedConditions{}, secret)).To(BeFalse())
		})

		It("should pass for secrets with not empty hecToken", func() {
			secret := runtime.NewSecret("namespace", "mysec", map[string][]byte{
				constants.SplunkHECTokenKey: {'k', 'e', 'y'},
			})
			Expect(VerifySecretKeysForSplunk(&loggingv1.OutputSpec{}, loggingv1.NamedConditions{}, secret)).To(BeTrue())
		})
	})
})

func TestValidateIndexKeyMatch(t *testing.T) {
	inputs := map[string]bool{
		"kubernetes.labels.test-logging":     false,
		"kubernetes.labels.test:logging":     false,
		"kubernetes.labels.test/logging":     false,
		"kubernetes.labels.test_logging":     true,
		"kubernetes.labels.\"test-logging\"": true,
		"kubernetes.labels.\"test/logging\"": true,
		"kubernetes.labels.\"test:logging\"": true,
		"example.string.test":                true,
	}

	for key, must := range inputs {
		match := isIndexKeyMatch(key)
		assert.Equal(t, match, must, "isIndexKeyMatch('%s') = %v, must be: %v", key, match, must)
	}
}
