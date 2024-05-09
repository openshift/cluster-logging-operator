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
	"github.com/openshift/cluster-logging-operator/internal/status"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterLoggingSpec defines the desired state of ClusterLogging
// +k8s:openapi-gen=true
type ClusterLoggingSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator
	//
	// +kubebuilder:validation:Enum:=Managed;Unmanaged
	// +optional
	ManagementState ManagementState `json:"managementState,omitempty"`

	// Specification of the Visualization component for the cluster
	//
	// +nullable
	// +optional
	Visualization *VisualizationSpec `json:"visualization,omitempty"`

	// Specification of the Log Storage component for the cluster
	//
	// +nullable
	// +optional
	LogStore *LogStoreSpec `json:"logStore,omitempty"`

	// Specification of the Collection component for the cluster
	//
	// +nullable
	Collection *CollectionSpec `json:"collection,omitempty"`

	// Deprecated. Specification of the Curation component for the cluster
	// This component was specifically for use with Elasticsearch and was
	// replaced by index management spec
	//
	// +nullable
	// +optional
	// +deprecated
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:hidden"}
	Curation *CurationSpec `json:"curation,omitempty"`

	// Deprecated. Specification for Forwarder component for the cluster
	// See spec.collection.fluentd
	//
	// +nullable
	// +optional
	// +deprecated
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:hidden"}
	Forwarder *ForwarderSpec `json:"forwarder,omitempty"`
}

// ClusterLoggingStatus defines the observed state of ClusterLogging
// +k8s:openapi-gen=true
type ClusterLoggingStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +optional
	Visualization VisualizationStatus `json:"visualization"`

	// +optional
	LogStore LogStoreStatus `json:"logStore"`

	// Deprecated.
	// +optional
	// +deprecated
	// +nullable
	Collection *CollectionStatus `json:"collection,omitempty"`

	// +optional
	// +deprecated
	Curation *CurationStatus `json:"curation,omitempty"`

	// +optional
	Conditions status.Conditions `json:"conditions,omitempty"`
}

