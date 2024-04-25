package v1

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/status"

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
	ConditionReady status.ConditionType = "Ready"

	// Degraded indicates partial service is available.
	//
	// Degraded=True means the operands can fulfill some of the `spec`, but not all,
	// even when Ready=True.
	//
	// Degraded=False with Ready=True means the operands are providing full service.
	//
	// Degraded=Unknown means the operator is in transition.
	//
	ConditionDegraded status.ConditionType = "Degraded"

	ValidationCondition status.ConditionType = "Validation"
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

	ValidationFailureReason status.ConditionReason = "ValidationFailure"
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

func (nc NamedConditions) IsAllReady() bool {
	for _, conditions := range nc {
		for _, cond := range conditions {
			if cond.Type == ConditionReady && cond.IsFalse() {
				return false
			}
		}
	}
	return true
}

// Synchronize synchronizes the current NamedCondition with a new NamedCondition.
// This is not the same as simply replacing the NamedCondition: Conditions contain the LastTransitionTime
// field which is left unmodified by Synchronize for noops. Whereas all updates and additions shall use the current
// (= now) timestamp. In short, ignore any timestamp in newNamedCondition, and for noops use the timestamp from nc
// or use time.Now().
func (nc NamedConditions) Synchronize(newNamedCondition NamedConditions) error {
	if nc == nil {
		return fmt.Errorf("cannot operate on a nil map in NamedConditions.Synchronize()")
	}
	for name, newConditions := range newNamedCondition {
		oldConditions, ok := nc[name]
		// If map entry doesn't exist, create it.
		if !ok {
			oldConditions = status.Conditions{}
		}
		// Synchronize oldConditions.Conditions.
		synchronizeConditions(&oldConditions, &newConditions)
		// Write back the value, otherwise this would be a noop.
		nc[name] = oldConditions
	}
	// Delete all map entries in old map which do not exist in new map.
	for name := range nc {
		if _, ok := newNamedCondition[name]; !ok {
			delete(nc, name)
		}
	}
	return nil
}

var CondReady = Condition{Type: ConditionReady, Status: corev1.ConditionTrue}

func CondNotReady(r ConditionReason, format string, args ...interface{}) Condition {
	return NewCondition(ConditionReady, corev1.ConditionFalse, r, format, args...)
}

func CondInvalid(format string, args ...interface{}) Condition {
	return CondNotReady(ReasonInvalid, format, args...)
}

// synchronizeConditions is a helper used by *ClusterLogForwarderStatus.Synchronize and NamedConditions.Synchronize.
func synchronizeConditions(oldConditions, newConditions *status.Conditions) {
	// Set all conditions from newConditions in oldConditions.
	// SetCondition adds (or updates) the set of conditions, and sets the current timestamp in case of an add or update.
	// The timestamp won't be updated if the condition exists and does not need an update.
	for _, cond := range *newConditions {
		oldConditions.SetCondition(cond)
	}
	// Remove any superfluous conditions.
	var conditionsToRemove []status.ConditionType
	for _, oldCond := range *oldConditions {
		if newConditions.GetCondition(oldCond.Type) == nil {
			conditionsToRemove = append(conditionsToRemove, oldCond.Type)
		}
	}
	for _, ctr := range conditionsToRemove {
		oldConditions.RemoveCondition(ctr)
	}
}
