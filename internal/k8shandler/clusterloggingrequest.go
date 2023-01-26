package k8shandler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLoggingRequest struct {
	Client client.Client
	Reader client.Reader

	//ClusterID is the unique identifier of the cluster in which the operator is deployed
	ClusterID string

	Cluster       *logging.ClusterLogging
	EventRecorder record.EventRecorder
	// ForwarderRequest is a logforwarder instance
	ForwarderRequest *logging.ClusterLogForwarder

	// ForwarderSpec is the normalized and sanitized logforwarder spec
	ForwarderSpec logging.ClusterLogForwarderSpec

	// OutputSecrets are retrieved during validation and used for generation.
	OutputSecrets map[string]*corev1.Secret
}

type ClusterLogForwarderVerifier struct {
	VerifyOutputSecret func(output *logging.OutputSpec, conds logging.NamedConditions) bool
}

func (clusterRequest *ClusterLoggingRequest) IncludesManagedStorage() bool {
	return clusterRequest.Cluster != nil && clusterRequest.Cluster.Spec.LogStore != nil
}

// true if equals "Managed" or empty
func (clusterRequest *ClusterLoggingRequest) isManaged() bool {
	return clusterRequest.Cluster.Spec.ManagementState == logging.ManagementStateManaged ||
		clusterRequest.Cluster.Spec.ManagementState == ""
}

func (clusterRequest *ClusterLoggingRequest) Create(object client.Object) error {
	err := clusterRequest.Client.Create(context.TODO(), object)
	return err
}

// Update the runtime Object or return error
func (clusterRequest *ClusterLoggingRequest) Update(object client.Object) (err error) {
	if err = clusterRequest.Client.Update(context.TODO(), object); err != nil {
		log.Error(err, "Error updating ", object.GetObjectKind())
	}
	return err
}

// UpdateStatus modifies the status sub-resource or returns an error.
func (clusterRequest *ClusterLoggingRequest) UpdateStatus(object client.Object) (err error) {
	if err = clusterRequest.Client.Status().Update(context.TODO(), object); err != nil {
		// making this debug because we should be throwing the returned error if we are never
		// able to update the status
		log.V(2).Error(err, "Error updating status")
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) Get(objectName string, object client.Object) error {
	namespacedName := types.NamespacedName{Name: objectName, Namespace: clusterRequest.Cluster.Namespace}

	log.V(3).Info("Getting object", "namespacedName", namespacedName, "object", object)

	return clusterRequest.Client.Get(context.TODO(), namespacedName, object)
}

func (clusterRequest *ClusterLoggingRequest) GetClusterResource(objectName string, object client.Object) error {
	namespacedName := types.NamespacedName{Name: objectName}
	log.V(3).Info("Getting ClusterResource object", "namespacedName", namespacedName, "object", object)
	err := clusterRequest.Client.Get(context.TODO(), namespacedName, object)
	log.V(3).Error(err, "Response")
	return err
}

func (clusterRequest *ClusterLoggingRequest) List(selector map[string]string, object client.ObjectList) error {
	log.V(3).Info("Listing selector object", "selector", selector, "object", object)

	listOpts := []client.ListOption{
		client.InNamespace(clusterRequest.Cluster.Namespace),
		client.MatchingLabels(selector),
	}

	return clusterRequest.Client.List(
		context.TODO(),
		object,
		listOpts...,
	)
}

func (clusterRequest *ClusterLoggingRequest) Delete(object client.Object) error {
	log.V(3).Info("Deleting", "object", object)
	return clusterRequest.Client.Delete(context.TODO(), object)
}

func (clusterRequest *ClusterLoggingRequest) UpdateCondition(t logging.ConditionType, message string, reason logging.ConditionReason, status v1.ConditionStatus) error {
	instance, err := clusterRequest.getClusterLogging(true)
	if err != nil {
		return err
	}

	if logging.SetCondition(&instance.Status.Conditions, t, status, reason, message) {
		return clusterRequest.UpdateStatus(instance)
	}
	return nil
}