// This is the struct that will contain information pertinent to Log visualization (Kibana)
type VisualizationSpec struct {

	// The type of Visualization to configure
	//
	// +kubebuilder:validation:Enum=ocp-console;kibana
	Type VisualizationType `json:"type"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Visualization Node Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:nodeSelector"}
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Define the tolerations the Pods will accept
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Visualization Pod Tolerations",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Toleration"}
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// Specification of the Kibana Visualization component
	//
	// +deprecated
	// +nullable
	// +optional
	Kibana *KibanaSpec `json:"kibana,omitempty"`

	// OCPConsole is the specification for the OCP console plugin
	//
	// +nullable
	// +optional
	OCPConsole *OCPConsoleSpec `json:"ocpConsole,omitempty"`
}

type KibanaSpec struct {
	// The resource requirements for Kibana
	//
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kibana Resource Requirements",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +deprecated
	// +nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kibana Node Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:nodeSelector"}
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Define the tolerations the Pods will accept
	//
	// +deprecated
	// +nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kibana Tolerations",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Toleration"}
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// Number of instances to deploy for a Kibana deployment
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kibana Size",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:podCount"}
	Replicas *int32 `json:"replicas,omitempty"`

	// Specification of the Kibana Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

type OCPConsoleSpec struct {

	// LogsLimit is the max number of entries returned for a query.
	//
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="OCP Console Log Limit",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:ocpConsoleLogLimit"}
	LogsLimit int `json:"logsLimit,omitempty"`

	// Timeout is the max duration before a query timeout
	//
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="OCP Console Query Timeout",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:ocpConsoleTimeout"}
	Timeout FluentdTimeUnit `json:"timeout,omitempty"`
}

type ProxySpec struct {
	// +nullable
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

// The LogStoreSpec contains information about how logs are stored.
type LogStoreSpec struct {
	// The Type of Log Storage to configure. The operator currently supports either using ElasticSearch
	// managed by elasticsearch-operator or Loki managed by loki-operator (LokiStack) as a default log store.
	//
	// When using ElasticSearch as a log store this operator also manages the ElasticSearch deployment.
	//
	// When using LokiStack as a log store this operator does not manage the LokiStack, but only creates
	// configuration referencing an existing LokiStack deployment. The user is responsible for creating and
	// managing the LokiStack himself.
	//
	// +kubebuilder:validation:Enum=elasticsearch;lokistack
	// +kubebuilder:default:=lokistack
	Type LogStoreType `json:"type"`

	// Specification of the Elasticsearch Log Store component
	// +deprecated
	Elasticsearch *ElasticsearchSpec `json:"elasticsearch,omitempty"`

	// LokiStack contains information about which LokiStack to use for log storage if Type is set to LogStoreTypeLokiStack.
	//
	// The cluster-logging-operator does not create or manage the referenced LokiStack.
	LokiStack LokiStackStoreSpec `json:"lokistack,omitempty"`

	// Retention policy defines the maximum age for an Elasticsearch index after which it should be deleted
	//
	// +nullable
	// +optional
	// +deprecated
	RetentionPolicy *RetentionPoliciesSpec `json:"retentionPolicy,omitempty"`
}

type RetentionPoliciesSpec struct {
	// +nullable
	App *RetentionPolicySpec `json:"application,omitempty"`

	// +nullable
	Infra *RetentionPolicySpec `json:"infra,omitempty"`

	// +nullable
	Audit *RetentionPolicySpec `json:"audit,omitempty"`
}

type RetentionPolicySpec struct {
	// +optional
	MaxAge elasticsearch.TimeUnit `json:"maxAge"`

	// How often to run a new prune-namespaces job
	// +optional
	PruneNamespacesInterval elasticsearch.TimeUnit `json:"pruneNamespacesInterval"`

	// The per namespace specification to delete documents older than a given minimum age
	// +optional
	Namespaces []elasticsearch.IndexManagementDeleteNamespaceSpec `json:"namespaceSpec,omitempty"`

	// The threshold percentage of ES disk usage that when reached, old indices should be deleted (e.g. 75)
	// +optional
	DiskThresholdPercent int64 `json:"diskThresholdPercent,omitempty"`
}

type ElasticsearchSpec struct {
	// The resource requirements for Elasticsearch
	//
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Elasticsearch Resource Requirements",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Number of nodes to deploy for Elasticsearch
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Elasticsearch Size",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:podCount"}
	NodeCount int32 `json:"nodeCount,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Elasticsearch Node Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:nodeSelector"}
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// The storage specification for Elasticsearch data nodes
	//
	// +nullable
	// +optional
	Storage elasticsearch.ElasticsearchStorageSpec `json:"storage,omitempty"`

	// +optional
	RedundancyPolicy elasticsearch.RedundancyPolicyType `json:"redundancyPolicy,omitempty"`

	// Specification of the Elasticsearch Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

// LokiStackStoreSpec is used to set up cluster-logging to use a LokiStack as logging storage.
// It points to an existing LokiStack in the same namespace.
type LokiStackStoreSpec struct {
	// Name of the LokiStack resource.
	//
	// +required
	Name string `json:"name"`
}

// This is the struct that will contain information pertinent to Log and event collection
type CollectionSpec struct {

	// TODO make type required in v2 once Logs is removed. For now assume default which is vector

	// The type of Log Collection to configure
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Implementation",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:fluentd","urn:alm:descriptor:com.tectonic.ui:select:vector"}
	// +kubebuilder:validation:Optional
	Type LogCollectionType `json:"type"`

	// Deprecated. Specification of Log Collection for the cluster
	// See spec.collection
	// +nullable
	// +optional
	// +deprecated
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:hidden"}
	Logs *LogCollectionSpec `json:"logs,omitempty"`

	// CollectorSpec is the common specification that applies to any collector
	// +nullable
	// +optional
	CollectorSpec `json:",inline"`

	// Fluentd represents the configuration for forwarders of type fluentd.
	// +nullable
	// +optional
	Fluentd *FluentdForwarderSpec `json:"fluentd,omitempty"`
}

// Specification of Log Collection for the cluster
// See spec.collection
// +deprecated
type LogCollectionSpec struct {
	// The type of Log Collection to configure
	Type LogCollectionType `json:"type"`

	// Specification of the Fluentd Log Collection component
	CollectorSpec `json:"fluentd,omitempty"`
}

type EventCollectionSpec struct {
	Type EventCollectionType `json:"type"`
}

