package collector

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileCollectorConfig reconciles a collector config specifically for the collector defined by the factory
func (f *Factory) ReconcileCollectorConfig(k8sClient client.Client, reader client.Reader, namespace, collectorConfig string, owner metav1.OwnerReference) error {
	log.V(3).Info("Updating ConfigMap and Secrets")
	configMap := runtime.NewConfigMap(
		namespace,
		f.ResourceNames.ConfigMap,
		map[string]string{
			vector.ConfigFile:    collectorConfig,
			vector.RunVectorFile: fmt.Sprintf(vector.RunVectorScript, vector.GetDataPath(namespace, f.ResourceNames.ForwarderName)),
		},
		f.CommonLabelInitializer)

	utils.AddOwnerRefToObject(configMap, owner)
	if err := reconcile.Configmap(k8sClient, reader, configMap, comparators.CompareLabels); err != nil {
		return err
	}
	return RemoveLegacyCollectorConfigMap(k8sClient, namespace, f.ResourceNames.ForwarderName, owner)
}

// RemoveLegacyCollectorConfigMap deletes the pre-LOG-9591 ConfigMap named "{forwarderName}-config"
// when it is owned by this ClusterLogForwarder. Ownership is required so a LokiStack (or other)
// ConfigMap that shares the legacy name is not removed.
func RemoveLegacyCollectorConfigMap(k8sClient client.Client, namespace, forwarderName string, owner metav1.OwnerReference) error {
	legacyName := factory.LegacyCollectorConfigMapName(forwarderName)
	current := &corev1.ConfigMap{}
	key := types.NamespacedName{Namespace: namespace, Name: legacyName}
	if err := k8sClient.Get(context.TODO(), key, current); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get legacy collector configmap %s/%s: %w", namespace, legacyName, err)
	}
	if !utils.HasSameOwner(current.OwnerReferences, []metav1.OwnerReference{owner}) {
		log.V(3).Info("Skipping delete of legacy collector ConfigMap not owned by this forwarder",
			"namespace", namespace, "name", legacyName)
		return nil
	}
	if err := k8sClient.Delete(context.TODO(), current); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failure deleting legacy collector configmap %s/%s: %w", namespace, legacyName, err)
	}
	log.V(3).Info("Deleted legacy collector ConfigMap", "namespace", namespace, "name", legacyName)
	return nil
}
