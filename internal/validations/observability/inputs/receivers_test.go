package inputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/initialize"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("#ValidateReceiver", func() {

	Context("when validating a receiver input", func() {
		var (
			spec               obs.InputSpec
			secrets            = map[string]*corev1.Secret{}
			configMaps         = map[string]*corev1.ConfigMap{}
			expConditionTypeRE = obs.ConditionTypeValidInputPrefix + "-.*"
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
			conds := ValidateReceiver(spec, secrets, configMaps, utils.NoOptions)
			Expect(conds).To(BeEmpty())
		})
		It("should fail when a receiver type but has no receiver spec", func() {
			spec.Receiver = nil
			conds := ValidateReceiver(spec, secrets, configMaps, utils.NoOptions)
			Expect(conds).To(HaveCondition(expConditionTypeRE, false, obs.ReasonMissingSpec, "myreceiver has nil receiver spec"))
		})
		It("should fail when receiver type is HTTP but does not have http receiver spec", func() {
			spec.Receiver.Type = obs.ReceiverTypeHTTP
			conds := ValidateReceiver(spec, secrets, configMaps, utils.NoOptions)
			Expect(conds).To(HaveCondition(expConditionTypeRE, false, obs.ReasonMissingSpec, "myreceiver has nil HTTP receiver spec"))
		})
		It("should fail when receiver type is HTTP but does not specify an incoming format", func() {
			spec.Receiver.Type = obs.ReceiverTypeHTTP
			spec.Receiver.HTTP = &obs.HTTPReceiver{}
			conds := ValidateReceiver(spec, secrets, configMaps, utils.NoOptions)
			Expect(conds).To(HaveCondition(expConditionTypeRE, false, obs.ReasonValidationFailure, "myreceiver does not specify a format"))
		})
		It("should pass for a valid HTTP receiver spec", func() {
			spec.Receiver.Type = obs.ReceiverTypeHTTP
			spec.Receiver.HTTP = &obs.HTTPReceiver{
				Format: obs.HTTPReceiverFormatKubeAPIAudit,
			}
			conds := ValidateReceiver(spec, secrets, configMaps, utils.NoOptions)
			Expect(conds).To(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `input.*is valid`))
		})
		It("should pass for a valid syslog receiver spec", func() {
			spec.Receiver.Type = obs.ReceiverTypeSyslog
			conds := ValidateReceiver(spec, secrets, configMaps, utils.NoOptions)
			Expect(conds).To(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, `input.*is valid`))
		})
		It("should fail validate secrets if spec'd", func() {
			spec.Receiver.Type = obs.ReceiverTypeSyslog
			spec.Receiver.TLS = &obs.InputTLSSpec{
				CA: &obs.ValueReference{
					Key:           "foo",
					ConfigMapName: "immissing",
				},
			}
			conds := ValidateReceiver(spec, secrets, configMaps, utils.NoOptions)
			Expect(conds).To(Not(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, "")))
		})
		Context("for secrets provied by the cert signing service", func() {
			It("should skip validation", func() {

				context := utils.Options{
					initialize.GeneratedSecrets: []*corev1.Secret{
						runtime.NewSecret("", "immissing", map[string][]byte{
							"foo":                   {},
							constants.ClientCertKey: {},
						}),
					},
				}

				spec.Receiver.Type = obs.ReceiverTypeSyslog
				spec.Receiver.TLS = &obs.InputTLSSpec{
					CA: &obs.ValueReference{
						Key:        "foo",
						SecretName: "immissing",
					},
				}
				conds := ValidateReceiver(spec, secrets, configMaps, context)
				Expect(conds).To(HaveCondition(expConditionTypeRE, true, obs.ReasonValidationSuccess, ""))
			})

		})
	})
})
