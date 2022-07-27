package v1

import (
	"reflect"
	"strings"

	sets "k8s.io/apimachinery/pkg/util/sets"
)

// Reserved input names.
const (
	InputNameApplication    = "application"    // Non-infrastructure container logs.
	InputNameInfrastructure = "infrastructure" // Infrastructure containers and system logs.
	InputNameAudit          = "audit"          // System audit logs.
)

var ReservedInputNames = sets.NewString(InputNameApplication, InputNameInfrastructure, InputNameAudit)

func IsInputTypeName(s string) bool { return ReservedInputNames.Has(s) }

// OutputNameDefault is the Default log store output name and version
const OutputNameDefault = "default"

// DefaultESVersion is the version of ES deployed by default
const DefaultESVersion = 6

// FirstESVersionWithoutType (e.g. v8) is the first version without types
const FirstESVersionWithoutType = 8

// IsReservedOutputName returns true if s is a reserved output name.
func IsReservedOutputName(s string) bool { return s == OutputNameDefault }

// IsOutputTypeName returns true if s capitalized is a field name in OutputTypeSpec
func IsOutputTypeName(s string) bool {
	_, ok := reflect.TypeOf(OutputTypeSpec{}).FieldByName(strings.Title(s)) //nolint:staticcheck
	return ok
}

// Get all subordinate condition messages for condition of type "Ready" and False
// A 'true' Ready condition with a message means some error with pipeline but it is still valid
func (status ClusterLogForwarderStatus) GetReadyConditionMessages() []string {
	var messages = []string{}
	for _, nc := range []NamedConditions{status.Pipelines, status.Inputs, status.Outputs} {
		for _, conds := range nc {
			currCond := conds.GetCondition(ConditionReady)
			if !conds.IsTrueFor(ConditionReady) {
				messages = append(messages, currCond.Message)
				// If a pipeline is "degraded" then it should have an extra error message
			} else if len(conds) > 1 {
				messages = append(messages, conds.GetCondition("Error").Message)
			}
		}
	}
	return messages
}

// IsReady returns true if all of the subordinate conditions are ready.
func (status ClusterLogForwarderStatus) IsReady() bool {
	for _, nc := range []NamedConditions{status.Pipelines, status.Inputs, status.Outputs} {
		for _, conds := range nc {
			if !conds.IsTrueFor(ConditionReady) {
				return false
			}
		}
	}
	return true
}

// RouteMap maps input names to connected outputs or vice-versa.
type RouteMap map[string]sets.String

func (m RouteMap) Insert(k, v string) {
	if m[k] == nil {
		m[k] = sets.NewString()
	}
	m[k].Insert(v)
}

func (m RouteMap) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Routes maps connected input and output names.
type Routes struct {
	ByInput, ByOutput RouteMap
}

func NewRoutes(pipelines []PipelineSpec) Routes {
	r := Routes{
		ByInput:  map[string]sets.String{},
		ByOutput: map[string]sets.String{},
	}
	for _, p := range pipelines {
		for _, inRef := range p.InputRefs {
			for _, outRef := range p.OutputRefs {
				r.ByInput.Insert(inRef, outRef)
				r.ByOutput.Insert(outRef, inRef)
			}
		}
	}
	return r
}

// OutputMap returns a map of names to outputs.
func (spec *ClusterLogForwarderSpec) OutputMap() map[string]*OutputSpec {
	m := map[string]*OutputSpec{}
	for i := range spec.Outputs {
		m[spec.Outputs[i].Name] = &spec.Outputs[i]
	}
	return m
}

// True if spec has a default output.
func (spec *ClusterLogForwarderSpec) HasDefaultOutput() bool {
	_, ok := spec.OutputMap()[OutputNameDefault]
	return ok
}

// InputMap returns a map of input names to InputSpec.
func (spec *ClusterLogForwarderSpec) InputMap() map[string]*InputSpec {
	m := map[string]*InputSpec{}
	for i := range spec.Inputs {
		m[spec.Inputs[i].Name] = &spec.Inputs[i]
	}
	return m
}

// Types returns the set of input types that are used to by the input spec.
func (input *InputSpec) Types() sets.String {
	result := sets.NewString()
	if input.Application != nil {
		result.Insert(InputNameApplication)
	}
	if input.Infrastructure != nil {
		result.Insert(InputNameInfrastructure)
	}
	if input.Audit != nil {
		result.Insert(InputNameAudit)
	}
	return result
}

// HasPolicy returns whether the input spec has flow control policies defined in it.
func (input *InputSpec) HasPolicy() bool {
	if input.Application != nil &&
		(input.Application.ContainerLimit != nil ||
			input.Application.GroupLimit != nil) {
		return true
	}
	return false
}

func (input *InputSpec) GetMaxRecordsPerSecond() int64 {
	if input.Application.ContainerLimit != nil {
		return input.Application.ContainerLimit.MaxRecordsPerSecond
	} else {
		return input.Application.GroupLimit.MaxRecordsPerSecond
	}
}

// HasPolicy returns whether the output spec has flow control policies defined in it.
func (output *OutputSpec) HasPolicy() bool {
	return output.Limit != nil
}

func (output *OutputSpec) GetMaxRecordsPerSecond() int64 {
	return output.Limit.MaxRecordsPerSecond
}
