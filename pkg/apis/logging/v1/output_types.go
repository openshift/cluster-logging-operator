package v1

import (
	"reflect"
	"strings"
)

// Default log store output name.
const OutputNameDefault = "default"

func IsReservedOutputName(s string) bool { return s == OutputNameDefault }

// Output type constants, must match JSON tags of OutputTypeSpec fields.
const (
	OutputTypeElasticsearch = "elasticsearch"
	OutputTypeFluentForward = "fluentForward"
	OutputTypeSyslog        = "syslog"
)

// Output defines a destination for log messages.
type OutputSpec struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +required
	Name string `json:"name"`

	// Type of output plugin, for example 'syslog'
	//
	// +required
	Type string `json:"type"`

	// URL to send log messages to.
	//
	// Must be an absolute URL, with a scheme. Valid URL schemes depend on `type`.
	// Special schemes 'tcp', 'udp' and 'tls' are used for output types that don't
	// define their own URL scheme.  Example:
	//
	//     { type: syslog, url: tls://syslog.example.com:1234 }
	//
	// TLS with server authentication is enabled by the URL scheme, for
	// example 'tls' or 'https'.  See `secret` for TLS client authentication.
	//
	// +optional
	URL string `json:"url"`

	// OutputTypeSpec provides optional extra configuration that is specific to the
	// output `type`
	//
	// +optional
	OutputTypeSpec `json:",inline"`

	// Secret for secure communication.
	// Secrets must be stored in the namespace containing the cluster logging operator.
	//
	// Client-authenticated TLS is enabled if the secret contains keys `tls.crt`,
	// `tls.key` and `ca.crt`. Output types with password authentication will use
	// keys `password` and `username`, not the exposed 'username@password' part of
	// the `url`.
	//
	// +optional
	Secret *OutputSecretSpec `json:"secret,omitempty"`

	// Insecure must be true for intentionally insecure outputs.
	// Has no function other than a marker to help avoid configuration mistakes.
	//
	// +optional
	Insecure bool `json:"insecure,omitempty"`
}

// OutputSecretSpec is a secret reference containing name only, no namespace.
type OutputSecretSpec struct {
	// Name of a secret in the namespace configured for log forwarder secrets.
	//
	// +required
	Name string `json:"name"`
}

// OutputTypeSpec is a union of optional additional configuration specific to an
// output type.
type OutputTypeSpec struct {
	// +optional
	Syslog *Syslog `json:"syslog,omitempty"`
	// +optional
	FluentForward *FluentForward `json:"fluentForward,omitempty"`
	// +optional
	Elasticsearch *Elasticsearch `json:"elasticsearch,omitempty"`
}

var otsType = reflect.TypeOf(OutputTypeSpec{})

func IsOutputTypeName(s string) bool {
	_, ok := otsType.FieldByName(strings.Title(s))
	return ok
}

// Syslog provides optional extra properties for output type `syslog`
type Syslog struct {
	// Severity to set on outgoing syslog records.
	//
	// Severity values are defined in https://tools.ietf.org/html/rfc5424#section-6.2.1
	// The value can be a decimal integer or one of these case-insensitive keywords:
	//
	//     Emergency Alert Critical Error Warning Notice Informational Debug
	//
	// +optional
	Severity string `json:"severity,omitempty"`

	// Facility to set on outgoing syslog records.
	//
	// Facility values are defined in https://tools.ietf.org/html/rfc5424#section-6.2.1.
	// The value can be a decimal integer. Facility keywords are not standardized,
	// this API recognizes at least the following case-insensitive keywords
	// (defined by https://en.wikipedia.org/wiki/Syslog#Facility_Levels):
	//
	//     kernel user mail daemon auth syslog lpr news
	//     uucp cron authpriv ftp ntp security console solaris-cron
	//     local0 local1 local2 local3 local4 local5 local6 local7
	//
	// +optional
	Facility string `json:"facility,omitempty"`

	// TrimPrefix is a prefix to trim from the tag.
	//
	// +optional
	TrimPrefix string `json:"trimPrefix,omitempty"`

	// TagKey specifies a record field  to  use as tag.
	//
	// +optional
	TagKey string `json:"tagKey,omitempty"`

	// PayloadKey specifies record field to use as payload.
	//
	// +optional
	PayloadKey string `json:"payloadKey,omitempty"`
}

// Placeholders for configuration of other types

type FluentForward struct{}
type Elasticsearch struct{}
