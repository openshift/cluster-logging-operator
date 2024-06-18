package observability

import (
	"context"
	log "github.com/ViaQ/logerr/v2/log/static"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchClusterLogForwarder returns a copy of the ClusterLogForwarder
func FetchClusterLogForwarder(k8sClient client.Client, namespace, name string) (*obs.ClusterLogForwarder, error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := obsruntime.NewClusterLogForwarder(namespace, name, runtime.Initialize)
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		return proto, err
	}
	// Do not modify cached copy
	return proto.DeepCopy(), nil
}

// FetchSecrets from a list of names in a given namespace
func FetchSecrets(k8sClient client.Client, namespace string, names ...string) (secrets []*corev1.Secret, err error) {
	log := log.WithName("#FetchSecrets")
	for _, name := range names {
		key := types.NamespacedName{Name: name, Namespace: namespace}
		proto := runtime.NewSecret(namespace, name, nil)
		log.V(4).Info("loading secret", "key", key)
		if err = k8sClient.Get(context.TODO(), key, proto); err == nil {
			log.V(4).Info("found secret", "key", key)
			secrets = append(secrets, proto)
		} else if errors.IsNotFound(err) {
			log.V(1).Info("secret not found", "key", key)
		} else {
			log.V(0).Error(err, "unable to fetch secret", "key", key)
			return nil, err
		}
	}
	return secrets, nil
}

// FetchConfigMaps from a list of names in a given namespace
func FetchConfigMaps(k8sClient client.Client, namespace string, names ...string) (configMaps []*corev1.ConfigMap, err error) {
	log := log.WithName("#FetchConfigMaps")
	for _, name := range names {
		key := types.NamespacedName{Name: name, Namespace: namespace}
		log.V(4).Info("loading configmap", "key", key)
		proto := runtime.NewConfigMap(namespace, name, nil)
		if err = k8sClient.Get(context.TODO(), key, proto); err == nil {
			log.V(4).Info("found configmap", "key", key)
			configMaps = append(configMaps, proto)
		} else if errors.IsNotFound(err) {
			log.V(1).Info("configmap not found", "key", key)
		} else {
			log.V(0).Error(err, "unable to fetch configmap", "key", key)
			return nil, err
		}
	}
	return configMaps, nil
}
