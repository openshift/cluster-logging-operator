package openshift

import (
	"encoding/json"
	"fmt"
)

const (
	Labels = "openshiftLabels"
)

func NewLabels(labels map[string]string) (string, error) {
	if len(labels) != 0 {
		s, _ := json.Marshal(labels)
		return fmt.Sprintf(".openshift.labels = %s", s), nil
	}
	return "", nil
}
