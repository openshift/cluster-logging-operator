package runtime

import (
	scheduling "k8s.io/api/scheduling/v1"
)

//NewPriorityClass is a constructor to create a PriorityClass
func NewPriorityClass(name string, priorityValue int32, globalDefault bool, description string) *scheduling.PriorityClass {
	pc := &scheduling.PriorityClass{
		Value:         priorityValue,
		GlobalDefault: globalDefault,
		Description:   description,
	}
	Initialize(pc, "", name)
	return pc
}
