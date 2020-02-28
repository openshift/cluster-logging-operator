package v1

import (
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"
	corev1 "k8s.io/api/core/v1"
)

// Output defines a destination for log messages.
type Output struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +required
	Name string `json:"name"`

	// OutputType is inlined with a required `type` and optional extra configuration.
	OutputType `json:",inline"`

	// URL to send log messages to.
	//
	// Must be an absolute URL, with a scheme. Valid URL schemes depend on `type`.
	// Special schemes 'tcp', 'udp' and 'tls' are used for output types that don't
	// define their own URL scheme.  Example:
	//
	//     { type: syslog, url: tls://syslog.example.com:1234 }
	//
	// TLS with server authentication is enabled by the URL scheme, for
	// example 'tls' or 'https'.  See `secretRef` for TLS client authentication.
	//
	// +required
	URL string `json:"url"`

	// SecretRef refers to a `secret` object for secure communication.
	//
	// Client-authenticated TLS is enabled if the secret contains keys
	// `tls.crt`, `tls.key` and `ca.crt`. Output types with password
	// authentication will use keys `password` and `username`, not
	// the exposed 'username@password' part of the `url`.
	//
	// For a `ClusterLogForwarder`, `secretRef.namespace` defaults to the
	// `cluster-logging` namespace where secrets are normally stored.
	//
	// +optional
	SecretRef corev1.SecretReference `json:"secretRef,omitempty"`

	// Insecure must be true for intentionally insecure outputs.
	// Has no function other than a marker to help avoid configuration mistakes.
	//
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// FIXME(alanconway) remove this unless it can be supported in initial release.
	//
	// Reconnect configures how the output handles connection failures.
	// Auto-reconnect is enabled by default.
	//
	// +optional
	// Reconnect *Reconnect `json:"reconnect,omitempty"`
}

// +kubebuilder:validation:Enum=Unreliable;Retry
type Reliability string

const (
	// Unreliable may drop data after a reconnect (at-most-once).
	Unreliable = "Unreliable"

	// Resend "in doubt" data after a reconnect. May cause duplicates (at-least-once).
	// May enable buffering, blocking and/or acknowledgment features of the output type.
	Resend = "Resend"
)

// Reconnect configures reconnect behavior after a disconnect.
type Reconnect struct {
	// FirstDelayMilliseconds is the time to wait after a disconnect before
	// the first reconnect attempt. If reconnect fails, the delay is doubled
	// on each subsequent attempt. The default is determined by the output type.
	//
	// +optional
	FirstDelayMilliseconds int64 `json:"firstDelayMilliseconds,omitempty"`

	// MaxDelaySeconds is the maximum delay between failed re-connect
	// attempts, and also the maximum time to wait for an unresponsive
	// connection attempt. The default is determined by the output type.
	//
	// +optional
	MaxDelayMilliseconds int64 `json:"maxDelayMilliseconds,omitempty"`

	// Reliability policy for data delivery after a re-connect.  This is
	// simple short-hand for configuring the output to a given level of
	// reliability.  The exact meaning depends on the output `type`.  The
	// default is determined by the output type.
	//
	// +optional
	Reliability Reliability `json:"reliability,omitempty"`
}

// OutputType is a union of supported output types.
//
// It always includes the required a `type` string and optionally a struct field
// with type-specific configuration.
type OutputType struct {
	// Type of output plugin, for example 'syslog'
	//
	// +unionDiscriminator
	// +required
	Type string `json:"type"`

	// Syslog optional extra syslog-specific properties.
	//
	// +optional
	Syslog *outputs.Syslog `json:"syslog, omitempty"`

	// FIXME(alanconway) how to doc?
}
