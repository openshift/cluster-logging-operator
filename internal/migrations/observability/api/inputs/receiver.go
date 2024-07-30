package inputs

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapReceiverInput(loggingReceiver *logging.ReceiverSpec) *obs.ReceiverSpec {
	obsReceiver := &obs.ReceiverSpec{}
	// HTTP Receiver
	if loggingReceiver.HTTP != nil {
		obsReceiver.Type = obs.ReceiverTypeHTTP
		obsReceiver.Port = loggingReceiver.HTTP.Port
		obsReceiver.HTTP = &obs.HTTPReceiver{
			Format: obs.HTTPReceiverFormat(loggingReceiver.HTTP.Format),
		}

		// Syslog Receiver
	} else if loggingReceiver.Syslog != nil {
		obsReceiver.Type = obs.ReceiverTypeSyslog
		obsReceiver.Port = loggingReceiver.Syslog.Port
	}
	return obsReceiver
}
