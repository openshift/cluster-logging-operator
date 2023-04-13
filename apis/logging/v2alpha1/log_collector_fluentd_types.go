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

package v2alpha1

// LogCollectorFluentdSpec represents the configuration for forwarders of type fluentd.
type LogCollectorFluentdSpec struct {
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

type FluentdCollectorStatus struct {
	// +optional
	DaemonSet string `json:"daemonSet,omitempty"`
	// +optional
	Nodes map[string]string `json:"nodes,omitempty"`
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Fluentd status",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:podStatuses"}
	Pods PodStateMap `json:"pods,omitempty"`
	// +optional
	Conditions map[string][]Condition `json:"clusterCondition,omitempty"`
}

type FluentdNormalizerStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	ReplicaSets []string `json:"replicaSets"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string][]Condition `json:"clusterCondition,omitempty"`
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
	Conditions map[string][]Condition `json:"clusterCondition,omitempty"`
}

type PodStateMap map[PodStateType][]string

type PodStateType string

const (
	PodStateTypeReady    PodStateType = "ready"
	PodStateTypeNotReady PodStateType = "notReady"
	PodStateTypeFailed   PodStateType = "failed"
)
