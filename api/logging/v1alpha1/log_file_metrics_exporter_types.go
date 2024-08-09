package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogFileMetricExporterSpec defines the desired state of LogFileMetricExporter
type LogFileMetricExporterSpec struct {
	// The resource requirements for the LogFileMetricExporter
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="LogFileMetricExporter Resource Requirements",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="LogFileMetricExporter Node Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:ConfigMap"}
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Define the tolerations the Pods will accept
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="LogFileMetricExporter Pod Tolerations",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Toleration"}
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

const (
	// ReasonValid is used as a reason for the condition indicating that the LogFileMetricExporter is deployed.
	ReasonValid string = "Valid"
	// ReasonInvalid is used as a reason for the condition indicating that there was an issue with the LogFileMetricExporter instance.
	ReasonInvalid string = "Invalid"
)

// LogFileMetricExporterStatus defines the observed state of LogFileMetricExporter
type LogFileMetricExporterStatus struct {
	// Conditions of the Log File Metrics Exporter.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Log File Metrics Exporter Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:logFileMetricsExporterConditions"}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=lfme
// +kubebuilder:object:root=true
// A Log File Metric Exporter instance. LogFileMetricExporter is the Schema for the logFileMetricExporters API
// +operator-sdk:csv:customresourcedefinitions:displayName="Log File Metric Exporter",resources={{Pod,v1}, {Service,v1}, {ServiceMonitor,v1}, {DaemonSet, v1}}
type LogFileMetricExporter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogFileMetricExporterSpec   `json:"spec,omitempty"`
	Status LogFileMetricExporterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// LogFileMetricExporterList contains a list of LogFileMetricExporter
type LogFileMetricExporterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogFileMetricExporter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LogFileMetricExporter{}, &LogFileMetricExporterList{})
}
