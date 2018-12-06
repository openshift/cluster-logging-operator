package k8shandler

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
)

func createOrUpdatePersistentVolumeClaim(pvc v1.PersistentVolumeClaimSpec, newName string, namespace string) error {
	claim := persistentVolumeClaim(newName, namespace)
	err := sdk.Get(claim)
	if err != nil {
		// PVC doesn't exists, needs to be created.
		claim = createPersistentVolumeClaim(newName, namespace, pvc)
		logrus.Infof("Creating new PVC: %v", newName)
		err = sdk.Create(claim)
		if err != nil {
			return fmt.Errorf("Unable to create PVC: %v", err)
		}
	} else {
		logrus.Infof("Reusing existing PVC: %s", newName)
		// TODO for updates, don't forget to use retry.RetryOnConflict
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
