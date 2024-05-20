/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"time"

	openshiftv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// OutputType is used to define the type of output to be created.
//
// +kubebuilder:validation:Enum:=azureMonitor;cloudwatch;elasticsearch;http;kafka;loki;lokiStack;googleCloudLogging;splunk;syslog
type OutputType string

// Output type constants, must match JSON tags of OutputTypeSpec fields.
const (
	OutputTypeAzureMonitor       OutputType = "azureMonitor"
	OutputTypeCloudwatch         OutputType = "cloudwatch"
	OutputTypeElasticsearch      OutputType = "elasticsearch"
	OutputTypeGoogleCloudLogging OutputType = "googleCloudLogging"
	OutputTypeHTTP               OutputType = "http"
	OutputTypeKafka              OutputType = "kafka"
	OutputTypeLoki               OutputType = "loki"
	OutputTypeLokiStack          OutputType = "lokiStack"
	OutputTypeSplunk             OutputType = "splunk"
	OutputTypeSyslog             OutputType = "syslog"
)

var (
	// OutputTypes contains all supported output types.
	OutputTypes = []OutputType{
		OutputTypeAzureMonitor,
		OutputTypeCloudwatch,
		OutputTypeElasticsearch,
		OutputTypeGoogleCloudLogging,
		OutputTypeHTTP,
		OutputTypeKafka,
		OutputTypeLoki,
		OutputTypeLokiStack,
		OutputTypeSplunk,
		OutputTypeSyslog,
	}
)

// OutputSpec defines a destination for log messages.
type OutputSpec struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Type of output sink.
	//
	// +kubebuilder:validation:Required
	Type OutputType `json:"type"`

	// TLS contains settings for controlling options on TLS client connections.
	//
	// +kubebuilder:validation:Optional
	TLS *OutputTLSSpec `json:"tls,omitempty"`

	// Limit imposes a limit in records-per-second on the total aggregate rate of logs forwarded
	// to this output from any given collector container. The total log flow from an individual collector
	// container to this output cannot exceed the limit.  Generally, one collector is deployed per cluster node
	// Logs may be dropped to enforce the limit. Missing or 0 means no rate limit.
	//
	// +kubebuilder:validation:Optional
	Limit *LimitSpec `json:"rateLimit,omitempty"`

	// +kubebuilder:validation:Optional
	AzureMonitor *AzureMonitor `json:"azureMonitor,omitempty"`

	// +kubebuilder:validation:Optional
	Cloudwatch *Cloudwatch `json:"cloudwatch,omitempty"`

	// +kubebuilder:validation:Optional
	Elasticsearch *Elasticsearch `json:"elasticsearch,omitempty"`

	// +kubebuilder:validation:Optional
	GoogleCloudLogging *GoogleCloudLogging `json:"googleCloudLogging,omitempty"`

	// +kubebuilder:validation:Optional
	HTTP *HTTP `json:"http,omitempty"`

	// +kubebuilder:validation:Optional
	Kafka *Kafka `json:"kafka,omitempty"`

	// +kubebuilder:validation:Optional
	Loki *Loki `json:"loki,omitempty"`

	// +kubebuilder:validation:Optional
	LokiStack *LokiStack `json:"lokiStack,omitempty"`

	// +kubebuilder:validation:Optional
	Splunk *Splunk `json:"splunk,omitempty"`

	// +kubebuilder:validation:Optional
	Syslog *Syslog `json:"syslog,omitempty"`
}

// OutputTLSSpec contains options for TLS connections that are agnostic to the output type.
type OutputTLSSpec struct {
	TLSSpec `json:",inline"`
	// If InsecureSkipVerify is true, then the TLS client will be configured to skip validating server certificates.
	//
	// This option is *not* recommended for production configurations.
	//
	// +kubebuilder:validation:Optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// TLSSecurityProfile is the security profile to apply to the output connection.
	//
	// +kubebuilder:validation:Optional
	TLSSecurityProfile *openshiftv1.TLSSecurityProfile `json:"securityProfile,omitempty"`
}

