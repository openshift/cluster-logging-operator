package inputs

import (
	"fmt"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
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
		isAReceiver := func(input loggingv1.InputSpec) bool {
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
		case input.Application == nil && input.Infrastructure == nil && input.Audit == nil && input.Receiver == nil:
			badInput("inputspec must define one or more of application, infrastructure, audit or receiver")
		case input.HasPolicy() && input.Application.ContainerLimit != nil && input.Application.GroupLimit != nil:
			badInput("inputspec must define only one of container or group limit")
		case input.HasPolicy() && input.GetMaxRecordsPerSecond() < 0:
			badInput("inputspec cannot have a negative limit threshold")
		case input.Receiver != nil && !extras[constants.VectorName]:
			badInput("ReceiverSpecs are only supported for the vector log collector")
		case input.Receiver != nil && input.Receiver.ReceiverTypeSpec == nil:
			badInput("invalid ReceiverTypeSpec specified for receiver")
		case input.Receiver != nil && input.Receiver.Type != loggingv1.ReceiverTypeHttp && input.Receiver.Type != loggingv1.ReceiverTypeSyslog:
			badInput("invalid Type specified for receiver")
		case input.Receiver != nil && input.Receiver.Type == loggingv1.ReceiverTypeHttp && input.Receiver.Syslog != nil:
			badInput("mismatched Type specified for receiver, specified HTTP and have Syslog")
		case input.Receiver != nil && input.Receiver.Type == loggingv1.ReceiverTypeSyslog && input.Receiver.HTTP != nil:
			badInput("mismatched Type specified for receiver, specified Syslog and have HTTP")
		case isAReceiver(input) && input.Receiver.HTTP == nil && input.Receiver.Syslog == nil:
			badInput("ReceiverSpec must define either HTTP or Syslog receiver")
		case loggingv1.IsHttpReceiver(&input) && !validPort(input.Receiver.HTTP.Port):
			badInput("invalid port specified for HTTP receiver")
		case loggingv1.IsSyslogReceiver(&input) && !validPort(input.Receiver.Syslog.Port):
			badInput("invalid port specified for Syslog receiver")
		case loggingv1.IsHttpReceiver(&input) && input.Receiver.HTTP.Format != loggingv1.FormatKubeAPIAudit:
			badInput("invalid format specified for HTTP receiver")
		case loggingv1.IsSyslogReceiver(&input) && input.Receiver.Syslog.Protocol != "tcp" && input.Receiver.Syslog.Protocol != "udp":
			badInput("invalid protocol specified for Syslog receiver")
		default:
			status.Inputs.Set(input.Name, conditions.CondReady)
		}
	}
}
