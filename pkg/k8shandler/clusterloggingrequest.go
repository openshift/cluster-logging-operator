package k8shandler

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLoggingRequest struct {
	client  client.Client
	cluster *logging.ClusterLogging
}

// TODO: determine if this is even necessary
func (clusterRequest *ClusterLoggingRequest) isManaged() bool {
	return clusterRequest.cluster.Spec.ManagementState == logging.ManagementStateManaged
}

func (clusterRequest *ClusterLoggingRequest) Create(object runtime.Object) error {
	return clusterRequest.client.Create(context.TODO(), object)
}

func (clusterRequest *ClusterLoggingRequest) Update(object runtime.Object) error {
	return clusterRequest.client.Update(context.TODO(), object)
}

func (clusterRequest *ClusterLoggingRequest) Get(objectName string, object runtime.Object) error {
	namespacedName := types.NamespacedName{Name: objectName, Namespace: clusterRequest.cluster.Namespace}

	return clusterRequest.client.Get(context.TODO(), namespacedName, object)
}

func (clusterRequest *ClusterLoggingRequest) List(selector map[string]string, object runtime.Object) error {

	labelSelector := labels.SelectorFromSet(selector)

	return clusterRequest.client.List(
		context.TODO(),
		&client.ListOptions{Namespace: clusterRequest.cluster.Namespace, LabelSelector: labelSelector},
		object,
	)
}

func (clusterRequest *ClusterLoggingRequest) Delete(object runtime.Object) error {
	return clusterRequest.client.Delete(context.TODO(), object)
}
