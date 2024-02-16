package clusterlogforwarder

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func MigrateInputs(namespace, name string, spec loggingv1.ClusterLogForwarderSpec, logStore *loggingv1.LogStoreSpec, extras map[string]bool, logstoreSecretName, saTokenSecret string) (loggingv1.ClusterLogForwarderSpec, map[string]bool, []loggingv1.Condition) {
	for i, input := range spec.Inputs {
		if input.Receiver != nil && input.Receiver.ReceiverTypeSpec != nil {
			if input.Receiver.HTTP != nil && input.Receiver.Type == "" {
				input.Receiver.Type = loggingv1.ReceiverTypeHttp
				spec.Inputs[i] = input
			}
			if input.Receiver.Syslog != nil && input.Receiver.Type == "" {
				input.Receiver.Type = loggingv1.ReceiverTypeSyslog
				spec.Inputs[i] = input
			}
		}
	}

	inputs := map[string]loggingv1.InputSpec{}
	for _, p := range spec.Pipelines {
		for _, i := range p.InputRefs {
			switch i {
			case loggingv1.InputNameApplication:
				inputs[loggingv1.InputNameApplication] = loggingv1.InputSpec{
					Name:        loggingv1.InputNameApplication,
					Application: &loggingv1.Application{},
				}
				extras[constants.MigrateInputApplication] = true
			case loggingv1.InputNameInfrastructure:
				inputs[loggingv1.InputNameInfrastructure] = loggingv1.InputSpec{
					Name: loggingv1.InputNameInfrastructure,
					Infrastructure: &loggingv1.Infrastructure{
						Sources: loggingv1.InfrastructureSources.List(),
					},
				}
				extras[constants.MigrateInputInfrastructure] = true
			case loggingv1.InputNameAudit:
				inputs[loggingv1.InputNameAudit] = loggingv1.InputSpec{
					Name: loggingv1.InputNameAudit,
					Audit: &loggingv1.Audit{
						Sources: loggingv1.AuditSources.List(),
					},
				}
				extras[constants.MigrateInputAudit] = true
			}
		}
	}
	for _, i := range inputs {
		spec.Inputs = append(spec.Inputs, i)
	}
	return spec, extras, nil
}