// CollectorSpec is spec to define scheduling and resources for a collector
type CollectorSpec struct {
	// The resource requirements for the collector
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Resource Requirements",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Node Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:ConfigMap"}
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Define the tolerations the Pods will accept
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Pod Tolerations",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Toleration"}
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}

// This is the struct that will contain information pertinent to Log curation (Curator)
type CurationSpec struct {
	// The kind of curation to configure
	Type CurationType `json:"type"`

	// The specification of curation to configure
	CuratorSpec `json:"curator,omitempty"`
}

type CuratorSpec struct {
	// The resource requirements for Curator
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// The cron schedule that the Curator job is run. Defaults to "30 3 * * *"
	Schedule string `json:"schedule"`
}

// ForwarderSpec contains global tuning parameters for specific forwarder implementations.
// This field is not required for general use, it allows performance tuning by users
// familiar with the underlying forwarder technology.
// Currently supported: `fluentd`.
type ForwarderSpec struct {
	Fluentd *FluentdForwarderSpec `json:"fluentd,omitempty"`
}

// FluentdForwarderSpec represents the configuration for forwarders of type fluentd.
type FluentdForwarderSpec struct {
	InFile *FluentdInFileSpec `json:"inFile,omitempty"`
	Buffer *FluentdBufferSpec `json:"buffer,omitempty"`
}

const (
	// ThrowExceptionAction raises an exception when output buffer is full
	ThrowExceptionAction OverflowActionType = "throw_exception"
	// BlockAction blocks processing inputs when output buffer is full
	BlockAction OverflowActionType = "block"
	// DropOldestChunkAction drops oldest chunk to accept newly incoming chunks
	// when buffer is full
	DropOldestChunkAction OverflowActionType = "drop_oldest_chunk"
)

type OverflowActionType string

const (
	// Flush one chunk per time key if time is specified as chunk key
	FlushModeLazy FlushModeType = "lazy"
	// Flush chunks per specified time via FlushInterval
	FlushModeInterval FlushModeType = "interval"
	// Flush immediately after events appended to chunks
	FlushModeImmediate FlushModeType = "immediate"
)

type FlushModeType string

const (
	// RetryExponentialBackoff increases wait time exponentially between failures
	RetryExponentialBackoff RetryTypeType = "exponential_backoff"
	// RetryPeriodic to retry sending to output periodically on fixed intervals
	RetryPeriodic RetryTypeType = "periodic"
)

type RetryTypeType string

// FluentdSizeUnit represents fluentd's parameter type for memory sizes.
//
// For datatype pattern see:
// https://docs.fluentd.org/configuration/config-file#supported-data-types-for-values
//
// Notice: The OpenAPI validation pattern is an ECMA262 regular expression
// (See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#properties)
//
// +kubebuilder:validation:Pattern:="^([0-9]+)([kmgtKMGT]{0,1})$"
type FluentdSizeUnit string

// FluentdTimeUnit represents fluentd's parameter type for time.
//
// For data type pattern see:
// https://docs.fluentd.org/configuration/config-file#supported-data-types-for-values
//
// Notice: The OpenAPI validation pattern is an ECMA262 regular expression
// (See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#properties)
// +kubebuilder:validation:Pattern:="^([0-9]+)([smhd]{0,1})$"
type FluentdTimeUnit string

// FluentdInFileSpec represents a subset of fluentd in-tail plugin parameters
// to tune the configuration for all fluentd in-tail inputs.
//
// For general parameters refer to:
// https://docs.fluentd.org/input/tail#parameters
type FluentdInFileSpec struct {
	//ReadLinesLimit represents the number of lines to read with each I/O operation
	// +optional
	ReadLinesLimit int `json:"readLinesLimit"`
}

