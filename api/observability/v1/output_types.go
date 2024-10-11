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
// +kubebuilder:validation:Enum:=azureMonitor;cloudwatch;elasticsearch;http;kafka;loki;lokiStack;googleCloudLogging;splunk;syslog;otlp
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
	OutputTypeOTLP               OutputType = "otlp"
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
		OutputTypeOTLP,
	}
)

// OutputSpec defines a destination for log messages.
//
// +kubebuilder:validation:XValidation:rule="self.type != 'azureMonitor' || has(self.azureMonitor)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'cloudwatch' || has(self.cloudwatch)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'elasticsearch' || has(self.elasticsearch)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'googleCloudLogging' || has(self.googleCloudLogging)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'http' || has(self.http)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'kafka' || has(self.kafka)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'loki' || has(self.loki)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'lokiStack' || has(self.lokiStack)", message="Additional type specific spec is required for the output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'splunk' || has(self.splunk)", message="Additional type specific spec is required the for output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'syslog' || has(self.syslog)", message="Additional type specific spec is required the for output type"
// +kubebuilder:validation:XValidation:rule="self.type != 'otlp' || has(self.otlp)", message="Additional type specific spec is required the for output type"
type OutputSpec struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +kubebuilder:validation:Pattern:="^[a-z][a-z0-9-]*[a-z0-9]$"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Output Name"
	Name string `json:"name"`

	// Type of output sink.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Output Type"
	Type OutputType `json:"type"`

	// TLS contains settings for controlling options on TLS client connections.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLS Options"
	TLS *OutputTLSSpec `json:"tls,omitempty"`

	// Limit imposes a limit in records-per-second on the total aggregate rate of logs forwarded
	// to this output from any given collector container. The total log flow from an individual collector
	// container to this output cannot exceed the limit.  Generally, one collector is deployed per cluster node
	// Logs may be dropped to enforce the limit. Missing or 0 means no rate limit.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Rate Limiting"
	Limit *LimitSpec `json:"rateLimit,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Azure Monitor"
	AzureMonitor *AzureMonitor `json:"azureMonitor,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Amazon CloudWatch"
	Cloudwatch *Cloudwatch `json:"cloudwatch,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ElasticSearch"
	Elasticsearch *Elasticsearch `json:"elasticsearch,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Google Cloud Logging"
	GoogleCloudLogging *GoogleCloudLogging `json:"googleCloudLogging,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="HTTP Output"
	HTTP *HTTP `json:"http,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Apache Kafka"
	Kafka *Kafka `json:"kafka,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Grafana Loki"
	Loki *Loki `json:"loki,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="LokiStack"
	LokiStack *LokiStack `json:"lokiStack,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Splunk"
	Splunk *Splunk `json:"splunk,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Syslog Output"
	Syslog *Syslog `json:"syslog,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="OpenTelemetry Output"
	OTLP *OTLP `json:"otlp,omitempty"`
}

// OutputTLSSpec contains options for TLS connections that are agnostic to the output type.
type OutputTLSSpec struct {
	TLSSpec `json:",inline"`
	// If InsecureSkipVerify is true, then the TLS client will be configured to skip validating server certificates.
	//
	// This option is *not* recommended for production configurations.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Skip Certificate Validation",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// TLSSecurityProfile is the security profile to apply to the output connection.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLS Security Profile"
	TLSSecurityProfile *openshiftv1.TLSSecurityProfile `json:"securityProfile,omitempty"`
}

type URLSpec struct {
	// URL to send log records to.
	// Basic TLS is enabled if the URL scheme requires it (for example 'https' or 'tls').
	// The 'username@password' part of `url` is ignored.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="isURL(self)", message="invalid URL"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Destination URL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	URL string `json:"url"`
}

// BaseOutputTuningSpec tuning parameters for an output
type BaseOutputTuningSpec struct {
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Delivery Mode"
	DeliveryMode DeliveryMode `json:"deliveryMode,omitempty"`

	// MaxWrite limits the maximum payload in terms of bytes of a single "send" to the output.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Batch Size"
	MaxWrite *resource.Quantity `json:"maxWrite,omitempty"`

	// MinRetryDuration is the minimum time to wait between attempts to retry after delivery a failure.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Minimum Retry Duration"
	MinRetryDuration *time.Duration `json:"minRetryDuration,omitempty"`

	// MaxRetryDuration is the maximum time to wait between retry attempts after a delivery failure.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Maximum Retry Duration"
	MaxRetryDuration *time.Duration `json:"maxRetryDuration,omitempty"`
}

// DeliveryMode sets the delivery mode for log forwarding.
//
// +kubebuilder:validation:Enum:=AtLeastOnce;AtMostOnce
type DeliveryMode string

const (
	// DeliveryModeAtLeastOnce: if the forwarder crashes or is re-started, any logs that were read before
	// the crash but not sent to their destination will be re-read and re-sent. Note it is possible
	// that some logs are duplicated in the event of a crash - log records are delivered at-least-once.
	DeliveryModeAtLeastOnce DeliveryMode = "AtLeastOnce"

	// DeliveryModeAtMostOnce: The forwarder makes no effort to recover logs lost during a crash. This mode may give
	// better throughput, but could result in more log loss.
	DeliveryModeAtMostOnce DeliveryMode = "AtMostOnce"
)

// HTTPAuthentication provides options for setting common authentication credentials.
// This is mostly used with outputs using HTTP or a derivative as transport.
type HTTPAuthentication struct {
	// Token specifies a bearer token to be used for authenticating requests.
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Bearer Token"
	Token *BearerToken `json:"token,omitempty"`

	// Username to use for authenticating requests.
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Username"
	Username *SecretReference `json:"username,omitempty"`

	// Password to use for authenticating requests.
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Password"
	Password *SecretReference `json:"password,omitempty"`
}

// AzureMonitorAuthentication contains configuration for authenticating requests to a AzureMonitor output.
type AzureMonitorAuthentication struct {
	// SharedKey points to the secret containing the shared key used for authenticating requests.
	//
	// +nullable
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Shared Key"
	SharedKey *SecretReference `json:"sharedKey"`
}

type AzureMonitor struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *AzureMonitorAuthentication `json:"authentication"`

	// CustomerId che unique identifier for the Log Analytics workspace.
	// https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-uri-parameters
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Customer ID",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	CustomerId string `json:"customerId"`

	// LogType the record type of the data that is being submitted.
	// Can only contain letters, numbers, and underscores (_), and may not exceed 100 characters.
	// https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-headers
	//
	// +kubebuilder:validation:Pattern:="^[a-zA-Z0-9][a-zA-Z0-9_]{0,99}$"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Type",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	LogType string `json:"logType,omitempty"`

	// AzureResourceId the Resource ID of the Azure resource the data should be associated with.
	// https://learn.microsoft.com/en-us/azure/azure-monitor/logs/data-collector-api?tabs=powershell#request-headers
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Azure Resource ID",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	AzureResourceId string `json:"azureResourceId,omitempty"`

	// Host alternative host for dedicated Azure regions. (for example for China region)
	// https://docs.azure.cn/en-us/articles/guidance/developerdifferences#check-endpoints-in-azure
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Azure Host",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Host string `json:"host,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *BaseOutputTuningSpec `json:"tuning,omitempty"`
}

type CloudwatchTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	// It is an error if the compression type is not supported by the output.
	//
	// +kubebuilder:validation:Enum:=gzip;none;snappy;zlib;zstd
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression"
	Compression string `json:"compression,omitempty"`
}

// Cloudwatch provides configuration for the output type `cloudwatch`
type Cloudwatch struct {
	// URL to send log records to.
	//
	// The 'username@password' part of `url` is ignored.
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:XValidation:rule="self == '' ||  isURL(self)", message="invalid URL"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Destination URL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	URL string `json:"url,omitempty"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *CloudwatchAuthentication `json:"authentication"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +nullable
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *CloudwatchTuningSpec `json:"tuning,omitempty"`

	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Amazon Region",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Region string `json:"region"`

	// GroupName defines the strategy for grouping logstreams
	//
	// The GroupName can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Group Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	GroupName string `json:"groupName"`
}

// CloudwatchAuthType sets the authentication type used for CloudWatch.
//
// +kubebuilder:validation:Enum:=awsAccessKey;iamRole
type CloudwatchAuthType string

const (
	// CloudwatchAuthTypeAccessKey requires auth to use static keys
	CloudwatchAuthTypeAccessKey CloudwatchAuthType = "awsAccessKey"

	// CloudwatchAuthTypeIAMRole requires auth to use IAM Role and optional token
	CloudwatchAuthTypeIAMRole CloudwatchAuthType = "iamRole"
)

