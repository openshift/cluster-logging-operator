# Temporary notes on name changes, delete before merging.

# Summary of user impact

The 6.0 API has not changed dramatically from 5.9 - it has been extended dramatically.
Many existing users will be able to cut/paste most of their current configurations.

The areas that may need change are:

Authentication: Fields related to authentication have been grouped and standardized.
Outputs using TLS or other authentication will need to be modified to the new format.
This is a significant change but provides more consistent and complete authentication options.

Tuning: A new 'tuning' section collects configuration related to performance, rather than routing semantics.
Some existing tuning-related fields have moved to this section and require changes. (TODO ref details).

Filters: these have been rationalized and extended. Changes will be required to existing filters
and to the pipelines that use them.

NOTE: The Makefile in this directory provides some automated tools to help summarize and review the changes.

# Rationale

## Upcase to conform to api guide (most did not exist in 5.9)

// +kubebuilder:validation:Enum:=AtLeastOnce;AtMostOnce
type DeliveryMode string

// +kubebuilder:validation:Enum:=None;KubernetesMinimal
type EnrichmentType string

// +kubebuilder:validation:Enum:=Container
type ApplicationSource string

// +kubebuilder:validation:Enum:=Container;Node
type InfrastructureSource string

// +kubebuilder:validation:Enum:=secret;serviceAccount
type BearerTokenFrom string

## Kept lower case - "proper" names

### Values referring to API fields

// +kubebuilder:validation:Enum:=azureMonitor;cloudwatch;elasticsearch;http;kafka;loki;lokiStack;googleCloudLogging;splunk;syslog;otlptype OutputType string
type OutputType string

// +kubebuilder:validation:Enum:=http;syslog
type ReceiverType string

// +kubebuilder:validation:Enum:=secret;serviceAccount
type BearerTokenFrom string

### External identifiers (including Viaq)

// +kubebuilder:validation:Pattern:="^[a-zA-Z0-9][a-zA-Z0-9_]{0,99}$"
LogType string `json:"logType,omitempty"`

// +kubebuilder:validation:Enum:=audit;application;infrastructure;receiver
type InputType string

// +kubebuilder:validation:Enum:=openShiftLabels;detectMultilineException;kubeAPIAudit;parse;prune
type FilterType string

// +kubebuilder:validation:Enum:=gzip;none;snappy;zlib;zstd
// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Compression"
Compression string `json:"compression,omitempty"`

// +kubebuilder:validation:Enum:=billingAccount;folder;project;organization
type GoogleCloudLoggingIDType string

// +kubebuilder:validation:Enum:=awsAccessKey;iamRole
type CloudwatchAuthType string

// +kubebuilder:validation:Enum:=billingAccount;folder;project;organization
type GoogleCloudLoggingIDType string

