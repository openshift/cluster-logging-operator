package openshift

import (
	"encoding/json"
	"fmt"
)

type LabelsFilter map[string]string

func NewLabelsFilter(labels map[string]string) LabelsFilter {
	return labels
}

func (labels LabelsFilter) VRL() (string, error) {
	if len(labels) != 0 {
		s, _ := json.Marshal(labels)
		return fmt.Sprintf("._internal.openshift.labels = .openshift.labels = %s", s), nil
	}
	return "", nil
}
