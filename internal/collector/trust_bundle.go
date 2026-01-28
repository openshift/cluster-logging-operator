package collector

import (
	"context"
	"fmt"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	DefaultPollInterval = 5 * time.Second
	DefaultTimeOut      = 30 * time.Second
)

// ReconcileTrustedCABundleConfigMap creates or returns an existing Trusted CA Bundle ConfigMap.
// By setting label "config.openshift.io/inject-trusted-cabundle: true", the cert is automatically filled/updated.
func ReconcileTrustedCABundleConfigMap(k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) error {
	cm := runtime.NewConfigMap(namespace, name, nil)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8sClient, cm, func() error {
		cm.Labels = map[string]string{constants.InjectTrustedCABundleLabel: "true"}
		cm.OwnerReferences = []metav1.OwnerReference{owner}
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled TrustedCABundle ConfigMap - operation: %s", op))
	}
	return err
}

// WaitForTrustedCAToBePopulated polls for the given configmap to
func WaitForTrustedCAToBePopulated(k8sClient client.Client, namespace, name string, pollInterval, timeout time.Duration) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	err := wait.PollUntilContextTimeout(context.TODO(), pollInterval, timeout, true, func(ctx context.Context) (done bool, err error) {
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
