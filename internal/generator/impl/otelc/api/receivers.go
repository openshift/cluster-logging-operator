package api

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"gopkg.in/yaml.v3"
)

// Receivers is a set of receivers by their id.
// In OpenTelemetry Collector configuration, receiver IDs follow the pattern "type" or "type/name"
// where the type determines the receiver implementation.
type Receivers map[string]types.Receiver

func (receiverMap *Receivers) Add(receiver types.Receiver) {
	(*receiverMap)[receiver.ID()] = receiver
}

func (receiverMap *Receivers) UnmarshalYAML(value *yaml.Node) error {
	if *receiverMap == nil {
		*receiverMap = make(Receivers)
	}
	return unmarshalComponentMap(value, "receivers", decodeReceiver, func(id string, receiver types.Receiver) {
		(*receiverMap)[id] = receiver
	})
}

// decodeReceiver decodes a YAML node into a specific receiver type
func decodeReceiver(receiverType, receiverID string, node *yaml.Node) (types.Receiver, error) {
	switch types.ReceiverType(receiverType) {
	case types.ReceiverTypeFileLog:
		var r receivers.FileLog
		if err := node.Decode(&r); err != nil {
			return nil, fmt.Errorf("failed to unmarshal file_log receiver %s: %w", receiverID, err)
		}
		return &r, nil
	default:
		return nil, fmt.Errorf("unknown receiver type %s for receiver %s", receiverType, receiverID)
	}
}
