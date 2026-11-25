package component

// Provider is an internal forwarder component that is
// provided as input to other components
type Provider interface {
	// InputIDs are the ids of config elements to use as input to other components
	InputIDs() []string
}

// Consumer is an internal forwarder component that receives input from another component (e.g. transform, sink)
type Consumer interface {
	AddInputFrom(n Provider)
}
