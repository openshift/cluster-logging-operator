package inputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("#ValidateReceiver", func() {

	Context("when validating a receiver input", func() {
		var (
			spec obs.InputSpec
		)
		BeforeEach(func() {
			spec = obs.InputSpec{
				Name:     "myreceiver",
				Type:     obs.InputTypeReceiver,
				Receiver: &obs.ReceiverSpec{},
			}
		})
		It("should skip the validation when not a receiver type", func() {
			spec.Type = obs.InputTypeApplication
			conds := ValidateReceiver(spec)
			Expect(conds).To(BeEmpty())
		})
		It("should fail when a receiver type but has no receiver spec", func() {
			spec.Receiver = nil
			conds := ValidateReceiver(spec)
			Expect(conds).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonValidationFailure, "myreceiver has nil receiver spec"))
		})
		It("should fail when receiver type is HTTP but does not have http receiver spec", func() {
			spec.Receiver.Type = obs.ReceiverTypeHTTP
			conds := ValidateReceiver(spec)
			Expect(conds).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonValidationFailure, "myreceiver has nil HTTP receiver spec"))
		})
		It("should fail when receiver type is HTTP but does not specify an incoming format", func() {
			spec.Receiver.Type = obs.ReceiverTypeHTTP
			spec.Receiver.HTTP = &obs.HTTPReceiver{}
			conds := ValidateReceiver(spec)
			Expect(conds).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonValidationFailure, "myreceiver does not specify a format"))
		})
		It("should pass for a valid HTTP receiver spec", func() {
			spec.Receiver.Type = obs.ReceiverTypeHTTP
			spec.Receiver.HTTP = &obs.HTTPReceiver{
				Format: obs.HTTPReceiverFormatKubeAPIAudit,
			}
			conds := ValidateReceiver(spec)
			Expect(conds).To(BeEmpty())
		})
		It("should pass for a valid syslog receiver spec", func() {
			spec.Receiver.Type = obs.ReceiverTypeSyslog
			conds := ValidateReceiver(spec)
			Expect(conds).To(BeEmpty())
		})
	})
})
