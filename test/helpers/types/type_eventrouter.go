package types

import (
	v1 "k8s.io/api/core/v1"
)

// EventRouterLog is a Viaq wrappered log from the eventrouter of
// kubernetes core events
type EventRouterLog struct {
	ViaQCommon `json:",inline,omitempty"`

	// The Kubernetes-specific metadata
	Kubernetes KubernetesWithEvent `json:"kubernetes,omitempty"`

	// OldEvent is a core KubernetesEvent that was replaced by
	// kubernetes.event
	OldEvent *v1.Event `json:"old_event,omitempty"`
}

type KubernetesWithEvent struct {
	Kubernetes `json:",inline,omitempty"`

	// Event is the core KubernetesEvent
	Event ViaqEventRouterEvent `json:"event,omitempty"`
}

type ViaqEventRouterEvent struct {
	v1.Event `json:",inline,omitempty"`

	//Verb is indicates if event was created or updated
	Verb string `json:"verb,omitempty"`
}

// EventData encodes an eventrouter event and previous event, with a verb for
// whether the event is created or updated. This is the format as collected
// from the eventrouter
type EventData struct {
	Verb     string    `json:"verb"`
	Event    *v1.Event `json:"event"`
	OldEvent *v1.Event `json:"old_event,omitempty"`
}
