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

import (
	"github.com/openshift/cluster-logging-operator/internal/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogCollector configures the log collector used by a LogForwarder.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=lc
// +kubebuilder:printcolumn:name="Management State",JSONPath=".spec.managementState",type=string
// +operator-sdk:csv:customresourcedefinitions:displayName="Cluster Logging",resources={{Pod,v1},{Deployment,v1},{ReplicaSet,v1},{ConfigMap,v1},{Service,v1},{Route,v1},{CronJob,v1},{Role,v1},{RoleBinding,v1},{ServiceAccount,v1},{ServiceMonitor,v1},{persistentvolumeclaims,v1}}
type LogCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec specifies the desired behavior of LogCollector
	Spec LogCollectorSpec `json:"spec,omitempty"`

	// Status defines the observed state of LogCollector
	Status LogCollectorStatus `json:"status,omitempty"`
}

// LogCollectorList contains a list of LogCollector
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type LogCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogCollector `json:"items"`
}

type LogCollectorSpec struct {
	// ManagementState enable/disable management by the operator.
	// +optional
	ManagementState ManagementState `json:"managementState,omitempty"`

	// Type of underlying collector implementation to deploy.
	//
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Type",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:fluentd","urn:alm:descriptor:com.tectonic.ui:select:vector"}
	Type LogCollectionType `json:"type,omitempty"`

	// Resources defines resource limits for collector containers.
	Resources *LogCollectorResourcesSpec `json:"resources,omitempty"`

	// NodeSelector for scheduling  collector pods.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations for scheduling  collector pods.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Fluentd holds configuration specific to fluentd.
	Fluentd *LogCollectorFluentdSpec `json:"fluentd,omitempty"`

	// Vector holds configuration specific to vector.
	Vector *LogCollectorVectorSpec `json:"vector,omitempty"`
}
type LogCollectorResourcesSpec struct {
	// Collector resource requirements. The collector collects and forwards logs.
	Collector *corev1.ResourceRequirements `json:"collector,omitempty"`

	// MetricExporter resource requirements. The metric exporter montoris log files and provides metrics.
	MetricExporter *corev1.ResourceRequirements `json:"metricExporter,omitempty"`
}

type LogCollectorStatus struct {
	Conditions status.Conditions `json:"conditions,omitempty"`
}

func init() {
	SchemeBuilder.Register((*LogCollector)(nil), (*LogCollectorList)(nil))
}

type LogCollectorVectorSpec struct {
	// Placeholder for future extensions.
}