// CloudwatchAuthentication contains configuration for authenticating requests to a Cloudwatch output.
// +kubebuilder:validation:XValidation:rule="self.type != 'awsAccessKey' || has(self.awsAccessKey)", message="Additional type specific spec is required for authentication"
// +kubebuilder:validation:XValidation:rule="self.type != 'iamRole' || has(self.iamRole)", message="Additional type specific spec is required for authentication"
type CloudwatchAuthentication struct {
	// Type is the type of cloudwatch authentication to configure
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Type"
	Type CloudwatchAuthType `json:"type"`

	// AWSAccessKey points to the AWS access key id and secret to be used for authentication.
	//
	// +kubebuilder:validation:Optional
	// +nullable
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Access Key"
	AWSAccessKey *CloudwatchAWSAccessKey `json:"awsAccessKey,omitempty"`

	// IAMRole points to the secret containing the role ARN to be used for authentication.
	// This can be used for authentication in STS-enabled clusters when additionally specifying
	// a web identity token
	//
	// +kubebuilder:validation:Optional
	// +nullable
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Amazon IAM Role"
	IAMRole *CloudwatchIAMRole `json:"iamRole,omitempty"`
}

type CloudwatchIAMRole struct {
	// RoleARN points to the secret containing the role ARN to be used for authentication.
	// This is used for authentication in STS-enabled clusters.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="RoleARN Secret"
	RoleARN SecretReference `json:"roleARN"`

	// Token specifies a bearer token to be used for authenticating requests.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Token"
	Token BearerToken `json:"token"`
}

type CloudwatchAWSAccessKey struct {
	// KeyId points to the AWS access key id to be used for authentication.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret with Access Key ID"
	KeyId SecretReference `json:"keyId"`

	// KeySecret points to the AWS access key secret to be used for authentication.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret with Access Key Secret"
	KeySecret SecretReference `json:"keySecret"`
}

type ElasticsearchTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;gzip;zlib
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression"
	Compression string `json:"compression,omitempty"`
}

type Elasticsearch struct {
	URLSpec `json:",inline"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *ElasticsearchTuningSpec `json:"tuning,omitempty"`

	// Index is the index for the logs. This supports template syntax to allow dynamic per-event values.
	//
	// The Index can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Index",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Index string `json:"index"`

	// Version specifies the version of Elasticsearch to be used.
	// Must be one of: 6-8
	//
	// +kubebuilder:validation:Minimum:=6
	// +kubebuilder:validation:Maximum:=8
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ElasticSearch Version",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	Version int `json:"version"`
}

// GoogleCloudLoggingAuthentication contains configuration for authenticating requests to a GoogleCloudLogging output.
type GoogleCloudLoggingAuthentication struct {
	// Credentials points to the secret containing the `google-application-credentials.json`.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret with Credentials File"
	Credentials *SecretReference `json:"credentials"`
}

type GoogleCloudLoggingTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`
}

