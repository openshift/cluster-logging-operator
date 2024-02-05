package inputs

import (
	"fmt"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

// Verify and set status.Inputs conditions
func Verify(inputs []loggingv1.InputSpec, status *loggingv1.ClusterLogForwarderStatus, extras map[string]bool) {
	// Collect input conditions
	status.Inputs = loggingv1.NamedConditions{}

	// Check input names
	for i, input := range inputs {
		i, input := i, input // Don't bind range variables.
		badInput := func(format string, args ...interface{}) {
			if input.Name == "" {
				input.Name = fmt.Sprintf("input_%v_", i)
			}
			status.Inputs.Set(input.Name, conditions.CondInvalid(format, args...))
		}

		validPort := func(port int32) bool {
			return port > 1023 && port < 65536
		}
		isReceiverSpecDefined := func(input loggingv1.InputSpec) bool {
			return input.Receiver != nil && input.Receiver.ReceiverTypeSpec != nil
		}

		switch {
		case input.Name == "":
			badInput("input must have a name")
		case loggingv1.ReservedInputNames.Has(input.Name):
			if !extras[fmt.Sprintf("%s%s", constants.MigrateInputPrefix, input.Name)] {
				badInput("input name %q is reserved", input.Name)
			}
		case len(status.Inputs[input.Name]) > 0:
			badInput("duplicate name: %q", input.Name)
		// Check if inputspec has application, infrastructure, audit or receiver specs
		case !hasOneType(input):
			badInput("inputspec must define one and only one of: application, infrastructure, audit or receiver")
		case !validApplication(input, status, extras):
		case !validInfrastructure(input, status, extras):
		case !validAudit(input, status, extras):
		case input.Receiver != nil && !extras[constants.VectorName]:
			badInput("ReceiverSpecs are only supported for the vector log collector")
		case input.Receiver != nil && input.Receiver.Type != loggingv1.ReceiverTypeHttp && input.Receiver.Type != loggingv1.ReceiverTypeSyslog:
			badInput("invalid Type specified for receiver")
		case isReceiverSpecDefined(input) && input.Receiver.IsHttpReceiver() && input.Receiver.Syslog != nil:
			badInput("mismatched Type specified for receiver, specified HTTP and have Syslog")
		case isReceiverSpecDefined(input) && input.Receiver.IsSyslogReceiver() && input.Receiver.HTTP != nil:
			badInput("mismatched Type specified for receiver, specified Syslog and have HTTP")
		case isReceiverSpecDefined(input) && input.Receiver.IsHttpReceiver() && !validPort(input.Receiver.HTTP.Port):
			badInput("invalid port specified for HTTP receiver")
		case isReceiverSpecDefined(input) && input.Receiver.IsSyslogReceiver() && !validPort(input.Receiver.Syslog.Port):
			badInput("invalid port specified for Syslog receiver")
		case isReceiverSpecDefined(input) && input.Receiver.IsHttpReceiver() && input.Receiver.HTTP.Format != loggingv1.FormatKubeAPIAudit:
			badInput("invalid format specified for HTTP receiver")
		default:
			status.Inputs.Set(input.Name, conditions.CondReady)
		}
	}
}

func hasOneType(spec loggingv1.InputSpec) bool {
	totTypes := 0
	if spec.Application != nil {
		totTypes += 1
	}
	if spec.Infrastructure != nil {
		totTypes += 1
	}
	if spec.Audit != nil {
		totTypes += 1
	}
	if spec.Receiver != nil {
		totTypes += 1
	}
	return totTypes == 1
}

var (
	conditionInfraValidationSourcesFailure = loggingv1.NewCondition(loggingv1.ValidationCondition,
		corev1.ConditionTrue,
		loggingv1.ValidationFailureReason,
		"infrastructure inputs must define at least one valid source: %s", strings.Join(loggingv1.InfrastructureSources.List(), ","))
	conditionAuditValidationSourcesFailure = loggingv1.NewCondition(loggingv1.ValidationCondition,
		corev1.ConditionTrue,
		loggingv1.ValidationFailureReason,
		"infrastructure inputs must define at least one valid source: %s", strings.Join(loggingv1.AuditSources.List(), ","))
)

func validInfrastructure(spec loggingv1.InputSpec, status *loggingv1.ClusterLogForwarderStatus, extras map[string]bool) bool {
	if spec.Infrastructure != nil {
		switch {
		case len(spec.Infrastructure.Sources) == 0:
			status.Inputs.Set(spec.Name, conditionInfraValidationSourcesFailure)
		case !sets.NewString(spec.Infrastructure.Sources...).SubsetOf(&loggingv1.InfrastructureSources.Set):
			status.Inputs.Set(spec.Name, conditionInfraValidationSourcesFailure)
		}
	}
	return len(status.Inputs[spec.Name]) == 0
}
func validAudit(spec loggingv1.InputSpec, status *loggingv1.ClusterLogForwarderStatus, extras map[string]bool) bool {
	if spec.Audit != nil {
		switch {
		case len(spec.Audit.Sources) == 0:
			status.Inputs.Set(spec.Name, conditionAuditValidationSourcesFailure)
		case !sets.NewString(spec.Audit.Sources...).SubsetOf(&loggingv1.AuditSources.Set):
			status.Inputs.Set(spec.Name, conditionAuditValidationSourcesFailure)
		}
	}
	return len(status.Inputs[spec.Name]) == 0
}
