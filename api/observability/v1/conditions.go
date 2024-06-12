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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Condition metav1.Condition
type ConditionStatus = metav1.ConditionStatus

const (
	ConditionTrue  = metav1.ConditionTrue
	ConditionFalse = metav1.ConditionFalse

	// Ready indicates the service is ready.
	//
	// Ready=True means the operands are running and providing some service.
	// See the Degraded condition to distinguish full service from partial service.
	//
	// Ready=False means the operands cannot provide any service, and
	// the operator cannot recover without some external change. Either
	// the spec is invalid, or there is some environmental problem that is
	// outside of the the operator's control.
	//
	// Ready=Unknown means the operator is in transition.
	//
	ConditionReady string = "Ready"

	ValidationCondition string = "Validation"

	ConditionMigrate string = "Migrate"

	ConditionAuthorized string = GroupName + "/Authorized"

	ReasonServiceAccountDoesNotExist = "ServiceAccountDoesNotExist"

	// ReasonClusterRoleMissing means the collector serviceAccount is missing one or more clusterRoles needed to collect a log_type
	ReasonClusterRoleMissing = "ClusterRoleMissing"

	// ReasonClusterRolesExist means the collector serviceAccount is bound to all the cluster roles needed to collect a log_type
	ReasonClusterRolesExist = "ClusterRolesExist"

	// ReasonValidationFailure is used when a validation failed.
	ReasonValidationFailure string = "ValidationFailure"

	// ReasonMissingSpec applies when a type is specified without a defined spec (e.g. type application without obs.Application)
	ReasonMissingSpec = "MissingSpec"

	// ReasonMissingSources applies when an input type is spec'd without sources
	ReasonMissingSources = "MissingSources"

	// ReasonInvalidGlob when a namespace or container include/exclude is spec'd with an invalid glob pattern
	ReasonInvalidGlob = "InvalidGlob"

	// ReasonSecretNotFound when a secret is spec'd for an input or output and was not found
	ReasonSecretNotFound = "SecretNotFound"

	// ReasonSecretKeyNotFound when the key for a secret is spec'd for an input or output and was not found as a key in the secret
	ReasonSecretKeyNotFound = "SecretKeyNotFound"

	// ReasonValueEmpty when the value of the key for a secret or configmap is empty
	ReasonValueEmpty = "ValueEmpty"

	// ReasonConfigMapNotFound when a configmap is spec'd for an input or output and was not found
	ReasonConfigMapNotFound = "ConfigMapNotFound"

	// ReasonConfigMapKeyNotFound when the key for a configmap is spec'd for an input or output and was not found as a key in the configmap
	ReasonConfigMapKeyNotFound = "ConfigMapKeyNotFound"

	// ReasonFilterPruneHostname exists when a sink is referenced that does not allow pruning the hostname from the record (i.e. GCL)
	ReasonFilterPruneHostname = "FilterPruneHostname"

	ReasonPipelineInputRefNotFound  = "ReasonPipelineInputRefNotFound"
	ReasonPipelineOutputRefNotFound = "ReasonPipelineOutputRefNotFound"
	ReasonPipelineFilterRefNotFound = "ReasonPipelineFilterRefNotFound"

	ReasonMigrateOutput string = "Migrate"
)