// FluentdBufferSpec represents a subset of fluentd buffer parameters to tune
// the buffer configuration for all fluentd outputs. It supports a subset of
// parameters to configure buffer and queue sizing, flush operations and retry
// flushing.
//
// For general parameters refer to:
// https://docs.fluentd.org/configuration/buffer-section#buffering-parameters
//
// For flush parameters refer to:
// https://docs.fluentd.org/configuration/buffer-section#flushing-parameters
//
// For retry parameters refer to:
// https://docs.fluentd.org/configuration/buffer-section#retries-parameters
type FluentdBufferSpec struct {
	// ChunkLimitSize represents the maximum size of each chunk. Events will be
	// written into chunks until the size of chunks become this size.
	//
	// +optional
	ChunkLimitSize FluentdSizeUnit `json:"chunkLimitSize"`

	// TotalLimitSize represents the threshold of node space allowed per fluentd
	// buffer to allocate. Once this threshold is reached, all append operations
	// will fail with error (and data will be lost).
	//
	// +optional
	TotalLimitSize FluentdSizeUnit `json:"totalLimitSize"`

	// OverflowAction represents the action for the fluentd buffer plugin to
	// execute when a buffer queue is full. (Default: block)
	//
	// +kubebuilder:validation:Enum:=throw_exception;block;drop_oldest_chunk
	// +optional
	OverflowAction OverflowActionType `json:"overflowAction"`

	// FlushThreadCount reprents the number of threads used by the fluentd buffer
	// plugin to flush/write chunks in parallel.
	//
	// +optional
	FlushThreadCount int32 `json:"flushThreadCount"`

	// FlushMode represents the mode of the flushing thread to write chunks. The mode
	// allows lazy (if `time` parameter set), per interval or immediate flushing.
	//
	// +kubebuilder:validation:Enum:=lazy;interval;immediate
	// +optional
	FlushMode FlushModeType `json:"flushMode"`

	// FlushInterval represents the time duration to wait between two consecutive flush
	// operations. Takes only effect used together with `flushMode: interval`.
	//
	// +optional
	FlushInterval FluentdTimeUnit `json:"flushInterval"`

	// RetryWait represents the time duration between two consecutive retries to flush
	// buffers for periodic retries or a constant factor of time on retries with exponential
	// backoff.
	//
	// +optional
	RetryWait FluentdTimeUnit `json:"retryWait"`

	// RetryType represents the type of retrying flush operations. Flush operations can
	// be retried either periodically or by applying exponential backoff.
	//
	// +kubebuilder:validation:Enum:=exponential_backoff;periodic
	// +optional
	RetryType RetryTypeType `json:"retryType"`

	// RetryMaxInterval represents the maximum time interval for exponential backoff
	// between retries. Takes only effect if used together with `retryType: exponential_backoff`.
	//
	// +optional
	RetryMaxInterval FluentdTimeUnit `json:"retryMaxInterval"`

	// RetryTimeout represents the maximum time interval to attempt retries before giving up
	// and the record is disguarded.  If unspecified, the default will be used
	//
	// +optional
	RetryTimeout FluentdTimeUnit `json:"retryTimeout"`
}

type VisualizationStatus struct {
	// +optional
	KibanaStatus []elasticsearch.KibanaStatus `json:"kibanaStatus,omitempty"`
}

type KibanaStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Deployment string `json:"deployment"`
	// +optional
	ReplicaSets []string `json:"replicaSets"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type LogStoreStatus struct {
	// +optional
	ElasticsearchStatus []ElasticsearchStatus `json:"elasticsearchStatus,omitempty"`
}

type ElasticsearchStatus struct {
	// +optional
	ClusterName string `json:"clusterName"`
	// +optional
	NodeCount int32 `json:"nodeCount"`
	// +optional
	ReplicaSets []string `json:"replicaSets,omitempty"`
	// +optional
	Deployments []string `json:"deployments,omitempty"`
	// +optional
	StatefulSets []string `json:"statefulSets,omitempty"`
	// +optional
	ClusterHealth string `json:"clusterHealth,omitempty"`
	// +optional
	Cluster elasticsearch.ClusterHealth `json:"cluster"`
	// +optional
	Pods map[ElasticsearchRoleType]PodStateMap `json:"pods,omitempty"`
	// +optional
	ShardAllocationEnabled elasticsearch.ShardAllocationState `json:"shardAllocationEnabled"`
	// +optional
	ClusterConditions ElasticsearchClusterConditions `json:"clusterConditions,omitempty"`
	// +optional
	NodeConditions map[string]ElasticsearchClusterConditions `json:"nodeConditions,omitempty"`
}

type CollectionStatus struct {
	// +optional
	Logs LogCollectionStatus `json:"logs,omitempty"`
}

type LogCollectionStatus struct {
	// +optional
	FluentdStatus FluentdCollectorStatus `json:"fluentdStatus,omitempty"`
}

type EventCollectionStatus struct {
}

type FluentdCollectorStatus struct {
	// +optional
	DaemonSet string `json:"daemonSet,omitempty"`
	// +optional
	Nodes map[string]string `json:"nodes,omitempty"`
	// +optional
	Pods PodStateMap `json:"pods,omitempty"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type FluentdNormalizerStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	ReplicaSets []string `json:"replicaSets"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type NormalizerStatus struct {
	// +optional
	FluentdStatus []FluentdNormalizerStatus `json:"fluentdStatus,omitempty"`
}

type CurationStatus struct {
	// +optional
	CuratorStatus []CuratorStatus `json:"curatorStatus,omitempty"`
}

type CuratorStatus struct {
	// +optional
	CronJob string `json:"cronJobs"`
	// +optional
	Schedule string `json:"schedules"`
	// +optional
	Suspended bool `json:"suspended"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type PodStateMap map[PodStateType][]string

