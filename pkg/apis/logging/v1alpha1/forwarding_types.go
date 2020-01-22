package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//LogSourceType is an explicitly defined log source
type LogSourceType string

const (
	LogForwardingKind string = "LogForwarding"

	//LogSourceTypeApp are container logs from non-infra structure containers
	LogSourceTypeApp LogSourceType = "logs-app"

	//LogSourceTypeInfra are logs from infra structure containers or node logs
	LogSourceTypeInfra LogSourceType = "logs-infra"

	//LogSourceTypeAudit are audit logs from the nodes and the k8s and
	// openshift apiservers
	LogSourceTypeAudit LogSourceType = "logs-audit"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// LogForwarding is the Schema for the logforwardings API
// +k8s:openapi-gen=true
type LogForwarding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ForwardingSpec    `json:"spec,omitempty"`
	Status *ForwardingStatus `json:"status,omitempty"`
}

//ForwardingSpec specifies log forwarding pipelines from a defined sources to dest outputs
// +k8s:openapi-gen=true
type ForwardingSpec struct {
	DisableDefaultForwarding bool           `json:"disableDefaultForwarding,omitempty"`
	Outputs                  []OutputSpec   `json:"outputs,omitempty"`
	Pipelines                []PipelineSpec `json:"pipelines,omitempty"`
}

//PipelineSpec is the sources spec to named targets
type PipelineSpec struct {
	Name       string        `json:"name,omitempty"`
	SourceType LogSourceType `json:"inputSource,omitempty"`

	//OutputRefs is a list of  the names of outputs defined by forwarding.outputs
	OutputRefs []string `json:"outputRefs,omitempty"`
}

//OutputSpec specifies destination config for log message endpoints
type OutputSpec struct {
	Type     OutputType        `json:"type,omitempty"`
	Name     string            `json:"name,omitempty"`
	Endpoint string            `json:"endpoint,omitempty"`
	Secret   *OutputSecretSpec `json:"secret,omitempty"`
}

//OutputSecretSpec specifies secrets for pipelines
type OutputSecretSpec struct {

	//Name is the name of the secret to use with this output
	Name string `json:"name"`
}

//OutputType defines the type of endpoint that will receive messages
type OutputType string

const (
	//OutputTypeElasticsearch configures pipeline to send messages to elasticsearch
	OutputTypeElasticsearch OutputType = "elasticsearch"

	//OutputTypeForward configures the pipeline to send messages via Fluent's secure forward
	OutputTypeForward OutputType = "forward"

	//OutputTypeSyslog configures pipeline to send messages to an external syslog server through docebo/fluent-plugin-remote-syslog
	OutputTypeSyslog OutputType = "syslog"
)

//LogForwardingReason The reason for the current state
type LogForwardingReason string

//LogForwardingReasonName is for evaluating the resource name
const LogForwardingReasonName = "ResourceName"

//LogForwardingState is the state of LogForwarding
type LogForwardingState string

const (
	//LogForwardingStateAccepted is when the resources is accepted by the operator
	LogForwardingStateAccepted LogForwardingState = "Accepted"

	//LogForwardingStateRejected is when the resources is rejected by the operator
	LogForwardingStateRejected LogForwardingState = "Rejected"
)

//ForwardingStatus is the status of spec'd forwarding
// +k8s:openapi-gen=true
type ForwardingStatus struct {

	//State is the current state of LogForwarding instance
	State LogForwardingState `json:"state,omitempty"`

	// Reason is a one-word CamelCase reason for the condition's last transition.
	Reason LogForwardingReason `json:"reason,omitempty"`

	// Reason is a one-word CamelCase reason for the condition's last transition.
	Message string `json:"message,omitempty"`

	// LastUpdated represents the last time that the status was updated.
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	//LogSources lists the configured log sources
	LogSources []LogSourceType `json:"sources,omitempty"`

	//Pipelines is the status of the outputs
	Pipelines []PipelineStatus `json:"pipelines,omitempty"`

	//Outputs is the status of the outputs
	Outputs []OutputStatus `json:"outputs,omitempty"`
}

