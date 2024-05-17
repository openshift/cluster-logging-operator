package openshift

import (
	"encoding/json"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

const (
	Labels = string(obs.FilterTypeOpenshiftLabels)
)

func NewLabels(labels map[string]string) (string, error) {
	if labels != nil && len(labels) != 0 {
		s, _ := json.Marshal(labels)
		return fmt.Sprintf(".openshift.labels = %s", s), nil
	}
	return "", nil
}