type URLSpec struct {
	// URL to send log records to.
	//
	// An absolute URL, with a scheme. Valid schemes depend on `type`.
	// Special schemes `tcp`, `tls`, `udp` and `udps` are used for types that
	// have no scheme of their own. For example, to send syslog records using secure UDP:
	//
	//     { type: syslog, url: udps://syslog.example.com:1234 }
	//
	// Basic TLS is enabled if the URL scheme requires it (for example 'https' or 'tls').
	// The 'username@password' part of `url` is ignored.
	//
	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// BaseOutputTuningSpec tuning parameters for an output
type BaseOutputTuningSpec struct {
	Delivery DeliveryMode `json:"delivery,omitempty"`

	// MaxWrite limits the maximum payload in terms of bytes of a single "send" to the output.
	//
	// +kubebuilder:validation:Optional
	MaxWrite *resource.Quantity `json:"maxWrite,omitempty"`

	// MinRetryDuration is the minimum time to wait between attempts to retry after delivery a failure.
	//
	// +kubebuilder:validation:Optional
	MinRetryDuration *time.Duration `json:"minRetryDuration,omitempty"`

	// MaxRetryDuration is the maximum time to wait between retry attempts after a delivery failure.
	//
	// +kubebuilder:validation:Optional
	MaxRetryDuration *time.Duration `json:"maxRetryDuration,omitempty"`
}

// DeliveryMode sets the delivery mode for log forwarding.
//
// +kubebuilder:validation:Enum:=atLeastOnce;atMostOnce
// +kubebuilder:default:=atLeastOnce
type DeliveryMode string

const (
	// DeliveryModeAtLeastOnce: if the forwarder crashes or is re-started, any logs that were read before
	// the crash but not sent to their destination will be re-read and re-sent. Note it is possible
	// that some logs are duplicated in the event of a crash - log records are delivered at-least-once.
	DeliveryModeAtLeastOnce DeliveryMode = "atLeastOnce"

	// DeliveryModeAtMostOnce: The forwarder makes no effort to recover logs lost during a crash. This mode may give
	// better throughput, but could result in more log loss.
	DeliveryModeAtMostOnce DeliveryMode = "atMostOnce"
)

// HTTPAuthentication provides options for setting common authentication credentials.
// This is mostly used with outputs using HTTP or a derivative as transport.
type HTTPAuthentication struct {
	// Token specifies a bearer token to be used for authenticating requests.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Token *BearerToken `json:"token,omitempty"`

	// Username to use for authenticating requests.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Username *SecretKey `json:"username,omitempty"`

	// Password to use for authenticating requests.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Password *SecretKey `json:"password,omitempty"`
}

// AzureMonitorAuthentication contains configuration for authenticating requests to a AzureMonitor output.
type AzureMonitorAuthentication struct {
	// SharedKey points to the secret containing the shared key used for authenticating requests.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	SharedKey *SecretKey `json:"sharedKey,omitempty"`
}

type AzureMonitor struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *AzureMonitorAuthentication `json:"authentication,omitempty"`

	// CustomerId che unique identifier for the Log Analytics workspace.
	// https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-uri-parameters
	CustomerId string `json:"customerId,omitempty"`

	// LogType the record type of the data that is being submitted.
	// Can only contain letters, numbers, and underscores (_), and may not exceed 100 characters.
	// https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-headers
	LogType string `json:"logType,omitempty"`

	// AzureResourceId the Resource ID of the Azure resource the data should be associated with.
	// https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-headers
	//
	// +kubebuilder:validation:Optional
	AzureResourceId string `json:"azureResourceId,omitempty"`

	// Host alternative host for dedicated Azure regions. (for example for China region)
	// https://docs.azure.cn/en-us/articles/guidance/developerdifferences#check-endpoints-in-azure
	//
	// +kubebuilder:validation:Optional
	Host string `json:"host,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	Tuning *BaseOutputTuningSpec `json:"tuning,omitempty"`
}

type CloudwatchTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	// It is an error if the compression type is not supported by the output.
	//
	// +kubebuilder:validation:Enum:=gzip;none;snappy;zlib;zstd;lz4
	// +kubebuilder:default:=none
	Compression string `json:"compression,omitempty"`
}

// CloudwatchAuthentication contains configuration for authenticating requests to a Cloudwatch output.
type CloudwatchAuthentication struct {
	// AccessKeyID points to the AWS access key id to be used for authentication.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	AccessKeyID *SecretKey `json:"accessKeyID,omitempty"`

	// AccessKeySecret points to the AWS access key secret to be used for authentication.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	AccessKeySecret *SecretKey `json:"accessKeySecret,omitempty"`

	// Credentials points to the secret containing the "credentials file" to be used for authentication.
	// Mostly used for authentication in STS-enabled clusters.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Credentials *SecretKey `json:"credentials,omitempty"`

	// RoleARN points to the secret containing the role ARN to be used for authentication.
	// This is used for authentication in STS-enabled clusters.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	RoleARN *SecretKey `json:"roleARN,omitempty"`
}

// Cloudwatch provides configuration for the output type `cloudwatch`
type Cloudwatch struct {
	// URL to send log records to.
	//
	// The 'username@password' part of `url` is ignored.
	//
	// +kubebuilder:validation:Optional
	URL string `json:"url"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *CloudwatchAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Tuning *CloudwatchTuningSpec `json:"tuning,omitempty"`

	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// GroupBy defines the strategy for grouping logstreams
	//
	// +kubebuilder:validation:Required
	GroupBy LogGroupByType `json:"groupBy"`

	// GroupPrefix Add this prefix to all group names.
	//
	// Useful to avoid group name clashes if an AWS account is used for multiple clusters and
	// used verbatim (e.g. "" means no prefix). The default prefix is cluster-name/log-type
	//
	// +kubebuilder:validation:Optional
	GroupPrefix string `json:"groupPrefix,omitempty"`
}