// GoogleCloudLogging provides configuration for sending logs to Google Cloud Logging.
type GoogleCloudLogging struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *GoogleCloudLoggingAuthentication `json:"authentication,omitempty"`

	// ID must be one of the required ID fields for the output
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Logging ID"
	ID GoogleCloudLoggingId `json:"id"`

	// LogId is the log ID to which to publish logs. This identifies log stream.
	//
	// The LogId can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Stream ID",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	LogId string `json:"logId"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *GoogleCloudLoggingTuningSpec `json:"tuning,omitempty"`
}

type GoogleCloudLoggingId struct {
	// Type is the ID type provided
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Logging ID Type"
	Type GoogleCloudLoggingIdType `json:"type"`

	// Value is the value of the ID
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Logging ID Value",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Value string `json:"value"`
}

// GoogleCloudLoggingIdType specifies the type of the provided ID value.
//
// +kubebuilder:validation:Enum:=billingAccount;folder;project;organization
type GoogleCloudLoggingIdType string

const (
	GoogleCloudLoggingIdTypeBillingAccount GoogleCloudLoggingIdType = "billingAccount"
	GoogleCloudLoggingIdTypeFolder         GoogleCloudLoggingIdType = "folder"
	GoogleCloudLoggingIdTypeProject        GoogleCloudLoggingIdType = "project"
	GoogleCloudLoggingIdTypeOrganization   GoogleCloudLoggingIdType = "organization"
)

type HTTPTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;gzip;snappy;zlib
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression"
	Compression string `json:"compression,omitempty"`
}

// HTTP provided configuration for sending json encoded logs to a generic HTTP endpoint.
type HTTP struct {
	URLSpec `json:",inline"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *HTTPTuningSpec `json:"tuning,omitempty"`

	// Headers specify optional headers to be sent with the request
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Headers"
	Headers map[string]string `json:"headers,omitempty"`

	// Timeout specifies the Http request timeout in seconds. If not set, 10secs is used.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Timeout",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	Timeout int `json:"timeout,omitempty"`

	// Method specifies the Http method to be used for sending logs. If not set, 'POST' is used.
	//
	// +kubebuilder:validation:Enum:=GET;HEAD;POST;PUT;DELETE;OPTIONS;TRACE;PATCH
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="HTTP Method"
	Method string `json:"method,omitempty"`
}

type KafkaTuningSpec struct {
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Delivery Mode"
	DeliveryMode DeliveryMode `json:"deliveryMode,omitempty"`

	// MaxWrite limits the maximum payload in terms of bytes of a single "send" to the output.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Batch Size"
	MaxWrite *resource.Quantity `json:"maxWrite,omitempty"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;snappy;zstd;lz4
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression"
	Compression string `json:"compression,omitempty"`
}

// KafkaAuthentication contains configuration for authenticating requests to a Kafka output.
type KafkaAuthentication struct {
	// SASL contains options configuring SASL authentication.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="SASL Options"
	SASL *SASLAuthentication `json:"sasl,omitempty"`
}

type SASLAuthentication struct {
	// Username points to the secret to be used as SASL username.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret with Username"
	Username *SecretReference `json:"username,omitempty"`

	// Username points to the secret to be used as SASL password.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret with Password"
	Password *SecretReference `json:"password,omitempty"`

	// Mechanism sets the SASL mechanism to use.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="SASL Mechanism",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Mechanism string `json:"mechanism,omitempty"`
}