type PodStateType string

const (
	PodStateTypeReady    PodStateType = "ready"
	PodStateTypeNotReady PodStateType = "notReady"
	PodStateTypeFailed   PodStateType = "failed"
)

type LogStoreType string

const (
	//  NOTE: update the +kubebuilder:validation:Enum comment on LogStoreSpec.Type if you add values here.
	LogStoreTypeElasticsearch LogStoreType = "elasticsearch"
	LogStoreTypeLokiStack     LogStoreType = "lokistack"
)

type ElasticsearchRoleType string

const (
	ElasticsearchRoleTypeClient ElasticsearchRoleType = "client"
	ElasticsearchRoleTypeData   ElasticsearchRoleType = "data"
	ElasticsearchRoleTypeMaster ElasticsearchRoleType = "master"
)

type VisualizationType string

const (
	VisualizationTypeKibana     VisualizationType = "kibana"
	VisualizationTypeOCPConsole VisualizationType = "ocp-console"
)

type CurationType string

const (
	CurationTypeCurator CurationType = "curator"
)

type LogCollectionType string

const (
	LogCollectionTypeFluentd LogCollectionType = "fluentd"
	LogCollectionTypeVector  LogCollectionType = "vector"
)

func (ct LogCollectionType) IsSupportedCollector() bool {
	return ct == LogCollectionTypeVector
}

type EventCollectionType string

type NormalizerType string

type ManagementState string

const (
	// Managed means that the operator is actively managing its resources and trying to keep the component active.
	// It will only upgrade the component if it is safe to do so
	ManagementStateManaged ManagementState = "Managed"
	// Unmanaged means that the operator will not take any action related to the component
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

const (
	IncorrectCRName     ConditionType = "IncorrectCRName"
	ContainerWaiting    ConditionType = "ContainerWaiting"
	ContainerTerminated ConditionType = "ContainerTerminated"
	Unschedulable       ConditionType = "Unschedulable"
	NodeStorage         ConditionType = "NodeStorage"
	CollectorDeadEnd    ConditionType = "CollectorDeadEnd"
)

// `operator-sdk generate crds` does not allow map-of-slice, must use a named type.
type ClusterConditions []Condition
type ElasticsearchClusterConditions []elasticsearch.ClusterCondition

// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=cl
// +kubebuilder:printcolumn:name="Management State",JSONPath=".spec.managementState",type=string
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// A Red Hat OpenShift Logging instance. ClusterLogging is the Schema for the clusterloggings API
// +operator-sdk:csv:customresourcedefinitions:displayName="Cluster Logging",resources={{Pod,v1},{Deployment,v1},{ReplicaSet,v1},{ConfigMap,v1},{Service,v1},{Route,v1},{CronJob,v1},{Role,v1},{RoleBinding,v1},{ServiceAccount,v1},{ServiceMonitor,v1},{persistentvolumeclaims,v1}}
type ClusterLogging struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of ClusterLogging
	Spec ClusterLoggingSpec `json:"spec,omitempty"`

	// Status defines the observed state of ClusterLogging
	Status ClusterLoggingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ClusterLoggingList contains a list of ClusterLogging
type ClusterLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogging `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogging{}, &ClusterLoggingList{})
}