// LogGroupByType defines a fixed strategy type
//
// +kubebuilder:validation:Enum:=logType;namespaceName;namespaceUUID
type LogGroupByType string

const (
	// LogGroupByLogType is the strategy to group logs by source(e.g. app, infra)
	LogGroupByLogType LogGroupByType = "logType"

	// LogGroupByNamespaceName is the strategy to use for grouping logs by namespace. Infrastructure and
	// audit logs are always grouped by "logType"
	LogGroupByNamespaceName LogGroupByType = "namespaceName"

	// LogGroupByNamespaceUUID  is the strategy to use for grouping logs by namespace UUID. Infrastructure and
	// audit logs are always grouped by "logType"
	LogGroupByNamespaceUUID LogGroupByType = "namespaceUUID"
)

type IndexSpec struct {
	// Index is the tenant for the logs. This supports template syntax
	// to allow dynamic per-event values. Defaults to the log type (i.e. application, audit, infrastructure)
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:default:="{{.log_type}}"
	Index string `json:"index,omitempty"`
}

type ElasticsearchTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;gzip;zlib;zstd
	// +kubebuilder:default:=none
	Compression string `json:"compression,omitempty"`
}

type Elasticsearch struct {
	URLSpec `json:",inline"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	Tuning *ElasticsearchTuningSpec `json:"tuning,omitempty"`

	// defaults to: log_type-write
	IndexSpec `json:",inline"`

	// Version specifies the version of Elasticsearch to be used.
	// Must be one of: 6-8, where 8 is the default
	//
	// +kubebuilder:validation:Minimum:=6
	// +kubebuilder:validation:Maximum:=8
	// +kubebuilder:default:=8
	Version int `json:"version,omitempty"`
}

// GoogleCloudLoggingAuthentication contains configuration for authenticating requests to a GoogleCloudLogging output.
type GoogleCloudLoggingAuthentication struct {
	// Credentials points to the secret containing the `google-application-credentials.json`.
	//
	// +kubebuilder:validation:Required
	Credentials *SecretKey `json:"credentials"`
}

type GoogleCloudLoggingTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`
}

// GoogleCloudLogging provides configuration for sending logs to Google Cloud Logging.
// Exactly one of billingAccountID, organizationID, folderID, or projectID must be set.
type GoogleCloudLogging struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *GoogleCloudLoggingAuthentication `json:"authentication,omitempty"`

	// +kubebuilder:validation:Optional
	BillingAccountID string `json:"billingAccountId,omitempty"`

	// +kubebuilder:validation:Optional
	OrganizationID string `json:"organizationId,omitempty"`

	// +kubebuilder:validation:Optional
	FolderID string `json:"folderId,omitempty"`

	// +kubebuilder:validation:Optional
	ProjectID string `json:"projectId,omitempty"`

	// LogID is the log ID to which to publish logs. This identifies log stream.
	LogID string `json:"logId,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	Tuning *GoogleCloudLoggingTuningSpec `json:"tuning,omitempty"`
}

type HttpTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;gzip;snappy;zlib
	// +kubebuilder:default:=none
	Compression string `json:"compression,omitempty"`
}

// HTTP provided configuration for sending json encoded logs to a generic HTTP endpoint.
type HTTP struct {
	URLSpec `json:",inline"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Tuning *HttpTuningSpec `json:"tuning,omitempty"`

	// Headers specify optional headers to be sent with the request
	//
	// +kubebuilder:validation:Optional
	Headers map[string]string `json:"headers,omitempty"`

	// Timeout specifies the Http request timeout in seconds. If not set, 10secs is used.
	//
	// +kubebuilder:validation:Optional
	Timeout int `json:"timeout,omitempty"`

	// Method specifies the Http method to be used for sending logs. If not set, 'POST' is used.
	//
	// +kubebuilder:validation:Enum:=GET;HEAD;POST;PUT;DELETE;OPTIONS;TRACE;PATCH
	// +kubebuilder:validation:Optional
	Method string `json:"method,omitempty"`
}

