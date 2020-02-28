package outputs

// Syslog provides optional extra properties for `type: syslog`
type Syslog struct {
	// Severity to set on outgoing syslog records.
	//
	// The string value may contain an integer or one of these names:
	//     emergency alert critical error warning notice informational debug
	//
	// +optional
	Severity string `json:"severity,omitempty"`

	// Facility to set on outgoing syslog records.
	//
	// The string value may contain an integer or one of these names:
	//     kernel user mail daemon auth syslog lpr news
	//     uucp cron authpriv ftp ntp security console cron
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
