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
func FetchSecrets(k8sClient client.Client, namespace string, names ...string) ([]*corev1.Secret, error) {
	secrets := []*corev1.Secret{}
	for _, name := range names {
		key := types.NamespacedName{Name: name, Namespace: namespace}
		proto := runtime.NewSecret(namespace, name, nil)
		if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
			secrets = append(secrets, proto)
		} else if errors.IsNotFound(err) {
			log.V(1).Info("Secret not found", "namespace", namespace, "name", name)
		} else {
			return nil, err
		}
	}
	return secrets, nil
}

// FetchConfigMaps from a list of names in a given namespace
func FetchConfigMaps(k8sClient client.Client, namespace string, names ...string) ([]*corev1.ConfigMap, error) {
	configMaps := []*corev1.ConfigMap{}
	for _, name := range names {
		key := types.NamespacedName{Name: name, Namespace: namespace}
		proto := runtime.NewConfigMap(namespace, name, nil)
		if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
			configMaps = append(configMaps, proto)
		} else if errors.IsNotFound(err) {
			log.V(1).Info("Secret not found", "namespace", namespace, "name", name)
		} else {
			return nil, err
		}
	}
	return configMaps, nil
}
