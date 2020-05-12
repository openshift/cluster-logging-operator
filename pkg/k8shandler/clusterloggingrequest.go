package k8shandler

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	collector "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLoggingRequest struct {
	client  client.Client
	cluster *logging.ClusterLogging

	//forwardingRequest is a logforwarding instance
	ForwardingRequest *logforward.LogForwarding

	//ForwardingSpec is the normalized and sanitized logforwarding spec
	ForwardingSpec logforward.ForwardingSpec

	Collector *collector.CollectorSpec
}

// TODO: determine if this is even necessary
func (clusterRequest *ClusterLoggingRequest) isManaged() bool {
	return clusterRequest.cluster.Spec.ManagementState == logging.ManagementStateManaged
}

func (clusterRequest *ClusterLoggingRequest) Create(object runtime.Object) error {
	logrus.Tracef("Creating: %v", object)
	err := clusterRequest.client.Create(context.TODO(), object)
	logrus.Tracef("Response: %v", err)
	return err
}

//Update the runtime Object or return error
func (clusterRequest *ClusterLoggingRequest) Update(object runtime.Object) (err error) {
	logrus.Tracef("Updating: %v", object)
	if err = clusterRequest.client.Update(context.TODO(), object); err != nil {
		logrus.Errorf("Error updating %v: %v", object.GetObjectKind(), err)
	}
	return err
}

//Update the runtime Object status or return error
func (clusterRequest *ClusterLoggingRequest) UpdateStatus(object runtime.Object) (err error) {
	logrus.Tracef("Updating Status: %v", object)
	if err = clusterRequest.client.Status().Update(context.TODO(), object); err != nil {
		// making this debug because we should be throwing the returned error if we are never
		// able to update the status
		logrus.Debugf("Error updating status: %v", err)
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) Get(objectName string, object runtime.Object) error {
	namespacedName := types.NamespacedName{Name: objectName, Namespace: clusterRequest.cluster.Namespace}

	logrus.Debugf("Getting namespacedName: %v, object: %v", namespacedName, object)

	return clusterRequest.client.Get(context.TODO(), namespacedName, object)
}

func (clusterRequest *ClusterLoggingRequest) GetClusterResource(objectName string, object runtime.Object) error {
	namespacedName := types.NamespacedName{Name: objectName}
	logrus.Debugf("Getting ClusterResource namespacedName: %v, object: %v", namespacedName, object)
	err := clusterRequest.client.Get(context.TODO(), namespacedName, object)
	logrus.Debugf("Response: %v", err)
	return err
}

func (clusterRequest *ClusterLoggingRequest) List(selector map[string]string, object runtime.Object) error {
	logrus.Debugf("Listing selector: %v, object: %v", selector, object)

	labelSelector := labels.SelectorFromSet(selector)

	return clusterRequest.client.List(
		context.TODO(),
		&client.ListOptions{Namespace: clusterRequest.cluster.Namespace, LabelSelector: labelSelector},
		object,
	)
}

func (clusterRequest *ClusterLoggingRequest) Delete(object runtime.Object) error {
	logrus.Debugf("Deleting: %v", object)
	return clusterRequest.client.Delete(context.TODO(), object)
}
