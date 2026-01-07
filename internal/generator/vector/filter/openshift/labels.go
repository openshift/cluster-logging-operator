package openshift

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
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

func New(labels map[string]string, inputs ...string) *transforms.Remap {
	if len(labels) != 0 {
		s, _ := json.Marshal(labels)
		vrl := fmt.Sprintf("._internal.openshift.labels = .openshift.labels = %s", s)
		return transforms.NewRemap(vrl, inputs...)
	}
	return nil
}
