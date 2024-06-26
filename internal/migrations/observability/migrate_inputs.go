package observability

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MigrateInputs creates instances of inputSpec when a pipeline references one of the reserved input names (e.g. application)
// This function is destructive and replaces any inputs that use the reserved input names
func MigrateInputs(spec obs.ClusterLogForwarder, options utils.Options) (obs.ClusterLogForwarder, []metav1.Condition) {
	inputs := internalobs.Inputs(spec.Spec.Inputs).Map()
	for _, p := range spec.Spec.Pipelines {
		for _, i := range p.InputRefs {
			switch i {
			case string(obs.InputTypeApplication):
				inputs[string(obs.InputTypeApplication)] = obs.InputSpec{
					Type:        obs.InputTypeApplication,
					Name:        string(obs.InputTypeApplication),
					Application: &obs.Application{},
				}
			case string(obs.InputTypeInfrastructure):
				inputs[string(obs.InputTypeInfrastructure)] = obs.InputSpec{
					Type: obs.InputTypeInfrastructure,
					Name: string(obs.InputTypeInfrastructure),
					Infrastructure: &obs.Infrastructure{
						Sources: obs.InfrastructureSources,
					},
				}
			case string(obs.InputTypeAudit):
				inputs[string(obs.InputTypeAudit)] = obs.InputSpec{
					Type: obs.InputTypeAudit,
					Name: string(obs.InputTypeAudit),
					Audit: &obs.Audit{
						Sources: obs.AuditSources,
					},
				}
			}
		}
	}
	spec.Spec.Inputs = []obs.InputSpec{}
	for _, i := range inputs {
		i = migrateInputReceiver(i, spec.Name, options)
		spec.Spec.Inputs = append(spec.Spec.Inputs, i)
	}
	return spec, nil
}

func migrateInputReceiver(spec obs.InputSpec, forwarderName string, options utils.Options) obs.InputSpec {
	if spec.Type != obs.InputTypeReceiver {
		return spec
	}
	if spec.Receiver != nil && spec.Receiver.TLS != nil {
		return spec
	}
	secretName := fmt.Sprintf("%s-%s", forwarderName, spec.Name)
	spec.Receiver.TLS = &obs.InputTLSSpec{
		Key: &obs.SecretKey{
			Key: constants.ClientPrivateKey,
			Secret: &corev1.LocalObjectReference{
				Name: secretName,
			},
		},
		Certificate: &obs.ConfigMapOrSecretKey{
			Key: constants.ClientCertKey,
			Secret: &corev1.LocalObjectReference{
				Name: secretName,
			},
		},
	}
	secrets := []*corev1.Secret{
		runtime.NewSecret("", secretName, map[string][]byte{
			constants.ClientPrivateKey: {},
			constants.ClientCertKey:    {},
		}),
	}
	utils.Update(options, GeneratedSecrets, secrets, func(existing []*corev1.Secret) []*corev1.Secret {
		return append(existing, secrets...)
	})
	return spec
}
