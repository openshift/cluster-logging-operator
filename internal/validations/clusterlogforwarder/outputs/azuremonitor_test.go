package outputs

//TODO: MIGRATE OR REMOVE ME
//import (
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
//	"github.com/openshift/cluster-logging-operator/internal/constants"
//	"github.com/openshift/cluster-logging-operator/internal/runtime"
//	"github.com/openshift/cluster-logging-operator/internal/status"
//	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
//	. "github.com/openshift/cluster-logging-operator/test/matchers"
//)
//
//var _ = Describe("Validate AzureMonitor:", func() {
//
//	Context("when validating Output Type", func() {
//
//		It("should be valid", func() {
//			azureMonitor := &loggingv1.AzureMonitor{
//				CustomerId: "customer",
//				LogType:    "application",
//			}
//			valid, st := VerifyAzureMonitorLog("azureMonitor", azureMonitor)
//			Expect(valid).To(BeTrue())
//			Expect(map[string]status.Condition{"azureMonitor": st}).To(HaveCondition("Ready", true, "", ""))
//		})
//
//		It("should fail if no logType", func() {
//			azureMonitor := &loggingv1.AzureMonitor{
//				CustomerId: "customer",
//			}
//			valid, st := VerifyAzureMonitorLog("azureMonitor", azureMonitor)
//			Expect(valid).To(BeFalse())
//			Expect(map[string]status.Condition{"azureMonitor": st}).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"azureMonitor\": LogType must be set."))
//		})
//
//		It("should fail if logType bad format: not allowed symbols", func() {
//			azureMonitor := &loggingv1.AzureMonitor{
//				CustomerId: "customer",
//				LogType:    "bad-format",
//			}
//			valid, st := VerifyAzureMonitorLog("azureMonitor", azureMonitor)
//			Expect(valid).To(BeFalse())
//			Expect(map[string]status.Condition{"azureMonitor": st}).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"azureMonitor\": LogType names must start with a letter/number, contain only letters/numbers/underscores \\(_\\), and be between 1-100 characters."))
//		})
//
//		It("should fail if logType bad format: longer than 100", func() {
//			azureMonitor := &loggingv1.AzureMonitor{
//				CustomerId: "customer",
//				LogType:    string(rand.Word(101)),
//			}
//			valid, st := VerifyAzureMonitorLog("azureMonitor", azureMonitor)
//			Expect(valid).To(BeFalse())
//			Expect(map[string]status.Condition{"azureMonitor": st}).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"azureMonitor\": LogType names must start with a letter/number, contain only letters/numbers/underscores \\(_\\), and be between 1-100 characters."))
//		})
//
//		It("should fail if logType bad format: begins on '_'", func() {
//			azureMonitor := &loggingv1.AzureMonitor{
//				CustomerId: "customer",
//				LogType:    "_testType",
//			}
//			valid, st := VerifyAzureMonitorLog("azureMonitor", azureMonitor)
//			Expect(valid).To(BeFalse())
//			Expect(map[string]status.Condition{"azureMonitor": st}).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"azureMonitor\": LogType names must start with a letter/number, contain only letters/numbers/underscores \\(_\\), and be between 1-100 characters."))
//		})
//
//		It("should fail if no customerId", func() {
//			azureMonitor := &loggingv1.AzureMonitor{
//				LogType: "application",
//			}
//			valid, st := VerifyAzureMonitorLog("azureMonitor", azureMonitor)
//			Expect(valid).To(BeFalse())
//			Expect(map[string]status.Condition{"azureMonitor": st}).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"azureMonitor\": CustomerId must be set."))
//		})
//	})
//
//	Context("when validating secret for Azure Monitor", func() {
//
//		It("should fail for secret that are missing shared_key", func() {
//			secret := runtime.NewSecret("namespace", "mysec", map[string][]byte{
//				"foo": {'b', 'a', 'r'},
//			})
//			Expect(VerifySharedKeysForAzure(&loggingv1.OutputSpec{}, loggingv1.NamedConditions{}, secret)).To(BeFalse())
//		})
//
//		It("should fail for secret that has empty shared_key", func() {
//			secret := runtime.NewSecret("namespace", "mysec", map[string][]byte{
//				constants.SharedKey: nil,
//			})
//			Expect(VerifySharedKeysForAzure(&loggingv1.OutputSpec{}, loggingv1.NamedConditions{}, secret)).To(BeFalse())
//		})
//
//		It("should pass for secrets with shared_key", func() {
//			secret := runtime.NewSecret("namespace", "mysec", map[string][]byte{
//				constants.SharedKey: {'k', 'e', 'y'},
//			})
//			Expect(VerifySharedKeysForAzure(&loggingv1.OutputSpec{}, loggingv1.NamedConditions{}, secret)).To(BeTrue())
//		})
//	})
//})
