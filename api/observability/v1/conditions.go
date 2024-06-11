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

const (

	// ConditionTrue means the condition is met
	ConditionTrue = metav1.ConditionTrue

	// ConditionFalse means the condition is not met
	ConditionFalse = metav1.ConditionFalse

	// ConditionAuthorized identifies the state of authorization for the service
	ConditionAuthorized string = GroupName + "/Authorized"

	ConditionMigrate string = "Migrate"

	// ConditionReady indicates the service is ready.
	//
	// Ready=True means the operands are running and providing some service.
	// Ready=False means the operands cannot provide any service, and
	// the operator cannot recover without some external change. Either
	// the spec is invalid, or there is some environmental problem that is
	// outside of the operator's control.
	ConditionReady string = "Ready"

	// ConditionValidInputPrefix prefixes a named input to identify its validation state
	ConditionValidInputPrefix = GroupName + "/ValidInput"

	// ConditionValidOutputPrefix prefixes a named output to identify its validation state
	ConditionValidOutputPrefix = GroupName + "/ValidOutput"

	// ConditionValidPipelinePrefix prefixes a named pipeline to identify its validation state
	ConditionValidPipelinePrefix = GroupName + "/ValidPipeline"

	// ConditionValidFilterPrefix prefixes a named filter to identify its validation state
	ConditionValidFilterPrefix = GroupName + "/ValidFilter"

	// ReasonClusterRolesExist means the collector serviceAccount is bound to all the cluster roles needed to collect a log_type
	ReasonClusterRolesExist = "ClusterRolesExist"

	// ReasonClusterRoleMissing means the collector serviceAccount is missing one or more clusterRoles needed to collect a log_type
	ReasonClusterRoleMissing = "ClusterRoleMissing"

	ReasonMigrateOutput string = "Migrate"

	// ReasonMissingSpec applies when a type is specified without a defined spec (e.g. type application without obs.Application)
	ReasonMissingSpec = "MissingSpec"

	ReasonServiceAccountDoesNotExist = "ServiceAccountDoesNotExist"

	// ReasonValidationSuccess is used when validation succeeds.
	ReasonValidationSuccess = "ValidationSuccess"

	// ReasonValidationFailure is used when validation fails.
	ReasonValidationFailure string = "ValidationFailure"
)
