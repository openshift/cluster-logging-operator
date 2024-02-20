package v1

// NOTE: The Enum validation on OutputSpec.Type must be updated if the list of
// known types changes.

// Output type constants, must match JSON tags of OutputTypeSpec fields.
const (
	OutputTypeCloudwatch         = "cloudwatch"
	OutputTypeElasticsearch      = "elasticsearch"
	OutputTypeFluentdForward     = "fluentdForward"
	OutputTypeSyslog             = "syslog"
	OutputTypeKafka              = "kafka"
	OutputTypeLoki               = "loki"
	OutputTypeGoogleCloudLogging = "googleCloudLogging"
	OutputTypeSplunk             = "splunk"
	OutputTypeHttp               = "http"
	OutputTypeAzureMonitor       = "azureMonitor"
)

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
	// +optional
	Cloudwatch *Cloudwatch `json:"cloudwatch,omitempty"`
	// +optional
	Loki *Loki `json:"loki,omitempty"`
	// +optional
	GoogleCloudLogging *GoogleCloudLogging `json:"googleCloudLogging,omitempty"`
	// +optional
	Splunk *Splunk `json:"splunk,omitempty"`
	// +optional
	Http *Http `json:"http,omitempty"`
	// +optional
	AzureMonitor *AzureMonitor `json:"azureMonitor,omitempty"`
}

// Cloudwatch provides configuration for the output type `cloudwatch`
//
// Note: the cloudwatch output recognizes the following keys in the Secret:
//
//	`aws_secret_access_key`: AWS secret access key.
//	`aws_access_key_id`: AWS secret access key ID.
//
// Or for sts-enabled clusters `credentials` or `role_arn` key specifying a properly formatted role arn
type Cloudwatch struct {
	// +required
	Region string `json:"region,omitempty"`

	//GroupBy defines the strategy for grouping logstreams
	// +required
	//+kubebuilder:validation:Enum:=logType;namespaceName;namespaceUUID
	GroupBy LogGroupByType `json:"groupBy,omitempty"`

	//GroupPrefix Add this prefix to all group names.
	//  Useful to avoid group name clashes if an AWS account is used for multiple clusters and
	//  used verbatim (e.g. "" means no prefix)
	//  The default prefix is cluster-name/log-type
	// +optional
	GroupPrefix *string `json:"groupPrefix,omitempty"`
}

// LogGroupByType defines a fixed strategy type
type LogGroupByType string

const (
	//LogGroupByLogType is the strategy to group logs by source(e.g. app, infra)
	LogGroupByLogType LogGroupByType = "logType"

	// LogGroupByNamespaceName is the strategy to use for grouping logs by namespace. Infrastructure and
	// audit logs are always grouped by "logType"
	LogGroupByNamespaceName LogGroupByType = "namespaceName"

	// LogGroupByNamespaceUUID  is the strategy to use for grouping logs by namespace UUID. Infrastructure and
	// audit logs are always grouped by "logType"
	LogGroupByNamespaceUUID LogGroupByType = "namespaceUUID"
)

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

	// AddLogSource adds log's source information to the log message
	// If the logs are collected from a process; namespace_name, pod_name, container_name is added to the log
	// In addition, it picks the originating process name and id(known as the `pid`) from the record
	// and injects them into the header field."
	//
	// +optional
	AddLogSource bool `json:"addLogSource,omitempty"`

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

	// Brokers specifies the list of broker endpoints of a Kafka cluster.
	// The list represents only the initial set used by the collector's Kafka client for the
	// first connection only. The collector's Kafka client fetches constantly an updated list
	// from Kafka. These updates are not reconciled back to the collector configuration.
	// If none provided the target URL from the OutputSpec is used as fallback.
	//
	// +optional
	Brokers []string `json:"brokers,omitempty"`
}

// FluentdForward does not provide additional fields, but note that
// the fluentforward output allows this additional keys in the Secret:
//
//	`shared_key`: (string) Key to enable fluent-forward shared-key authentication.
type FluentdForward struct{}

type Elasticsearch struct {
	ElasticsearchStructuredSpec `json:",inline"`

	// Version specifies the version of Elasticsearch to be used.
	// Must be one of:
	//  - 6 - Default for internal ES store
	//  - 7
	//  - 8 - Latest for external ES store
	//
	// +kubebuilder:validation:Minimum:=6
	// +optional
	Version int `json:"version,omitempty"`
}

