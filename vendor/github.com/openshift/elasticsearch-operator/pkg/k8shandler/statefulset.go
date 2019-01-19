package k8shandler

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type statefulSetNode struct {
	resource apps.StatefulSet
}

func (node *statefulSetNode) getResource() runtime.Object {
	return &node.resource
}

func (node *statefulSetNode) getRevision(cfg *desiredNodeState) (string, error) {
	return "", nil
}

func (node *statefulSetNode) awaitingRollout(cfg *desiredNodeState, currentRevision string) (bool, error) {
	return false, nil
}

func (node *statefulSetNode) needsPause(cfg *desiredNodeState) (bool, error) {
	return false, nil
}

func (node *statefulSetNode) isDifferent(cfg *desiredNodeState) (bool, error) {
	// Check replicas number
	if cfg.getReplicas() != *node.resource.Spec.Replicas {
		return true, nil
	}

	// Check if the Variables are the desired ones

	return false, nil
}

// isUpdateNeeded returns true if update is needed
func (node *statefulSetNode) isUpdateNeeded(cfg *desiredNodeState) bool {
	// This operator doesn't update nodes managed by StatefulSets in rolling fashion
	return false
}

func (node *statefulSetNode) query() error {
	err := sdk.Get(&node.resource)
	return err
}

// constructNodeStatefulSet creates the StatefulSet for the node
func (node *statefulSetNode) constructNodeResource(cfg *desiredNodeState, owner metav1.OwnerReference) (runtime.Object, error) {

	replicas := cfg.getReplicas()

	statefulSet := node.resource
	//statefulSet(cfg.DeployName, node.resource.ObjectMeta.Namespace)
	statefulSet.ObjectMeta.Labels = cfg.getLabels()

	statefulSet.Spec = apps.StatefulSetSpec{
		Replicas:    &replicas,
		ServiceName: cfg.DeployName,
		Selector: &metav1.LabelSelector{
			MatchLabels: cfg.getLabels(),
		},
		Template: cfg.constructPodTemplateSpec(),
	}

	pvc, ok, err := cfg.generateMasterPVC()
	if err != nil {
		return &statefulSet, err
	}
	if ok {
		statefulSet.Spec.VolumeClaimTemplates = []v1.PersistentVolumeClaim{
			pvc,
		}
	}

	addOwnerRefToObject(&statefulSet, owner)

	return &statefulSet, nil
}

func (node *statefulSetNode) delete() error {
	err := sdk.Delete(&node.resource)
	if err != nil {
		return fmt.Errorf("Unable to delete StatefulSet %v: ", err)
	}
	return nil
}
