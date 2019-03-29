package k8shandler

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createOrUpdatePersistentVolumeClaim(pvc v1.PersistentVolumeClaimSpec, newName string, namespace string) error {

	claim := createPersistentVolumeClaim(newName, namespace, pvc)
	err := sdk.Create(claim)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Unable to create PVC: %v", err)
		}
	}

	return nil
}

func createPersistentVolumeClaim(pvcName, namespace string, volSpec v1.PersistentVolumeClaimSpec) *v1.PersistentVolumeClaim {
	pvc := persistentVolumeClaim(pvcName, namespace)
	pvc.Spec = volSpec
	return pvc
}

func persistentVolumeClaim(pvcName, namespace string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
		},
	}
}