type KafkaTuningSpec struct {
	Delivery DeliveryMode `json:"delivery,omitempty"`

	// MaxWrite limits the maximum payload in terms of bytes of a single "send" to the output.
	//
	// +kubebuilder:validation:Optional
	MaxWrite *resource.Quantity `json:"maxWrite,omitempty"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;snappy;zstd;lz4
	// +kubebuilder:default:=none
	Compression string `json:"compression,omitempty"`
}

// KafkaAuthentication contains configuration for authenticating requests to a Kafka output.
type KafkaAuthentication struct {
	// SASL contains options configuring SASL authentication.
	//
	// +kubebuilder:validation:Optional
	SASL *SASLAuthentication `json:"sasl,omitempty"`
}

type SASLAuthentication struct {
	// Username points to the secret to be used as SASL username.
	//
	// +kubebuilder:validation:Optional
	Username *SecretKey `json:"username,omitempty"`

	// Username points to the secret to be used as SASL password.
	//
	// +kubebuilder:validation:Optional
	Password *SecretKey `json:"password,omitempty"`

	// Mechanism sets the SASL mechanism to use.
	//
	// +kubebuilder:validation:Optional
	Mechanism string `json:"mechanism,omitempty"`
}

// Kafka provides optional extra properties for `type: kafka`
type Kafka struct {
	URLSpec `json:",inline"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *KafkaAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Tuning *KafkaTuningSpec `json:"tuning,omitempty"`

	// Topic specifies the target topic to send logs to.
	//
	// +kubebuilder:validation:Optional
	Topic string `json:"topic,omitempty"`

	// Brokers specifies the list of broker endpoints of a Kafka cluster.
	// The list represents only the initial set used by the collector's Kafka client for the
	// first connection only. The collector's Kafka client fetches constantly an updated list
	// from Kafka. These updates are not reconciled back to the collector configuration.
	// If none provided the target URL from the OutputSpec is used as fallback.
	//
	// +kubebuilder:validation:Optional
	Brokers []string `json:"brokers,omitempty"`
}

type LokiTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;gzip;snappy
	// +kubebuilder:default:=snappy
	Compression string `json:"compression,omitempty"`
}

