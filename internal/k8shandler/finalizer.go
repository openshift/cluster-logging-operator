package k8shandler

import (
	"github.com/ViaQ/logerr/v2/kverrors"
)

func (clusterRequest *ClusterLoggingRequest) appendFinalizer(identifier string) error {
	instance := clusterRequest.Cluster.DeepCopy()
	for _, f := range instance.GetFinalizers() {
		if f == identifier {
			// Skip if finalizer already exists
			return nil
		}
	}

	instance.Finalizers = append(instance.GetFinalizers(), identifier)
	if err := clusterRequest.Update(instance); err != nil {
		return kverrors.Wrap(err, "Can not update ClusterLogging finalizers.")
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeFinalizer(identifier string) error {
	instance := clusterRequest.Cluster.DeepCopy()

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
	if err := clusterRequest.Update(instance); err != nil {
		return kverrors.Wrap(err, "Failed to remove finalizer from ClusterLogging.")
	}

	return nil
}
