package collector

import (
	"context"
	"fmt"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileTrustedCABundleConfigMap creates or returns an existing Trusted CA Bundle ConfigMap.
// By setting label "config.openshift.io/inject-trusted-cabundle: true", the cert is automatically filled/updated.
func ReconcileTrustedCABundleConfigMap(er record.EventRecorder, k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) error {
	desired := runtime.NewConfigMap(
		namespace,
		name,
		map[string]string{
			constants.TrustedCABundleKey: "",
		},
	)
	desired.ObjectMeta.Labels = make(map[string]string)
	desired.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel] = "true"

	utils.AddOwnerRefToObject(desired, owner)

	reason := constants.EventReasonGetObject
	var current *corev1.ConfigMap
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current = &corev1.ConfigMap{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8sClient.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				reason = constants.EventReasonCreateObject
				if err = k8sClient.Create(context.TODO(), desired); err != nil {
					return err
				}
				return fmt.Errorf("waiting for %v ConfigMap to get created", key)
			}
			return fmt.Errorf("failed to get %v ConfigMap: %w", key, err)
		}
		if val := current.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel]; val == "true" && utils.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			return nil
		}
		if current.ObjectMeta.Labels == nil {
			current.ObjectMeta.Labels = map[string]string{}
		}
		current.ObjectMeta.Labels[constants.InjectTrustedCABundleLabel] = "true"
		current.OwnerReferences = desired.OwnerReferences
		reason = constants.EventReasonUpdateObject
		if err := k8sClient.Update(context.TODO(), current); err != nil {
			return err
		}
		return fmt.Errorf("waiting for %v ConfigMap to get created", key)
	})
	eventType := corev1.EventTypeNormal
	msg := fmt.Sprintf("%s ConfigMap %s/%s", reason, desired.Namespace, desired.Name)
	if retryErr != nil {
		eventType = corev1.EventTypeWarning
		msg = fmt.Sprintf("Unable to %s: %v", msg, retryErr)
	}
	er.Event(desired, eventType, reason, msg)
	if retryErr != nil {
		log.Error(retryErr, "collector.ReconcileTrustedCABundleConfigMap")
		return retryErr
	}

	return nil
}

// WaitForTrustedCAToBePopulated polls for the given configmap to
func WaitForTrustedCAToBePopulated(k8sClient client.Client, namespace, name string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 30*time.Second, true, func(ctx context.Context) (done bool, err error) {
		key := client.ObjectKey{Namespace: namespace, Name: name}
		if err := k8sClient.Get(context.TODO(), key, cm); err != nil {
			log.Error(err, "Error retrieving the Trusted CA Bundle")
			return false, nil
		}
		if caBundle, ok := cm.Data[constants.TrustedCABundleKey]; ok && caBundle != "" {
			return true, nil
		}
		log.V(4).Info("Configmap does not include the injected CA", "configmap", cm)
		return false, nil
	})
	if err != nil {
		log.V(4).Error(err, "Error polling for a populated Trusted CA bundle")
	}
	return cm
}