func NewForwardingStatus(state LogForwardingState, reason LogForwardingReason, message string) *ForwardingStatus {
	return &ForwardingStatus{
		State:       state,
		Reason:      reason,
		Message:     message,
		LastUpdated: metav1.Now(),
	}
}

//PipelineStatus is the status of a give pipeline
type PipelineStatus struct {
	//Name of the corresponding pipeline for this status
	Name string `json:"name,omitempty"`

	//State of the corresponding pipeline for this status
	State PipelineState `json:"state,omitempty"`

	Reason PipelineStateReason `json:"reason,omitempty"`

	Message string `json:"message,omitempty"`

	//Reasons for the state of the corresponding pipeline for this status
	Conditions []PipelineCondition `json:"conditions,omitempty"`

	// LastUpdated represents the last time that the status was updated.
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

func NewPipelineStatusNamed(name string) PipelineStatus {
	return PipelineStatus{
		Name:        name,
		LastUpdated: metav1.Now(),
	}
}
func NewPipelineStatus(name string, state PipelineState, reason PipelineStateReason, message string) PipelineStatus {
	return PipelineStatus{
		Name:        name,
		State:       state,
		Reason:      reason,
		Message:     message,
		LastUpdated: metav1.Now(),
	}
}

func (pipelineStatus *PipelineStatus) AddCondition(conditionType PipelineConditionType, reason PipelineConditionReason, message string) {
	pipelineStatus.Conditions = append(pipelineStatus.Conditions, PipelineCondition{
		Type:    conditionType,
		Reason:  reason,
		Status:  corev1.ConditionFalse,
		Message: message,
	})
}

type PipelineStateReason string

const (
	PipelineStateReasonConditionsMet    PipelineStateReason = "ConditionsMet"
	PipelineStateReasonConditionsNotMet PipelineStateReason = "ConditionsNotMet"
)

//PipelineState is the state of a spec'd pipeline
type PipelineState string

const (
	//PipelineStateAccepted is accepted by forwarding and includes all required fields to send messages to all spec'd outputs
	PipelineStateAccepted PipelineState = "Accepted"
	//PipelineStateDegraded is parially accepted by forwarding and includes some of the required fields to send messages to some of the spec'd outputs
	PipelineStateDegraded PipelineState = "Degraded"
	//PipelineStateDropped dropped by forwarding because its missing required fields to send messages to outputs
	PipelineStateDropped PipelineState = "Dropped"
)

type PipelineCondition struct {
	Type    PipelineConditionType   `json:"typ,omitempty"`
	Reason  PipelineConditionReason `json:"reason,omitempty"`
	Status  corev1.ConditionStatus  `json:"status,omitempty"`
	Message string                  `json:"message,omitempty"`
}

type PipelineConditionType string

const (
	//PipelineConditionTypeOutputRef related to verifying outputRefs
	PipelineConditionTypeOutputRef PipelineConditionType = "OutputRef"

	//PipelineConditionTypeSourceType related to verifying sourceTypes
	PipelineConditionTypeSourceType PipelineConditionType = "SourceType"

	//PipelineConditionTypeName related to verifying name
	PipelineConditionTypeName PipelineConditionType = "Name"
)

type PipelineConditionReason string

const (
	PipelineConditionReasonUnrecognizedOutputRef  PipelineConditionReason = "UnrecognizedOutputRef"
	PipelineConditionReasonUnrecognizedSourceType PipelineConditionReason = "UnrecognizedSourceType"
	PipelineConditionReasonUniqueName             PipelineConditionReason = "UniqueName"
	PipelineConditionReasonMissingName            PipelineConditionReason = "MissingName"
	PipelineConditionReasonMissingOutputs         PipelineConditionReason = "MissingOutputRefs"
	PipelineConditionReasonMissingSource          PipelineConditionReason = "MissingSource"
	PipelineConditionReasonReservedNameConflict   PipelineConditionReason = "ReservedNameConflict"
)

type OutputStateReason string

const (
	OutputStateReasonConditionsMet    OutputStateReason = "ConditionsMet"
	OutputStateReasonConditionsNotMet OutputStateReason = "ConditionsNotMet"
)

//OutputStatus of a given output
type OutputStatus struct {
	//Name of the corresponding output for this status
	Name string `json:"name,omitempty"`

	//State of the corresponding output for this status
	State OutputState `json:"state,omitempty"`

	//Reasons for the state of the corresponding output for this status
	Reason OutputStateReason `json:"reason,omitempty"`

	//Message about the corresponding output
	Message string `json:"message,omitempty"`

	//Reasons for the state of the corresponding pipeline for this status
	Conditions []OutputCondition `json:"conditions,omitempty"`

	// LastUpdated represents the last time that the status was updated.
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

func NewOutputStatusNamed(name string) OutputStatus {
	return OutputStatus{
		Name:        name,
		LastUpdated: metav1.Now(),
	}
}

func NewOutputStatus(name string, state OutputState, reason OutputStateReason, message string) OutputStatus {
	return OutputStatus{
		Name:        name,
		State:       state,
		Reason:      reason,
		Message:     message,
		LastUpdated: metav1.Now(),
	}
}

func (outputStatus *OutputStatus) AddCondition(conditionType OutputConditionType, reason OutputConditionReason, message string) {
	outputStatus.Conditions = append(outputStatus.Conditions, OutputCondition{
		Type:    conditionType,
		Reason:  reason,
		Status:  corev1.ConditionFalse,
		Message: message,
	})
}

//OutputState is the state of a spec'd output
type OutputState string

const (
	//OutputStateAccepted means the output is usuable by forwarding and is spec'd with all the required fields
	OutputStateAccepted OutputState = "Accepted"

	//OutputStateDropped means the output is unusuable by forwarding is missing some the required fields
	OutputStateDropped OutputState = "Dropped"
)

type OutputConditionType string

const (
	//OutputConditionTypeEndpoint related to verifying endpoints
	OutputConditionTypeEndpoint OutputConditionType = "Endpoint"

	//OutputConditionTypeSecret related to verifying secrets
	OutputConditionTypeSecret OutputConditionType = "Secet"

	//OutputConditionTypeName related to verifying name
	OutputConditionTypeName OutputConditionType = "Name"

	//OutputConditionTypeType related to verifying type
	OutputConditionTypeType OutputConditionType = "Type"
)

//OutputConditionReason provides a reason for the given state
type OutputConditionReason string

const (
	//OutputConditionReasonNonUniqueName is not unique amoung all defined outputs
	OutputConditionReasonNonUniqueName OutputConditionReason = "NonUniqueName"

	//OutputConditionReasonReservedNameConflict is not unique amoung all defined outputs
	OutputConditionReasonReservedNameConflict OutputConditionReason = "ReservedNameConflict"

	//OutputConditionReasonMissingName is missing a name
	OutputConditionReasonMissingName OutputConditionReason = "MissingName"

	//OutputConditionReasonMissingType is missing a name
	OutputConditionReasonMissingType OutputConditionReason = "MissingType"

	//OutputConditionReasonMissingEndpoint is missing the endpoint information, it is empty, or is an invalid format
	OutputConditionReasonMissingEndpoint OutputConditionReason = "MissingEndpoint"

	//OutputConditionReasonMissingSecretName is missing the name of the secret
	OutputConditionReasonMissingSecretName OutputConditionReason = "MissingSecretName"

	//OutputConditionReasonSecretDoesNotExist for secrets which don't exist
	OutputConditionReasonSecretDoesNotExist OutputConditionReason = "SecretDoesNotExist"

	//OutputConditionReasonSecretMissingSharedKey for secrets which don't exist
	OutputConditionReasonSecretMissingSharedKey OutputConditionReason = "SecretMissingSharedKey"

	//OutputConditionReasonUnrecognizedType has an unrecognized or supported output type
	OutputConditionReasonUnrecognizedType OutputConditionReason = "UnrecognizedType"
)

type OutputCondition struct {
	Type    OutputConditionType    `json:"type,omitempty"`
	Reason  OutputConditionReason  `json:"reason,omitempty"`
	Status  corev1.ConditionStatus `json:"status,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LogForwardingList contains a list of LogForwarding
type LogForwardingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogForwarding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LogForwarding{}, &LogForwardingList{})
}
