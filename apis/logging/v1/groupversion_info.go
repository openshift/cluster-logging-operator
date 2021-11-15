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

// Package v1 contains API Schema definitions for the logging v1 API group
// +kubebuilder:object:generate=true
// +groupName=logging.openshift.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "logging.openshift.io", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// +kubebuilder:rbac:groups=logging.openshift.io,namespace=openshift-logging,resources=*,verbs=*
// +kubebuilder:rbac:groups=core,namespace=openshift-logging,resources=pods;services;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts;serviceaccounts/finalizers;services/finalizers,verbs=*
// +kubebuilder:rbac:groups=apps,namespace=openshift-logging,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
// +kubebuilder:rbac:groups=route.openshift.io,namespace=openshift-logging,resources=routes;routes/custom-host,verbs="*"
// +kubebuilder:rbac:groups=batch,namespace=openshift-logging,resources=cronjobs,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,namespace=openshift-logging,resources=roles;rolebindings,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,namespace=openshift-logging,resources=prometheusrules;servicemonitors,verbs=*
// +kubebuilder:rbac:groups=apps,namespace=openshift-logging,resourceNames=cluster-logging-operator,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=console.openshift.io,resources=consoleexternalloglinks,verbs=*
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=*
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=*
// +kubebuilder:rbac:groups=oauth.openshift.io,resources=oauthclients,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=*
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies;infrastructures,verbs=get;list;watch
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=create
// +kubebuilder:rbac:groups=core,resources=pods;namespaces;services;services/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=security.openshift.io,namespace=openshift-logging,resources=securitycontextconstraints,resourceNames=log-collector-scc,verbs=use
