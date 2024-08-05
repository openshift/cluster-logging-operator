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

	// ConditionUnknown means unable to determine the condition
	ConditionUnknown = metav1.ConditionUnknown

	// ConditionTypeAuthorized identifies the state of authorization for the service
	ConditionTypeAuthorized = GroupName + "/Authorized"

	ConditionTypeLogLevel = GroupName + "/LogLevel"

	// ConditionTypeReady indicates the service is ready.
	//
	// Ready=True means the operands are running and providing some service.
	// Ready=False means the operands cannot provide any service, and
	// the operator cannot recover without some external change. Either
	// the spec is invalid, or there is some environmental problem that is
	// outside the operator's control.
	ConditionTypeReady string = "Ready"

	// ConditionTypeValid identifies the state of validation for the service
	ConditionTypeValid = GroupName + "/Valid"

	// ConditionTypeValidInputPrefix prefixes a named input to identify its validation state
	ConditionTypeValidInputPrefix = GroupName + "/ValidInput"

	// ConditionTypeValidOutputPrefix prefixes a named output to identify its validation state
	ConditionTypeValidOutputPrefix = GroupName + "/ValidOutput"

	// ConditionTypeValidPipelinePrefix prefixes a named pipeline to identify its validation state
	ConditionTypeValidPipelinePrefix = GroupName + "/ValidPipeline"

	// ConditionTypeValidFilterPrefix prefixes a named filter to identify its validation state
	ConditionTypeValidFilterPrefix = GroupName + "/ValidFilter"

	// ReasonClusterRolesExist means the collector serviceAccount is bound to all the cluster roles needed to collect a log_type
	ReasonClusterRolesExist = "ClusterRolesExist"

	// ReasonClusterRoleMissing means the collector serviceAccount is missing one or more clusterRoles needed to collect a log_type
	ReasonClusterRoleMissing = "ClusterRoleMissing"

	// ReasonDeploymentError means an error occurred trying to deploy the collector or some related component
	ReasonDeploymentError = "DeploymentError"

	// ReasonInitializationFailed indicates a failure initializing the reconciliation context
	ReasonInitializationFailed = "InitializationFailed"

	// ReasonFailureToRemoveStaleWorkload indicates a failure removing a stale workload after the deployment type changes
	ReasonFailureToRemoveStaleWorkload = "FailureToRemoveStaleWorkload"

	// ReasonManagementStateUnmanaged is used when the workload is in an Unmanaged state
	ReasonManagementStateUnmanaged = "ManagementStateUnmanaged"

	// ReasonMissingSpec applies when a type is specified without a defined spec (e.g. type application without obs.Application)
	ReasonMissingSpec = "MissingSpec"

	// ReasonLogLevelSupported indicates the support for the log level annotation value
	ReasonLogLevelSupported = "LogLevelSupported"

	// ReasonReconciliationComplete when the operator has initialized, validated, and deployed the resources for the workload
	ReasonReconciliationComplete = "ReconciliationComplete"

	// ReasonServiceAccountDoesNotExist when the ServiceAccount is not found
	ReasonServiceAccountDoesNotExist = "ServiceAccountDoesNotExist"

	// ReasonServiceAccountCheckFailure when there is a failure retrieving the ServiceAccount
	ReasonServiceAccountCheckFailure = "ServiceAccountCheckFailure"

	// ReasonValidationSuccess is used when validation succeeds.
	ReasonValidationSuccess = "ValidationSuccess"

	// ReasonValidationFailure is used when validation fails.
	ReasonValidationFailure = "ValidationFailure"

	// ReasonUnknownState is used when the operator can not determine the state of the deployment
	ReasonUnknownState = "UnknownState"
)
