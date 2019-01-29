package k8shandler

import (
	"fmt"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/sirupsen/logrus"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deploymentNode struct {
	resource apps.Deployment
}

func (node *deploymentNode) getResource() runtime.Object {
	return &node.resource
}

func (node *deploymentNode) getRevision(cfg *desiredNodeState) (string, error) {
	val, ok := node.resource.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]

	if ok {
		return val, nil
	}

	return "", fmt.Errorf("Unable to find revision annotation value for %v", cfg.DeployName)
}

func (node *deploymentNode) awaitingRollout(cfg *desiredNodeState, currentRevision string) (bool, error) {

	actualRevision, err := node.getRevision(cfg)
	if err != nil {
		return true, err
	}

	return actualRevision == currentRevision, nil
}

// needsPause determines whether a Deployment needs to be paused
// A Deployment doesn't need to be paused if it is already paused.
// Otherwise it need to be paused.
func (node *deploymentNode) needsPause(cfg *desiredNodeState) (bool, error) {

	if node.resource.Spec.Paused == false {
		logrus.Debugf("Deployment %v is not currently paused.", node.resource.Name)
		return true, nil
	}

	return false, nil
}

// Since this is called as part of doing an upgrade we check if deployments need to be
// paused again as a separate call to avoid unnecessary rollouts
func (node *deploymentNode) isDifferent(cfg *desiredNodeState) (bool, error) {

	// Check replicas number
	actualReplicas := *node.resource.Spec.Replicas
	if cfg.getReplicas() != actualReplicas {
		logrus.Debugf("Different number of replicas detected, updating deployment %q", cfg.DeployName)
		return true, nil
	}

	// Check if labels are correct
	for label, value := range cfg.Labels {
		val, ok := node.resource.Labels[label]
		if !ok || val != value {
			logrus.Debugf("Different labels detected, updating deployment %q", node.resource.GetName())
			return true, nil
		}
	}

	// Check if the Variables are the desired ones
	envVars := node.resource.Spec.Template.Spec.Containers[0].Env
	desiredVars := cfg.EnvVars
	if len(envVars) != len(desiredVars) {
		logrus.Debugf("Different environmental variables detected, updating deployment %q", node.resource.GetName())
		return true, nil
	}
	for index, value := range envVars {
		if value.ValueFrom == nil {
			if desiredVars[index] != value {
				logrus.Debugf("Different environmental variables detected, updating deployment %q", node.resource.GetName())
				return true, nil
			}
		} else {
			if desiredVars[index].ValueFrom.FieldRef.FieldPath != value.ValueFrom.FieldRef.FieldPath {
				logrus.Debugf("Different environmental variables detected, updating deployment %q", node.resource.GetName())
				return true, nil
			}
		}
	}

	// Check that storage configuration is the same
	// Maybe this needs to be split into a separate method since this
	// may indicate that we need a new cluster spin up, not rolling restart
	for _, volume := range node.resource.Spec.Template.Spec.Volumes {
		if volume.Name == "elasticsearch-storage" {
			switch {
			case volume.PersistentVolumeClaim != nil && &cfg.ESNodeSpec.Storage.StorageClassName != nil:
				desiredClaimName := fmt.Sprintf("%s-%s", cfg.ClusterName, cfg.DeployName)
				if volume.PersistentVolumeClaim.ClaimName == desiredClaimName {
					return false, nil
				}

				logrus.Warn("Detected change in storage")
				return true, nil
			case volume.EmptyDir != nil && cfg.ESNodeSpec.Storage == v1alpha1.ElasticsearchStorageSpec{}:
				return false, nil
			default:
				logrus.Warn("Detected change in storage")
				return true, nil
			}
		}
	}
	return false, nil
}

// isUpdateNeeded compares existing Deployment resource and its desired state.
// If a difference is found the existing resource will be updated. This triggers
// a new rollout, which is handeled by the operator in rolling fashion.
// Currently only Image and CPU & memory Resources trigger rolling restart
func (node *deploymentNode) isUpdateNeeded(cfg *desiredNodeState) bool {
	// Only Image and Resources (CPU & memory) differences trigger rolling restart
	for _, container := range node.resource.Spec.Template.Spec.Containers {
		if container.Name == "elasticsearch" {
			// Check image of Elasticsearch container
			if container.Image != cfg.Image {
				logrus.Debugf("Resource '%s' has different container image than desired", node.resource.Name)
				return true
			}
			// Check CPU limits
			if cfg.ESNodeSpec.Resources.Limits.Cpu().Cmp(*container.Resources.Limits.Cpu()) != 0 {
				logrus.Debugf("Resource '%s' has different CPU limit than desired", node.resource.Name)
				return true
			}
			// Check memory limits
			if cfg.ESNodeSpec.Resources.Limits.Memory().Cmp(*container.Resources.Limits.Memory()) != 0 {
				logrus.Debugf("Resource '%s' has different Memory limit than desired", node.resource.Name)
				return true
			}
			// Check CPU requests
			if cfg.ESNodeSpec.Resources.Requests.Cpu().Cmp(*container.Resources.Requests.Cpu()) != 0 {
				logrus.Debugf("Resource '%s' has different CPU Request than desired", node.resource.Name)
				return true
			}
			// Check memory requests
			if cfg.ESNodeSpec.Resources.Requests.Memory().Cmp(*container.Resources.Requests.Memory()) != 0 {
				logrus.Debugf("Resource '%s' has different Memory Request than desired", node.resource.Name)
				return true
			}
		}
	}
	return false
}

func (node *deploymentNode) query() error {
	err := sdk.Get(&node.resource)
	return err
}

// constructNodeDeployment creates the deployment for the node
func (node *deploymentNode) constructNodeResource(cfg *desiredNodeState, owner metav1.OwnerReference) (runtime.Object, error) {

	// Check if deployment exists

	// FIXME: remove hardcode

	replicas := cfg.getReplicas()

	// deployment := node.resource
	deployment := apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.DeployName,
			Namespace: cfg.Namespace,
		},
	}

	progressDeadlineSeconds := int32(1800)
	deployment.ObjectMeta.Labels = cfg.getLabels()
	deployment.Spec = apps.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: cfg.getLabelSelector(),
		},
		Strategy: apps.DeploymentStrategy{
			Type: "Recreate",
		},
		ProgressDeadlineSeconds: &progressDeadlineSeconds,
		Template:                cfg.constructPodTemplateSpec(),
		Paused:                  cfg.Paused,
	}

	// if storageClass != "default" {
	// 	deployment.Spec.VolumeClaimTemplates[0].Annotations = map[string]string{
	// 		"volume.beta.kubernetes.io/storage-class": storageClass,
	// 	}
	// }
	// sset, _ := json.Marshal(deployment)
	// s := string(sset[:])

	addOwnerRefToObject(&deployment, owner)

	return &deployment, nil
}

func (node *deploymentNode) delete() error {
	err := sdk.Delete(&node.resource)
	if err != nil {
		return fmt.Errorf("Unable to delete Deployment %v: ", err)
	}
	return nil
}
