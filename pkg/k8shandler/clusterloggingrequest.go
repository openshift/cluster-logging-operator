package k8shandler

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLoggingRequest struct {
	Client  client.Client
	Cluster *logging.ClusterLogging

	// ForwarderRequest is a logforwarder instance
	ForwarderRequest *logging.ClusterLogForwarder

	// ForwarderSpec is the normalized and sanitized logforwarder spec
	ForwarderSpec logging.ClusterLogForwarderSpec
}

// TODO: determine if this is even necessary
func (clusterRequest *ClusterLoggingRequest) isManaged() bool {
	return clusterRequest.Cluster.Spec.ManagementState == logging.ManagementStateManaged
}

func (clusterRequest *ClusterLoggingRequest) Create(object runtime.Object) error {
	logrus.Tracef("Creating: %v", object)
	err := clusterRequest.Client.Create(context.TODO(), object)
	logrus.Tracef("Response: %v", err)
	return err
}

//Update the runtime Object or return error
func (clusterRequest *ClusterLoggingRequest) Update(object runtime.Object) (err error) {
	logrus.Tracef("Updating: %v", object)
	if err = clusterRequest.Client.Update(context.TODO(), object); err != nil {
		logrus.Errorf("Error updating %v: %v", object.GetObjectKind(), err)
	}
	return err
}

//Update the runtime Object status or return error
func (clusterRequest *ClusterLoggingRequest) UpdateStatus(object runtime.Object) (err error) {
	logrus.Tracef("Updating Status: %v", object)
	if err = clusterRequest.Client.Status().Update(context.TODO(), object); err != nil {
		// making this debug because we should be throwing the returned error if we are never
		// able to update the status
		logrus.Debugf("Error updating status: %v", err)
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) Get(objectName string, object runtime.Object) error {
	namespacedName := types.NamespacedName{Name: objectName, Namespace: clusterRequest.Cluster.Namespace}

	logrus.Debugf("Getting namespacedName: %v, object: %v", namespacedName, object)

	return clusterRequest.Client.Get(context.TODO(), namespacedName, object)
}

func (clusterRequest *ClusterLoggingRequest) GetClusterResource(objectName string, object runtime.Object) error {
	namespacedName := types.NamespacedName{Name: objectName}
	logrus.Debugf("Getting ClusterResource namespacedName: %v, object: %v", namespacedName, object)
	err := clusterRequest.Client.Get(context.TODO(), namespacedName, object)
	logrus.Debugf("Response: %v", err)
	return err
}

func (clusterRequest *ClusterLoggingRequest) List(selector map[string]string, object runtime.Object) error {
	logrus.Debugf("Listing selector: %v, object: %v", selector, object)

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

func (clusterRequest *ClusterLoggingRequest) Delete(object runtime.Object) error {
	logrus.Debugf("Deleting: %v", object)
	return clusterRequest.Client.Delete(context.TODO(), object)
}

func (clusterRequest *ClusterLoggingRequest) UpdateCondition(t logging.ConditionType, message string, reason logging.ConditionReason, status v1.ConditionStatus) error {
	if logging.SetCondition(&clusterRequest.Cluster.Status.Conditions, t, status, reason, message) {
		return clusterRequest.UpdateStatus(clusterRequest.Cluster)
	}
	return nil
}
