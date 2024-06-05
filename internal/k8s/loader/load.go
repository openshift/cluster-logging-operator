package loader

import (
	"context"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchClusterLogging
func FetchClusterLogging(k8sClient client.Client, namespace, name string) (*logging.ClusterLogging, error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := runtime.NewClusterLogging(namespace, name)
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}
		return &logging.ClusterLogging{}, err
	}
	// Do not modify cached copy
	return proto.DeepCopy(), nil
}

// FetchClusterLogForwarder
func FetchClusterLogForwarder(k8sClient client.Client, namespace, name string) (*logging.ClusterLogForwarder, error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := runtime.NewClusterLogForwarder(namespace, name)
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}
		return proto, err
	}

	// Do not modify cached copy
	return proto.DeepCopy(), nil
}