// LokiStackTarget contains information about how to reach the LokiStack used as an output.
type LokiStackTarget struct {
	// Namespace of the in-cluster LokiStack resource.
	//
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// Name of the in-cluster LokiStack resource.
	//
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// LokiStack provides optional extra properties for `type: lokistack`
type LokiStack struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Target points to the LokiStack resources that should be used as a target for the output.
	//
	// +kubebuilder:validation:Required
	Target LokiStackTarget `json:"target"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	Tuning *LokiTuningSpec `json:"tuning,omitempty"`

	// LabelKeys is a list of log record keys that will be used as Loki labels with the corresponding log record value.
	//
	// If LabelKeys is not set, the default keys are `[log_type, kubernetes.namespace_name, kubernetes.pod_name, kubernetes_host]`
	//
	// Note: Loki label names must match the regular expression "[a-zA-Z_:][a-zA-Z0-9_:]*"
	// Log record keys may contain characters like "." and "/" that are not allowed in Loki labels.
	// Log record keys are translated to Loki labels by replacing any illegal characters with '_'.
	// For example the default log record keys translate to these Loki labels: `log_type`, `kubernetes_namespace_name`, `kubernetes_pod_name`, `kubernetes_host`
	//
	// Note: the set of labels should be small, Loki imposes limits on the size and number of labels allowed.
	// See https://grafana.com/docs/loki/latest/configuration/#limits_config for more.
	// Loki queries can also query based on any log record field (not just labels) using query filters.
	//
	// +kubebuilder:validation:Optional
	LabelKeys []string `json:"labelKeys,omitempty"`
}

// Loki provides optional extra properties for `type: loki`
type Loki struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Tuning *LokiTuningSpec `json:"tuning,omitempty"`

	URLSpec `json:",inline"`

	// LabelKeys is a list of log record keys that will be used as Loki labels with the corresponding log record value.
	//
	// If LabelKeys is not set, the default keys are `[log_type, kubernetes.namespace_name, kubernetes.pod_name, kubernetes_host]`
	//
	// Note: Loki label names must match the regular expression "[a-zA-Z_:][a-zA-Z0-9_:]*"
	// Log record keys may contain characters like "." and "/" that are not allowed in Loki labels.
	// Log record keys are translated to Loki labels by replacing any illegal characters with '_'.
	// For example the default log record keys translate to these Loki labels: `log_type`, `kubernetes_namespace_name`, `kubernetes_pod_name`, `kubernetes_host`
	//
	// Note: the set of labels should be small, Loki imposes limits on the size and number of labels allowed.
	// See https://grafana.com/docs/loki/latest/configuration/#limits_config for more.
	// Loki queries can also query based on any log record field (not just labels) using query filters.
	//
	// +kubebuilder:validation:Optional
	LabelKeys []string `json:"labelKeys,omitempty"`

	// TenantKey is the tenant for the logs. This supports vector's template syntax
	// to allow dynamic per-event values. Defaults to the log type (i.e. application, audit, infrastructure)
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:default:="{{.log_type}}"
	TenantKey string `json:"tenantKey,omitempty"`
}

type SplunkTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`
}

// SplunkAuthentication contains configuration for authenticating requests to a Splunk output.
type SplunkAuthentication struct {
	// Token points to the secret containing the Splunk HEC token used for authenticating requests.
	//
	// +kubebuilder:validation:Optional
	Token *SecretKey `json:"token,omitempty"`
}

// Splunk Deliver log data to Splunk’s HTTP Event Collector
// Provides optional extra properties for `type: splunk_hec` ('splunk_hec_logs' after Vector 0.23
type Splunk struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	Authentication *SplunkAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +nullable
	Tuning *SplunkTuningSpec `json:"tuning,omitempty"`

	URLSpec `json:",inline"`

	// defaults to: Splunk receiver's defined index
	IndexSpec `json:",inline"`
}

type SyslogRFCType string

const (
	SyslogRFC3164 SyslogRFCType = "RFC3164"
	SyslogRFC5424 SyslogRFCType = "RFC5424"
)

// Syslog provides optional extra properties for output type `syslog`
type Syslog struct {
	URLSpec `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=RFC3164;RFC5424
	// +kubebuilder:default:=RFC5424
	RFC SyslogRFCType `json:"rfc,omitempty"`

	// Severity to set on outgoing syslog records.
	//
	// Severity values are defined in https://tools.ietf.org/html/rfc5424#section-6.2.1
	// The value can be a decimal integer or one of these case-insensitive keywords:
	//
	//     Emergency Alert Critical Error Warning Notice Informational Debug
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=informational
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
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=user
	Facility string `json:"facility,omitempty"`

	// PayloadKey specifies record field to use as payload.
	//
	// +kubebuilder:validation:Optional
	PayloadKey string `json:"payloadKey,omitempty"`

	// AppName is APP-NAME part of the syslog-msg header
	//
	// AppName needs to be specified if using rfc5424. The maximum length of the final values is truncated to 48
	//
	// +kubebuilder:validation:Optional
	// TODO: DETERMIN HOW to default the app name that isnt based on fluentd assumptions of "tag" when this is empty
	AppName string `json:"appName,omitempty"`

	// ProcID is PROCID part of the syslog-msg header
	//
	// ProcID needs to be specified if using rfc5424. The maximum length of the final values is truncated to 128
	//
	// +kubebuilder:validation:Optional
	ProcID string `json:"procID,omitempty"`

	// MsgID is MSGID part of the syslog-msg header
	//
	// MsgID needs to be specified if using rfc5424.  The maximum length of the final values is truncated to 32
	//
	// +kubebuilder:validation:Optional
	MsgID string `json:"msgID,omitempty"`
}
