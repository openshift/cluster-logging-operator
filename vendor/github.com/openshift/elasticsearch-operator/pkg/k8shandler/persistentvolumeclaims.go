package k8shandler

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createOrUpdatePersistentVolumeClaim(pvc v1.PersistentVolumeClaimSpec, newName, namespace string, client client.Client) error {

	// for some reason if the PVC already exists but creating it again would violate
	// quota we get an error regarding quota not that it already exists
	// so check to see if it already exists
	claim := &v1.PersistentVolumeClaim{}

	if getErr := client.Get(context.TODO(), types.NamespacedName{Name: newName, Namespace: namespace}, claim); getErr != nil {
		if errors.IsNotFound(getErr) {
			claim = createPersistentVolumeClaim(newName, namespace, pvc)
			err := client.Create(context.TODO(), claim)
			if err != nil {
				if !errors.IsAlreadyExists(err) {
					return fmt.Errorf("Unable to create PVC: %v", err)
				}
			}
		} else {
			logrus.Debugf("Could not get PVC %v: %v", newName, getErr)
			return getErr
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
