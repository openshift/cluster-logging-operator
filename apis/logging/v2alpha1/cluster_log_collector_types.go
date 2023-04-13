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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterLogCollector configures the log collector used by a ClusterLogForwarder.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=lc
// +kubebuilder:printcolumn:name="Management State",JSONPath=".spec.managementState",type=string
// +operator-sdk:csv:customresourcedefinitions:displayName="Cluster Logging",resources={{Pod,v1},{Deployment,v1},{ReplicaSet,v1},{ConfigMap,v1},{Service,v1},{Route,v1},{CronJob,v1},{Role,v1},{RoleBinding,v1},{ServiceAccount,v1},{ServiceMonitor,v1},{persistentvolumeclaims,v1}}
type ClusterLogCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogCollectorSpec   `json:"spec,omitempty"`
	Status LogCollectorStatus `json:"status,omitempty"`
}
