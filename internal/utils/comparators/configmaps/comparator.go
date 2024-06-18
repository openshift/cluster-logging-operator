package configmaps

import (
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	corev1 "k8s.io/api/core/v1"
)

// AreSame compares configmaps for equality and return true equal otherwise false.  This comparison
// only compares the data of the configmap by default unless otherwise configured
func AreSame(actual *corev1.ConfigMap, desired *corev1.ConfigMap, options ...comparators.ComparisonOption) bool {
	log.V(5).Info("Compare configmaps", "actual", actual)
	log.V(5).Info("Compare configmaps", "desired", desired)
	log.V(5).Info("Compare configmaps", "options", options)
	dataAreEqual := reflect.DeepEqual(actual.Data, desired.Data)
	labelsAreEqual := true
	annotationsAreEqual := true
	for _, opt := range options {
		switch opt {
		case comparators.CompareAnnotations:
			annotationsAreEqual = reflect.DeepEqual(actual.Annotations, desired.Annotations)
		case comparators.CompareLabels:
			labelsAreEqual = reflect.DeepEqual(actual.Labels, desired.Labels)
		}
	}
	log.V(5).Info("Compare configmaps", "dateAreEqual", dataAreEqual, "labelsAreEqual", labelsAreEqual, "annotationsAreEqual", annotationsAreEqual)
	return dataAreEqual && labelsAreEqual && annotationsAreEqual
}
