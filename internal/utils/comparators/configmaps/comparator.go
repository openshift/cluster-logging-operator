package configmaps

import (
	"reflect"

	"github.com/ViaQ/logerr/v2/log"
	corev1 "k8s.io/api/core/v1"
)

type ComparisonOption int

const (
	CompareAnnotations ComparisonOption = 0
	CompareLabels      ComparisonOption = 1
)

//AreSame compares configmaps for equality and return true equal otherwise false.  This comparison
//only compares the data of the configmap by default unless otherwise configured
func AreSame(actual *corev1.ConfigMap, desired *corev1.ConfigMap, options ...ComparisonOption) bool {
	logger := log.NewLogger("")
	logger.V(5).Info("Compare configmaps", "actual", actual)
	logger.V(5).Info("Compare configmaps", "desired", desired)
	logger.V(5).Info("Compare configmaps", "options", options)
	dataAreEqual := reflect.DeepEqual(actual.Data, desired.Data)
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
	logger.V(3).Info("Compare configmaps", "dateAreEqual", dataAreEqual, "labelsAreEqual", labelsAreEqual, "annotationsAreEqual", annotationsAreEqual)
	return dataAreEqual && labelsAreEqual && annotationsAreEqual
}