// Kafka provides optional extra properties for `type: kafka`
// +kubebuilder:validation:XValidation:rule="has(self.url) || self.brokers.size() > 0", message="URL or brokers required"
type Kafka struct {

	// URL to send log records to.
	//
	// The 'username@password' part of `url` is ignored.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:XValidation:rule="self == '' ||  isURL(self)", message="invalid URL"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Destination URL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	URL string `json:"url,omitempty"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *KafkaAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *KafkaTuningSpec `json:"tuning,omitempty"`

	// Topic specifies the target topic to send logs to. The value when not specified is 'topic'
	//
	// The Topic can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kafka Topic",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Topic string `json:"topic,omitempty"`

	// Brokers specifies the list of broker endpoints of a Kafka cluster.
	//
	// The list represents only the initial set used by the collector's Kafka client for the
	// first connection only. The collector's Kafka client fetches constantly an updated list
	// from Kafka. These updates are not reconciled back to the collector configuration.
	//
	// If none provided the target URL from the OutputSpec is used as fallback.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kafka Brokers"
	Brokers []URL `json:"brokers,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="isURL(self)", message="invalid URL"
type URL string

type LokiTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;gzip;snappy
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression"
	Compression string `json:"compression,omitempty"`
}

// LokiStackTarget contains information about how to reach the LokiStack used as an output.
type LokiStackTarget struct {
	// Namespace of the in-cluster LokiStack resource.
	//
	// If unset, this defaults to "openshift-logging".
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="LokiStack Namespace",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Namespace string `json:"namespace,omitempty"`

	// Name of the in-cluster LokiStack resource.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern:="^[a-z][a-z0-9-]{2,62}[a-z0-9]$"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="LokiStack Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Name string `json:"name"`
}

// LokiStackAuthentication is the authentication for LokiStack
type LokiStackAuthentication struct {
	// Token specifies a bearer token to be used for authenticating requests.
	//
	// +nullable
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Bearer Token"
	Token *BearerToken `json:"token"`
}

// LokiStack provides optional extra properties for `type: lokistack`
// +kubebuilder:validation:XValidation:rule="!has(self.labelKeys) || !has(self.dataModel) || self.dataModel == 'Viaq'", message="'labelKeys' cannot be set when data model is 'Otel'"
// +kubebuilder:validation:XValidation:rule="!has(self.tuning) || self.tuning.compression != 'snappy' || !has(self.dataModel) || self.dataModel == 'Viaq'", message="'snappy' compression cannot be used when data model is 'Otel'"
type LokiStack struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *LokiStackAuthentication `json:"authentication"`

	// Target points to the LokiStack resources that should be used as a target for the output.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Target LokiStack Reference"
	Target LokiStackTarget `json:"target"`

	// Tuning specs tuning for the output
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *LokiTuningSpec `json:"tuning,omitempty"`

	// LabelKeys can be used to customize which log record keys are mapped to Loki stream labels.
	//
	// Note: Loki label names must match the regular expression "[a-zA-Z_:][a-zA-Z0-9_:]*"
	// Log record keys may contain characters like "." and "/" that are not allowed in Loki labels.
	// Log record keys are translated to Loki labels by replacing any illegal characters with '_'.
	//
	// For example the default log record keys translate to these Loki labels:
	//
	// - log_type
	//
	// - kubernetes_container_name
	//
	// - kubernetes_namespace_name
	//
	// - kubernetes_pod_name
	//
	// Note: the set of labels should be small, Loki imposes limits on the size and number of labels allowed.
	// See https://grafana.com/docs/loki/latest/configuration/#limits_config for more.
	// Loki queries can also query based on any log record field (not just labels) using query filters.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Stream Label Configuration"
	LabelKeys *LokiStackLabelKeys `json:"labelKeys,omitempty"`

	// DataModel can be used to customize how log data is stored in LokiStack.
	//
	// There are two different models to choose from:
	//
	//  - Viaq
	//  - Otel
	//
	// When the data model is not set, it currently defaults to the "Viaq" data model.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Data Model"
	DataModel LokiStackDataModel `json:"dataModel,omitempty"`
}

// LokiStackLabelKeys contains the configuration that maps log record's keys to Loki labels used to identify streams.
type LokiStackLabelKeys struct {
	// Global contains a list of record keys which are used for all tenants.
	//
	// If LabelKeys is not set, the default keys are:
	//
	//  - log_type
	//
	//  - kubernetes.container_name
	//
	//  - kubernetes.namespace_name
	//
	//  - kubernetes.pod_name
	//
	// One additional label "kubernetes_host" is not part of the label keys configuration. It contains the hostname
	// where the collector is running and is always present.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Global Configuration"
	Global []string `json:"global,omitempty"`

	// Application contains the label keys configuration for the "application" tenant.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Application Tenant Configuration"
	Application *LokiStackTenantLabelKeys `json:"application,omitempty"`

	// Infrastructure contains the label keys configuration for the "infrastructure" tenant.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Infrastructure Tenant Configuration"
	Infrastructure *LokiStackTenantLabelKeys `json:"infrastructure,omitempty"`

	// Audit contains the label keys configuration for the "audit" tenant.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Audit Tenant Configuration"
	Audit *LokiStackTenantLabelKeys `json:"audit,omitempty"`
}

// LokiStackTenantLabelKeys contains options for customizing the mapping of log record keys to Loki stream labels for a single tenant.
type LokiStackTenantLabelKeys struct {
	// If IgnoreGlobal is true, then the tenant will not use the labels configured in the Global section of the label
	// keys configuration.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Ignore Global Settings",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	IgnoreGlobal bool `json:"ignoreGlobal,omitempty"`

	// LabelKeys contains a list of log record keys that are mapped to Loki stream labels.
	//
	// By default, this list is combined with the labels specified in the Global configuration.
	// This behavior can be changed by setting IgnoreGlobal to true.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Label Keys"
	LabelKeys []string `json:"labelKeys,omitempty"`
}

// LokiStackDataModel selects which data model is used to send and store the log data.
//
// +kubebuilder:validation:Enum:=Viaq;Otel
type LokiStackDataModel string

const (
	// LokiStackDataModelViaq selects the ViaQ data model for the LokiStack output.
	LokiStackDataModelViaq LokiStackDataModel = "Viaq"
	// LokiStackDataModelOpenTelemetry selects a data model based on the OpenTelemetry semantic conventions
	// and uses OTLP as transport.
	LokiStackDataModelOpenTelemetry LokiStackDataModel = "Otel"
)

// Loki provides optional extra properties for `type: loki`
type Loki struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *LokiTuningSpec `json:"tuning,omitempty"`

	URLSpec `json:",inline"`

	// LabelKeys can be used to customize which log record keys are mapped to Loki stream labels.
	//
	// If LabelKeys is not set, the default keys are:
	//
	// - log_type
	//
	// - kubernetes.container_name
	//
	// - kubernetes.namespace_name
	//
	// - kubernetes.pod_name
	//
	// One additional label "kubernetes_host" is not part of the label keys configuration. It contains the hostname
	// where the collector is running and is always present.
	//
	// Note: Loki label names must match the regular expression "[a-zA-Z_:][a-zA-Z0-9_:]*"
	// Log record keys may contain characters like "." and "/" that are not allowed in Loki labels.
	// Log record keys are translated to Loki labels by replacing any illegal characters with '_'.
	//
	// For example the default log record keys translate to these Loki labels:
	//
	// - log_type
	//
	// - kubernetes_container_name
	//
	// - kubernetes_namespace_name
	//
	// - kubernetes_pod_name
	//
	// Note: the set of labels should be small, Loki imposes limits on the size and number of labels allowed.
	// See https://grafana.com/docs/loki/latest/configuration/#limits_config for more.
	// Loki queries can also query based on any log record field (not just labels) using query filters.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Stream Label Configuration"
	LabelKeys []string `json:"labelKeys,omitempty"`

	// TenantKey is the tenant for the logs. This supports vector's template syntax to allow dynamic per-event values.
	//
	// The TenantKey can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tenant Key",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	TenantKey string `json:"tenantKey,omitempty"`
}

type SplunkTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	//
	// +kubebuilder:validation:Enum:=none;gzip
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression"
	Compression string `json:"compression,omitempty"`
}

// SplunkAuthentication contains configuration for authenticating requests to a Splunk output.
type SplunkAuthentication struct {
	// Token points to the secret containing the Splunk HEC token used for authenticating requests.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Splunk HEC Token"
	Token *SecretReference `json:"token"`
}

// Splunk Deliver log data to Splunk’s HTTP Event Collector
// Provides optional extra properties for `type: splunk_hec` ('splunk_hec_logs' after Vector 0.23
type Splunk struct {
	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *SplunkAuthentication `json:"authentication"`

	// Tuning specs tuning for the output
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *SplunkTuningSpec `json:"tuning,omitempty"`

	URLSpec `json:",inline"`

	// Index is the index for the logs. This supports template syntax to allow dynamic per-event values.
	//
	// The Index can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Index",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Index string `json:"index,omitempty"`
}

// SyslogRFCType sets which RFC the generated messages conform to.
//
// +kubebuilder:validation:Enum:=RFC3164;RFC5424
type SyslogRFCType string

const (
	SyslogRFC3164 SyslogRFCType = "RFC3164"
	SyslogRFC5424 SyslogRFCType = "RFC5424"
)

type SyslogTuningSpec struct {
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Delivery Mode"
	DeliveryMode DeliveryMode `json:"deliveryMode,omitempty"`
}

// Syslog provides optional extra properties for output type `syslog`
type Syslog struct {

	// An absolute URL, with a scheme. Valid schemes are: `tcp`, `tls`, `udp` and `udps`
	// For example, to send syslog records using secure UDP:
	//     url: udps://syslog.example.com:1234
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="isURL(self)", message="invalid URL"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Destination URL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	URL string `json:"url"`

	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Syslog RFC"
	RFC SyslogRFCType `json:"rfc"`

	// Severity to set on outgoing syslog records.
	//
	// Severity values are defined in https://tools.ietf.org/html/rfc5424#section-6.2.1
	//
	// The value can be a decimal integer or one of these case-insensitive keywords:
	//
	//     Emergency Alert Critical Error Warning Notice Informational Debug
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Severity",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Severity string `json:"severity,omitempty"`

	// Facility to set on outgoing syslog records.
	//
	// Facility values are defined in https://tools.ietf.org/html/rfc5424#section-6.2.1.
	//
	// The value can be a decimal integer. Facility keywords are not standardized,
	// this API recognizes at least the following case-insensitive keywords
	// (defined by https://en.wikipedia.org/wiki/Syslog#Facility_Levels):
	//
	//     kernel user mail daemon auth syslog lpr news
	//     uucp cron authpriv ftp ntp security console solaris-cron
	//     local0 local1 local2 local3 local4 local5 local6 local7
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Facility",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Facility string `json:"facility,omitempty"`

	// PayloadKey specifies record field to use as payload. This supports template syntax to allow dynamic per-event values.
	//
	// The PayloadKey must be a single field path encased in single curly brackets `{}`.
	//
	// Field paths must only contain alphanumeric and underscores. Any field with other characters must be quoted.
	//
	// If left empty, Syslog will use the whole message as the payload key
	//
	// Example:
	//
	//  1. {.bar}
	//
	//  2. {.foo.bar.baz}
	//
	//  3. {.foo.bar."baz/with/slashes"}
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^\{(\.[a-zA-Z0-9_]+|\."[^"]+")(\.[a-zA-Z0-9_]+|\."[^"]+")*\}$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Payload Key",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	PayloadKey string `json:"payloadKey,omitempty"`

	// AppName is APP-NAME part of the syslog-msg header.
	//
	// AppName needs to be specified if using rfc5424. The maximum length of the final values is truncated to 48
	// This supports template syntax to allow dynamic per-event values.
	//
	// The AppName can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="App Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	// TODO: DETERMIN HOW to default the app name that isnt based on fluentd assumptions of "tag" when this is empty
	AppName string `json:"appName,omitempty"`

	// ProcId is PROCID part of the syslog-msg header. This supports template syntax to allow dynamic per-event values.
	//
	// The ProcId can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// ProcId needs to be specified if using rfc5424. The maximum length of the final values is truncated to 128
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="PROCID",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	ProcId string `json:"procId,omitempty"`

	// MsgId is MSGID part of the syslog-msg header. This supports template syntax to allow dynamic per-event values.
	//
	// The MsgId can be a combination of static and dynamic values consisting of field paths followed by `||` followed by another field path or a static value.
	//
	// A dynamic value is encased in single curly brackets `{}` and MUST end with a static fallback value separated with `||`.
	//
	// Static values can only contain alphanumeric characters along with dashes, underscores, dots and forward slashes.
	//
	// Example:
	//
	//  1. foo-{.bar||"none"}
	//
	//  2. {.foo||.bar||"missing"}
	//
	//  3. foo.{.bar.baz||.qux.quux.corge||.grault||"nil"}-waldo.fred{.plugh||"none"}
	//
	// MsgId needs to be specified if using rfc5424.  The maximum length of the final values is truncated to 32
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^(([a-zA-Z0-9-_.\/])*(\{(\.[a-zA-Z0-9_]+|\."[^"]+")+((\|\|)(\.[a-zA-Z0-9_]+|\.?"[^"]+")+)*\|\|"[^"]*"\})*)*$`
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="MSGID",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	MsgId string `json:"msgId,omitempty"`

	// Enrichment is an additional modification the log message before forwarding it to the receiver
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Enrichment Type",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Enrichment EnrichmentType `json:"enrichment,omitempty"`

	// Tuning specs tuning for the output
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *SyslogTuningSpec `json:"tuning,omitempty"`
}

