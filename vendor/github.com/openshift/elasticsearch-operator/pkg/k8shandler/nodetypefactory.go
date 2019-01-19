package k8shandler

import (
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NodeTypeInterface interace represents individual Elasticsearch node
type NodeTypeInterface interface {
	getResource() runtime.Object
	isDifferent(cfg *desiredNodeState) (bool, error)
	isUpdateNeeded(cfg *desiredNodeState) bool
	needsPause(cfg *desiredNodeState) (bool, error)
	awaitingRollout(cfg *desiredNodeState, currentRevision string) (bool, error)
	getRevision(cfg *desiredNodeState) (string, error)
	constructNodeResource(cfg *desiredNodeState, owner metav1.OwnerReference) (runtime.Object, error)
	delete() error
	query() error
}

// NodeTypeFactory is a factory to construct either statefulset or deployment
type NodeTypeFactory func(name, namespace string) NodeTypeInterface

// NewDeploymentNode constructs deploymentNode struct for data nodes
func NewDeploymentNode(name, namespace string) NodeTypeInterface {
	depl := apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	node := deploymentNode{resource: depl}
	return &node
}

// NewStatefulSetNode constructs statefulSetNode struct for non-data nodes
func NewStatefulSetNode(name, namespace string) NodeTypeInterface {
	depl := apps.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	ss := statefulSetNode{resource: depl}
	return &ss
}
