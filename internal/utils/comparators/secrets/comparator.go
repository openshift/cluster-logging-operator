package secrets

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

type ComparisonOption int

const (
	CompareAnnotations ComparisonOption = 0
	CompareLabels      ComparisonOption = 1
)

// AreSame compares secrets for equality and return true equal otherwise false.  This comparison
// only compares the data of the secrets by default unless otherwise configured
func AreSame(actual *corev1.Secret, desired *corev1.Secret, options ...ComparisonOption) bool {
	log.V(5).Info("Compare secret", "actual", actual)
	log.V(5).Info("Compare secret", "desired", desired)
	log.V(5).Info("Compare secret", "options", options)
	dataAreEqual := reflect.DeepEqual(actual.Data, desired.Data)
	if !dataAreEqual {
		log.V(3).Info("Compare secrets", "dateAreEqual", dataAreEqual)
		return false
	}
	labelsAreEqual := true
	annotationsAreEqual := true
	for _, opt := range options {
		switch opt {
		case CompareAnnotations:
			annotationsAreEqual = reflect.DeepEqual(actual.Annotations, desired.Annotations)
		case CompareLabels:
			labelsAreEqual = reflect.DeepEqual(actual.Labels, desired.Labels)
		}
	}
	log.V(3).Info("Compare secrets", "dateAreEqual", dataAreEqual, "labelsAreEqual", labelsAreEqual, "annotationsAreEqual", annotationsAreEqual)
	return dataAreEqual && labelsAreEqual && annotationsAreEqual
}
