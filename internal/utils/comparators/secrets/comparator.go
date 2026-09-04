package secrets

import (
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	corev1 "k8s.io/api/core/v1"
)

// AreSame compares secrets for equality and return true equal otherwise false.  This comparison
// only compares the data of the secrets by default unless otherwise configured
func AreSame(actual *corev1.Secret, desired *corev1.Secret, options ...comparators.ComparisonOption) bool {
	log.V(5).Info("Compare secret", "actual", actual)
	log.V(5).Info("Compare secret", "desired", desired)
	log.V(5).Info("Compare secret", "options", options)
	dataAreEqual := reflect.DeepEqual(actual.Data, desired.Data)
	if !dataAreEqual {
		log.V(3).Info("Compare secrets", "dateAreEqual", dataAreEqual)
		return false
	}
	ownerEqual := utils.HasSameOwner(actual.OwnerReferences, desired.OwnerReferences)
	if !ownerEqual {
		log.V(3).Info("Compare secrets", "ownerEqual", ownerEqual)
		return false
	}
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
	log.V(3).Info("Compare secrets", "dateAreEqual", dataAreEqual, "labelsAreEqual", labelsAreEqual, "annotationsAreEqual", annotationsAreEqual)
	return dataAreEqual && labelsAreEqual && annotationsAreEqual
}
