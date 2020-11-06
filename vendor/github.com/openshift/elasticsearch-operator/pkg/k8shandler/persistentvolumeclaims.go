package k8shandler

import (
	"context"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createOrUpdatePersistentVolumeClaim(pvc v1.PersistentVolumeClaimSpec, newName, namespace, clusterName string, client client.Client) error {

	// for some reason if the PVC already exists but creating it again would violate
	// quota we get an error regarding quota not that it already exists
	// so check to see if it already exists
	claim := createPersistentVolumeClaim(newName, namespace, clusterName, pvc)

	current := &v1.PersistentVolumeClaim{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: claim.Name, Namespace: claim.Namespace}, current)
	if err == nil {
		return updatePersistentVolumeClaim(claim, client)
	}

	if !errors.IsNotFound(err) {
		logrus.Errorf("Could not get PVC %q: %v", newName, err)
		return err
	}

	err = client.Create(context.TODO(), claim)
	if err == nil {
		return nil
	}

	if !errors.IsAlreadyExists(err) {
		return fmt.Errorf("unable to create PVC: %w", err)
	}

	return updatePersistentVolumeClaim(claim, client)
}

func updatePersistentVolumeClaim(claim *v1.PersistentVolumeClaim, client client.Client) error {

	current := &v1.PersistentVolumeClaim{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := client.Get(context.TODO(), types.NamespacedName{Name: claim.Name, Namespace: claim.Namespace}, current); err != nil {
			if errors.IsNotFound(err) {
				// the object doesn't exist -- it was likely culled
				// recreate it on the next time through if necessary
				return nil
			}
			return fmt.Errorf("Failed to get %v PVC: %v", claim.Name, err)
		}

		if !reflect.DeepEqual(current.ObjectMeta.Labels, claim.ObjectMeta.Labels) {
			current.ObjectMeta.Labels = claim.ObjectMeta.Labels

			if err := client.Update(context.TODO(), current); err != nil {
				return err
			}
		}

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	return nil
}

func createPersistentVolumeClaim(pvcName, namespace, clusterName string, volSpec v1.PersistentVolumeClaimSpec) *v1.PersistentVolumeClaim {
	pvc := persistentVolumeClaim(pvcName, namespace, clusterName)
	pvc.Spec = volSpec
	return pvc
}

func persistentVolumeClaim(pvcName, namespace, clusterName string) *v1.PersistentVolumeClaim {

	pvcLabels := map[string]string{
		"logging-cluster": clusterName,
	}

	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
			Labels:    pvcLabels,
		},
	}
}
