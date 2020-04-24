package v1

import (
	"encoding/json"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType names a tri-state indicator (true, false, unknown) of the state
// of an object. An object may have multiple conditions of different types,
// but at most one condition value of each type.
type ConditionType string

const (
	// Ready object is providing service. All objects support this condition type.
	ConditionReady ConditionType = "Ready"
	// Degraded object can provide some service, but not everything requested in the spec.
	ConditionDegraded ConditionType = "Degraded"
)

// ConditionReason is an optional indicator of why the condition is in the state it is.
type ConditionReason string

const (
	// Invalid spec is ill-formed in some way, or contains unknown references.
	ReasonInvalid ConditionReason = "Invalid"
	// MissingResources spec refers to resources that can't be located.
	ReasonMissingResource ConditionReason = "MissingResource"
	// Unused spec defines a valid object but it is never used.
	ReasonUnused ConditionReason = "Unused"
	// Connecting object is unready because a connection is in progress.
	ReasonConnecting ConditionReason = "Connecting"
)

// Status is a package alias for corev1.ConditionStatus
type Status = corev1.ConditionStatus

type Condition struct {
	// Type of the condition
	Type ConditionType `json:"type"`

	// Status of the condition: must be "True", "False", or "Unknown"
	Status Status `json:"status"`

	// Reason is an optional CamelCase word describing the reason for the status.
	Reason ConditionReason `json:"reason,omitempty"`

	// Message is a human-readable description.
	Message string `json:"message,omitempty"`

	// LastTransitionTime is the last time this condition changed
	// in RFC3339 format, "2006-01-02T15:04:05Z07:00"
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

func (c Condition) IsTrue() bool    { return c.Status == corev1.ConditionTrue }
func (c Condition) IsFalse() bool   { return c.Status == corev1.ConditionFalse }
func (c Condition) IsUnknown() bool { return !c.IsTrue() && !c.IsFalse() }

var statusValue = map[bool]corev1.ConditionStatus{false: corev1.ConditionFalse, true: corev1.ConditionTrue}

// ConditionStatusOf returns the Status enum string for a boolean value.
func ConditionStatusOf(b bool) corev1.ConditionStatus { return statusValue[b] }

// NewCondition creates a condition with Type, Status, Reason, a fmt-style message
// and Now() as the LastTransitionTime
func NewCondition(t ConditionType, status bool, r ConditionReason, format ...interface{}) Condition {
	c := Condition{
		Type:               t,
		Reason:             r,
		Status:             statusValue[status],
		LastTransitionTime: metav1.Now(),
	}
	if len(format) > 0 {
		c.Message = fmt.Sprintf(format[0].(string), format[1:]...)
	}
	return c
}

// Conditions is a set of condition instances.
//
// It allows lookup and ensures uniqueness by type, but marshals as an array
// with deterministic ordering.
type Conditions map[ConditionType]Condition

func NewConditions(conds ...Condition) Conditions {
	cs := Conditions{}
	for _, c := range conds {
		cs[c.Type] = c
	}
	return cs
}

func (cs Conditions) Get(t ConditionType) Condition { return cs[t] }
func (cs Conditions) Has(t ConditionType) bool      { _, ok := cs[t]; return ok }

func (cs *Conditions) Set(c Condition) {
	if *cs == nil {
		*cs = Conditions{}
	}
	(*cs)[c.Type] = c
}

func (cs Conditions) SetNew(t ConditionType, status bool, r ConditionReason, format ...interface{}) {
	cs.Set(NewCondition(t, status, r, format...))
}

// Conditions marshals as an array.
func (cs Conditions) MarshalJSON() ([]byte, error) {
	list := make([]Condition, 0, len(cs))
	for _, c := range cs {
		list = append(list, c)
	}
	sort.Slice(list, func(a, b int) bool {
		return list[a].Type < list[b].Type
	})
	return json.Marshal(list)
}

type NamedConditions map[string]Conditions

func (nc NamedConditions) Has(name string) bool { return len(nc[name]) > 0 }

func (nc NamedConditions) Get(name string) Conditions {
	if nc[name] == nil {
		nc[name] = Conditions{}
	}
	return nc[name]
}
