package auth

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileServiceAccount reconciles the serviceaccount for collector or logfilemetricexporter
func ReconcileServiceAccount(er record.EventRecorder, k8sClient client.Client, namespace string, resNames *factory.ForwarderResourceNames, owner metav1.OwnerReference) (err error) {
	if namespace == constants.OpenshiftNS && (resNames.ServiceAccount == constants.CollectorServiceAccountName || resNames.ServiceAccount == constants.LogfilesmetricexporterName) {
		serviceAccount := runtime.NewServiceAccount(namespace, resNames.ServiceAccount)
		utils.AddOwnerRefToObject(serviceAccount, owner)
		serviceAccount.ObjectMeta.Finalizers = append(serviceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
		if serviceAccount, err = reconcile.ServiceAccount(er, k8sClient, serviceAccount); err != nil {
			return err
		}
		return reconcileServiceAccountTokenSecret(serviceAccount, k8sClient, namespace, resNames.ServiceAccountTokenSecret, owner)
	}
	return nil
}

func reconcileServiceAccountTokenSecret(serviceAccount *corev1.ServiceAccount, k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) error {
	desired := runtime.NewSecret(namespace, name, map[string][]byte{})
	desired.Annotations = map[string]string{
		corev1.ServiceAccountNameKey: serviceAccount.Name,
		corev1.ServiceAccountUIDKey:  string(serviceAccount.UID),
	}
	desired.Type = corev1.SecretTypeServiceAccountToken
	utils.AddOwnerRefToObject(desired, owner)
	current := &corev1.Secret{}
	if err := k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(desired), current); err == nil {
		accountName := desired.Annotations[corev1.ServiceAccountNameKey]
		accountUID := desired.Annotations[corev1.ServiceAccountUIDKey]
		if (accountName != serviceAccount.Name || accountUID != string(serviceAccount.UID)) &&
			!utils.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			// Delete secret, so that we can create a new one next loop
			if err := k8sClient.Delete(context.TODO(), current); err != nil {
				return nil
			}
			return fmt.Errorf("deleted stale secret: %s", name)
		}
		// Existing secret is up-to-date
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get %s token secret: %w", name, err)
	}

	if err := k8sClient.Create(context.TODO(), desired); err != nil {
		return fmt.Errorf("failed to create %s token secret: %w", name, err)
	}

	return nil
}
