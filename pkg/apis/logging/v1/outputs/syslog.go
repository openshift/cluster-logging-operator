package outputs

// Syslog provides optional extra properties for `type: syslog`
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

	// Rfc specifies the rfc to be used for sending syslog
	//
	// Rfc values can be one of:
	//  - RFC3164 (https://tools.ietf.org/html/rfc3164)
	//  - RFC5424 (https://tools.ietf.org/html/rfc5424)
	//
	// If unspecified, RFC5424 will be assumed.
	//
	// +kubebuilder:validation:Enum:=RFC3164;RFC5425
	// +optional
	RFC string `json:"RFC,omitempty"`

	// AppName is APP-NAME part of the syslog-msg header
	//
	// AppName needs to be specified if using rfc5424
	//
	// +optional
	AppName string `json:"appName,omitempty"`

	// ProcID is PROCID part of the syslog-msg header
	//
	// ProcID needs to be specified if using rfc5424
	//
	// +optional
	ProcID string `json:"procID,omitempty"`

	// MsgID is MSGID part of the syslog-msg header
	//
	// MsgID needs to be specified if using rfc5424
	//
	// +optional
	MsgID string `json:"msgID,omitempty"`
}
