package configmaps

import (
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

type ComparisonOption int

const (
	CompareAnnotations ComparisonOption = 0
	CompareLabels      ComparisonOption = 1
)

//AreSame compares configmaps for equality and return true equal otherwise false.  This comparison
//only compares the data of the configmap by default unless otherwise configured
func AreSame(actual *corev1.ConfigMap, desired *corev1.ConfigMap, options ...ComparisonOption) bool {
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

	return dataAreEqual && labelsAreEqual && annotationsAreEqual
}
