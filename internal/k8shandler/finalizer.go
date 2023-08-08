package k8shandler

import (
	"context"
	"github.com/ViaQ/logerr/v2/kverrors"
	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (clusterRequest *ClusterLoggingRequest) appendFinalizer(identifier string) error {
	instance, err := loader.FetchClusterLogging(clusterRequest.Client, clusterRequest.Cluster.Namespace, clusterRequest.Cluster.Name, true)
	if err != nil {
		return kverrors.Wrap(err, "Error getting ClusterLogging for appending finalizer.")
	}

	for _, f := range instance.GetFinalizers() {
		if f == identifier {
			// Skip if finalizer already exists
			return nil
		}
	}

	instance.Finalizers = append(instance.GetFinalizers(), identifier)
	if err := clusterRequest.Update(&instance); err != nil {
		return kverrors.Wrap(err, "Can not update ClusterLogging finalizers.")
	}

	return nil
}

func RemoveFinalizer(k8sClient client.Client, namespace, name, identifier string) error {
	instance, err := loader.FetchClusterLogging(k8sClient, namespace, name, true)
	if err != nil {
		return kverrors.Wrap(err, "Error getting ClusterLogging for removing finalizer.")
	}

	found := false
	finalizers := []string{}
	for _, f := range instance.GetFinalizers() {
		if f == identifier {
			found = true
			continue
		}

		finalizers = append(finalizers, f)
	}

	if !found {
		// Finalizer is not in list anymore
		return nil
	}

	instance.Finalizers = finalizers
	if err := k8sClient.Update(context.TODO(), &instance); err != nil {
		return kverrors.Wrap(err, "Failed to remove finalizer from ClusterLogging.")
	}

	return nil
}
