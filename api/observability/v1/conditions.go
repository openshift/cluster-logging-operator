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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
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
)

const (
	// ReasonInvalid is used when the spec is ill-formed in some way, or contains unknown references.
	ReasonInvalid string = "Invalid"

	// ReasonMissingResource is used when the spec refers to resources that can't be located.
	ReasonMissingResource string = "MissingResource"

	// ReasonUnused is used when the spec defines a valid object, but it is never used.
	ReasonUnused string = "Unused"

	// ReasonValidationFailure is used when a validation failed.
	ReasonValidationFailure string = "ValidationFailure"

	ReasonMigrateOutput string = "Migrate"
)

// ConditionMap contains a map of resource names to a list of their conditions.
type ConditionMap map[string][]metav1.Condition
