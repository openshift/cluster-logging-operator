// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1

// Output type constants, must match JSON tags of OutputTypeSpec fields.
const (
	OutputTypeElasticsearch  = "elasticsearch"
	OutputTypeFluentdForward = "fluentdForward"
	OutputTypeSyslog         = "syslog"
	OutputTypeKafka          = "kafka"
)

// NOTE: The Enum validation on OutputSpec.Type must be updated if the list of
// known types changes.

// OutputTypeSpec is a union of optional additional configuration specific to an
// output type. The fields of this struct define the set of known output types.
type OutputTypeSpec struct {
	// +optional
	Syslog *Syslog `json:"syslog,omitempty"`
	// +optional
	FluentdForward *FluentdForward `json:"fluentdForward,omitempty"`
	// +optional
	Elasticsearch *Elasticsearch `json:"elasticsearch,omitempty"`
	// +optional
	Kafka *Kafka `json:"kafka,omitempty"`
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

	// Tag specifies a record field to use as tag.
	//
	// +optional
	Tag string `json:"tag,omitempty"`

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
	// +kubebuilder:validation:Enum:=RFC3164;RFC5424
	// +kubebuilder:default:=RFC5424
	// +optional
	RFC string `json:"rfc,omitempty"`

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

// Kafka provides optional extra properties for `type: kafka`
type Kafka struct {
	// Topic specifies the target topic to send logs to.
	//
	// +optional
	Topic string `json:"topic,omitempty"`

	// Brokers specifies the list of brokers
	// to register in addition to the main output URL
	// on initial connect to enhance reliability.
	//
	// +optional
	Brokers []string `json:"brokers,omitempty"`
}

// Placeholders for configuration of other types

type FluentdForward struct{}
type Elasticsearch struct{}