// ElasticsearchStructuredSpec is spec related to structured log changes to determine the elasticsearch index
type ElasticsearchStructuredSpec struct {
	// StructuredTypeKey specifies the metadata key to be used as name of elasticsearch index
	// It takes precedence over StructuredTypeName
	//
	// +optional
	StructuredTypeKey string `json:"structuredTypeKey,omitempty"`

	// StructuredTypeName specifies the name of elasticsearch schema
	//
	// +optional
	StructuredTypeName string `json:"structuredTypeName,omitempty"`

	// EnableStructuredContainerLogs enables multi-container structured logs to allow
	// forwarding logs from containers within a pod to separate indices.  Annotating
	// the pod with key 'containerType.logging.openshift.io/<container-name>' and value
	// '<structure-type-name>' will forward those container logs to an alternate index
	// from that defined by the other 'structured' keys here
	//
	// +optional
	EnableStructuredContainerLogs bool `json:"enableStructuredContainerLogs,omitempty"`
}

// Loki provides optional extra properties for `type: loki`
type Loki struct {
	// TenantKey is a meta-data key field to use as the TenantID,
	// For example: 'TenantKey: kubernetes.namespace_name` will use the kubernetes
	// namespace as the tenant ID.
	//
	// +optional
	TenantKey string `json:"tenantKey,omitempty"`

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
	// +optional
	LabelKeys []string `json:"labelKeys,omitempty"`
}

// GoogleCloudLogging provides configuration for sending logs to Google Cloud Logging.
// Exactly one of billingAccountID, organizationID, folderID, or projectID must be set.
type GoogleCloudLogging struct {
	// +optional
	BillingAccountID string `json:"billingAccountId,omitempty"`

	// +optional
	OrganizationID string `json:"organizationId,omitempty"`

	// +optional
	FolderID string `json:"folderId,omitempty"`

	// +optional
	ProjectID string `json:"projectId,omitempty"`

	//LogID is the log ID to which to publish logs. This identifies log stream.
	LogID string `json:"logId,omitempty"`
}

// Splunk Deliver log data to Splunk’s HTTP Event Collector
// Provides optional extra properties for `type: splunk_hec` ('splunk_hec_logs' after Vector 0.23
type Splunk struct {
	// IndexKey is a meta-data key field to use to send events to.
	// For example: 'IndexKey: kubernetes.namespace_name` will use the kubernetes
	// namespace as the index.
	// If the IndexKey is not found, the default index defined within Splunk is used.
	// Only one of IndexKey or IndexName can be defined.
	// If IndexKey && IndexName are not specified, the default index defined within Splunk is used.
	// +optional
	IndexKey string `json:"indexKey,omitempty"`

	// IndexName is the name of the index to send events to.
	// Only one of IndexKey or IndexName can be defined.
	// If IndexKey && IndexName are not specified, the default index defined within Splunk is used.
	// +optional
	IndexName string `json:"indexName,omitempty"`

	// Deprecated. Fields to be added to Splunk index.
	// +optional
	// +deprecated
	Fields []string `json:"fields,omitempty"`
}

// Http provided configuration for sending json encoded logs to a generic http endpoint.
type Http struct {
	// Headers specify optional headers to be sent with the request
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// Timeout specifies the Http request timeout in seconds. If not set, 10secs is used.
	// +optional
	Timeout int `json:"timeout,omitempty"`

	// Method specifies the Http method to be used for sending logs. If not set, 'POST' is used.
	// +kubebuilder:validation:Enum:=GET;HEAD;POST;PUT;DELETE;OPTIONS;TRACE;PATCH
	// +optional
	Method string `json:"method,omitempty"`

	// Schema enables configuration of the way log records are normalized.
	//
	// Supported models: viaq(default), opentelemetry
	//
	// Logs are converted to the Open Telemetry specification according to schema value
	//
	// +kubebuilder:validation:Enum:=opentelemetry;viaq
	// +kubebuilder:default:viaq
	// +optional
	Schema string `json:"schema,omitempty"`
}

type AzureMonitor struct {
	//CustomerId che unique identifier for the Log Analytics workspace.
	//https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-uri-parameters
	CustomerId string `json:"customerId,omitempty"`

	//LogType the record type of the data that is being submitted.
	//Can only contain letters, numbers, and underscores (_), and may not exceed 100 characters.
	//https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-headers
	LogType string `json:"logType,omitempty"`

	//AzureResourceId the Resource ID of the Azure resource the data should be associated with.
	//https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-headers
	// +optional
	AzureResourceId string `json:"azureResourceId,omitempty"`

	//Host alternative host for dedicated Azure regions. (for example for China region)
	//https://docs.azure.cn/en-us/articles/guidance/developerdifferences#check-endpoints-in-azure
	// +optional
	Host string `json:"host,omitempty"`
}
