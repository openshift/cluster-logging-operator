// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/status"
	corev1 "k8s.io/api/core/v1"
)

// Aliases for convenience
type Condition = status.Condition
type ConditionType = status.ConditionType
type ConditionReason = status.ConditionReason
type Conditions = status.Conditions

func NewConditions(c ...Condition) Conditions { return status.NewConditions(c...) }

func NewCondition(t status.ConditionType, s corev1.ConditionStatus, r status.ConditionReason, format string, args ...interface{}) Condition {
	return Condition{Type: t, Status: s, Reason: r, Message: fmt.Sprintf(format, args...)}
}

const (
	// Ready object is providing service. All objects support this condition type.
	ConditionReady status.ConditionType = "Ready"
	// Degraded object can provide some service, but not everything requested in the spec.
	ConditionDegraded status.ConditionType = "Degraded"
)

const (
	// Invalid spec is ill-formed in some way, or contains unknown references.
	ReasonInvalid status.ConditionReason = "Invalid"
	// MissingResources spec refers to resources that can't be located.
	ReasonMissingResource status.ConditionReason = "MissingResource"
	// Unused spec defines a valid object but it is never used.
	ReasonUnused status.ConditionReason = "Unused"
	// Connecting object is unready because a connection is in progress.
	ReasonConnecting status.ConditionReason = "Connecting"
)

// SetCondition returns true if the condition changed or is new.
func SetCondition(cs *status.Conditions, t status.ConditionType, s corev1.ConditionStatus, r status.ConditionReason, format string, args ...interface{}) bool {
	return cs.SetCondition(NewCondition(t, s, r, format, args...))
}

type NamedConditions map[string]status.Conditions

func (nc NamedConditions) Set(name string, cond status.Condition) bool {
	conds := nc[name]
	ret := conds.SetCondition(cond)
	nc[name] = conds
	return ret
}

func (nc NamedConditions) SetCondition(name string, t status.ConditionType, s corev1.ConditionStatus, r status.ConditionReason, format string, args ...interface{}) bool {
	return nc.Set(name, NewCondition(t, s, r, format, args...))
}