// +kubebuilder:validation:Enum:=None;KubernetesMinimal
type EnrichmentType string

const (
	// EnrichmentTypeNone add no additional enrichment to the record
	EnrichmentTypeNone EnrichmentType = "None"

	// EnrichmentTypeKubernetesMinimal adds namespace_name, pod_name, and collector_name to the beginning of the message
	// body (e.g. namespace_name=myproject, container_name=server, pod_name=pod-123, message={"foo":"bar"}).  This may
	// result in the message body being an invalid JSON structure
	EnrichmentTypeKubernetesMinimal EnrichmentType = "KubernetesMinimal"
)

type OTLPTuningSpec struct {
	BaseOutputTuningSpec `json:",inline"`

	// Compression causes data to be compressed before sending over the network.
	// It is an error if the compression type is not supported by the output.
	//
	// +kubebuilder:validation:Enum:=gzip;none
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Compression string `json:"compression,omitempty"`
}

// OTLP defines configuration for sending logs via OTLP using OTEL semantic conventions
// https://opentelemetry.io/docs/specs/otlp/#otlphttp
type OTLP struct {
	// URL to send log records to.
	//
	// An absolute URL, with a valid http scheme. Must terminate with `/v1/logs`
	//
	// Basic TLS is enabled if the URL scheme requires it (for example 'https').
	// The 'username@password' part of `url` is ignored.
	//
	// +kubebuilder:validation:Pattern:=`^(https?):\/\/\S+\/v1\/logs$`
	// +kubebuilder:validation:XValidation:rule="isURL(self)", message="invalid URL"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Destination URL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	URL string `json:"url"`

	// Authentication sets credentials for authenticating the requests.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Authentication Options"
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`

	// Tuning specs tuning for the output
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tuning Options"
	Tuning *OTLPTuningSpec `json:"tuning,omitempty"`
}
