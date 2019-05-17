package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	scheduling "k8s.io/api/scheduling/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewPriorityClass is a constructor to create a PriorityClass
func NewPriorityClass(priorityclassName string, priorityValue int32, globalDefault bool, description string) *scheduling.PriorityClass {
	return &scheduling.PriorityClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PriorityClass",
			APIVersion: scheduling.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: priorityclassName,
		},
		Value:         priorityValue,
		GlobalDefault: globalDefault,
		Description:   description,
	}
}

//RemovePriorityClass removes a priority class of a given name
func (clusterRequest *ClusterLoggingRequest) RemovePriorityClass(priorityclassName string) error {
	collectionPriorityClass := NewPriorityClass(
		priorityclassName,
		1000000,
		false,
		"This priority class is for the Cluster-Logging Collector",
	)

	err := clusterRequest.Delete(collectionPriorityClass)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v priority class %v", priorityclassName, err)
	}

	return nil
}
