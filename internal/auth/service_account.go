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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileServiceAccount reconciles the serviceAccount for a workload
func ReconcileServiceAccount(k8sClient client.Client, namespace string, resNames *factory.ForwarderResourceNames, owner metav1.OwnerReference) (err error) {
	if namespace == constants.OpenshiftNS && resNames.ServiceAccount == constants.LogfilesmetricexporterName {
		serviceAccount := runtime.NewServiceAccount(namespace, resNames.ServiceAccount)
		utils.AddOwnerRefToObject(serviceAccount, owner)
		serviceAccount.ObjectMeta.Finalizers = append(serviceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
		if serviceAccount, err = reconcile.ServiceAccount(k8sClient, serviceAccount); err != nil {
			return err
		}
		_, err = ReconcileServiceAccountTokenSecret(serviceAccount, k8sClient, namespace, resNames.ServiceAccountTokenSecret, owner)
		return err
	}
	return nil
}

func ReconcileServiceAccountTokenSecret(sa *corev1.ServiceAccount, k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) (desired *corev1.Secret, err error) {

	if sa == nil {
		return nil, fmt.Errorf("serviceAccount not provided in-order to generate a token")
	}

	desired = runtime.NewSecret(namespace, name, map[string][]byte{})
	desired.Annotations = map[string]string{
		corev1.ServiceAccountNameKey: sa.Name,
		corev1.ServiceAccountUIDKey:  string(sa.UID),
	}
	desired.Type = corev1.SecretTypeServiceAccountToken
	utils.AddOwnerRefToObject(desired, owner)
	current := &corev1.Secret{}
	if err = k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(desired), current); err == nil {
		return current, nil
	} else if !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get %s token secret: %w", name, err)
	}

	if err := k8sClient.Create(context.TODO(), desired); err != nil {
		return nil, fmt.Errorf("failed to create %s token secret: %w", name, err)
	}

	return desired, nil
}
