package k8shandler

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	es "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateVisualization reconciles visualization component for cluster logging
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateVisualization() (err error) {
	if clusterRequest.Cluster.Spec.Visualization == nil || clusterRequest.Cluster.Spec.Visualization.Type == "" {
		return clusterRequest.removeKibana()
	}

	if err = clusterRequest.createOrUpdateKibanaCR(); err != nil {
		return
	}

	if err = clusterRequest.createOrUpdateKibanaSecret(); err != nil {
		return
	}

	if err = clusterRequest.UpdateKibanaStatus(); err != nil {
		return
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) UpdateKibanaStatus() (err error) {
	kibanaStatus, err := clusterRequest.getKibanaStatus()
	if err != nil {
		log.Error(err, "Failed to get Kibana status for", "clusterName", clusterRequest.Cluster.Name)
		return
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if !compareKibanaStatus(kibanaStatus, clusterRequest.Cluster.Status.Visualization.KibanaStatus) {
			clusterRequest.Cluster.Status.Visualization.KibanaStatus = kibanaStatus
			return clusterRequest.UpdateStatus(clusterRequest.Cluster)
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging Kibana status: %v", retryErr)
	}
	return nil
}

func compareKibanaStatus(lhs, rhs []es.KibanaStatus) bool {
	// there should only ever be a single kibana status object
	if len(lhs) != len(rhs) {
		return false
	}

	if len(lhs) > 0 {
		for index := range lhs {
			if lhs[index].Deployment != rhs[index].Deployment {
				return false
			}

			if lhs[index].Replicas != rhs[index].Replicas {
				return false
			}

			if len(lhs[index].ReplicaSets) != len(rhs[index].ReplicaSets) {
				return false
			}

			if len(lhs[index].ReplicaSets) > 0 {
				if !reflect.DeepEqual(lhs[index].ReplicaSets, rhs[index].ReplicaSets) {
					return false
				}
			}

			if len(lhs[index].Pods) != len(rhs[index].Pods) {
				return false
			}

			if len(lhs[index].Pods) > 0 {
				if !reflect.DeepEqual(lhs[index].Pods, rhs[index].Pods) {
					return false
				}
			}

			if len(lhs[index].Conditions) != len(rhs[index].Conditions) {
				return false
			}

			if len(lhs[index].Conditions) > 0 {
				if !reflect.DeepEqual(lhs[index].Conditions, rhs[index].Conditions) {
					return false
				}
			}
		}
	}

	return true
}

func (clusterRequest *ClusterLoggingRequest) removeKibana() (err error) {
	if clusterRequest.isManaged() {
		name := "kibana"
		proxyName := "kibana-proxy"

		if err = clusterRequest.removeKibanaCR(); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret(proxyName); err != nil {
			return
		}

	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaSecret() error {
	var secrets = map[string][]byte{}
	_ = Syncronize(func() error {
		secrets = map[string][]byte{
			"ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"key":  utils.GetWorkingDirFileContents("system.logging.kibana.key"),
			"cert": utils.GetWorkingDirFileContents("system.logging.kibana.crt"),
		}
		return nil
	})
	kibanaSecret := NewSecret(
		"kibana",
		clusterRequest.Cluster.Namespace,
		secrets)

	utils.AddOwnerRefToObject(kibanaSecret, utils.AsOwner(clusterRequest.Cluster))

	err := clusterRequest.CreateOrUpdateSecret(kibanaSecret)
	if err != nil {
		return err
	}

	_ = Syncronize(func() error {
		secrets = map[string][]byte{
			"session-secret": utils.GetWorkingDirFileContents(constants.KibanaSessionSecretName),
			"server-key":     utils.GetWorkingDirFileContents("kibana-internal.key"),
			"server-cert":    utils.GetWorkingDirFileContents("kibana-internal.crt"),
		}
		return nil
	})
	proxySecret := NewSecret(
		"kibana-proxy",
		clusterRequest.Cluster.Namespace,
		secrets)

	utils.AddOwnerRefToObject(proxySecret, utils.AsOwner(clusterRequest.Cluster))

	err = clusterRequest.CreateOrUpdateSecret(proxySecret)
	if err != nil {
		return err
	}

	return nil
}

func newKibanaCustomResource(cluster *logging.ClusterLogging, kibanaName string) *es.Kibana {
	visSpec := cluster.Spec.Visualization

	resources := visSpec.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: defaultKibanaMemory,
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultKibanaMemory,
				v1.ResourceCPU:    defaultKibanaCpuRequest,
			},
		}
	}

	var replicas int32
	if visSpec.Replicas != nil {
		replicas = *visSpec.Replicas
	} else {
		if cluster.Spec.LogStore != nil && cluster.Spec.LogStore.ElasticsearchSpec.NodeCount > 0 {
			replicas = 1
		} else {
			replicas = 0
		}
	}

	proxyResources := visSpec.ProxySpec.Resources
	if proxyResources == nil {
		proxyResources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultKibanaProxyMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultKibanaProxyMemory,
				v1.ResourceCPU:    defaultKibanaProxyCpuRequest,
			},
		}
	}

	cr := &es.Kibana{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kibana",
			APIVersion: es.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kibanaName,
			Namespace: cluster.Namespace,
		},
		Spec: es.KibanaSpec{
			ManagementState: es.ManagementStateManaged,
			Replicas:        replicas,
			Resources:       resources,
			NodeSelector:    visSpec.NodeSelector,
			Tolerations:     visSpec.Tolerations,
			ProxySpec: es.ProxySpec{
				Resources: proxyResources,
			},
		},
	}

	utils.AddOwnerRefToObject(cr, utils.AsOwner(cluster))
	return cr
}

func (clusterRequest *ClusterLoggingRequest) getKibanaCR() (*es.Kibana, error) {
	var kb = &es.Kibana{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kibana",
			APIVersion: es.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.Client.Get(context.TODO(),
		client.ObjectKey{
			Namespace: clusterRequest.Cluster.Namespace,
			Name:      constants.KibanaName,
		}, kb)

	if err != nil {
		return nil, err
	}
	return kb, nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaCR() error {
	cr := newKibanaCustomResource(clusterRequest.Cluster, constants.KibanaName)

	err := clusterRequest.Create(cr)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("unable to create kibana cr. E: %s", err.Error())
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateKibanaCRDIfRequired(cr)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) updateKibanaCRDIfRequired(cr *es.Kibana) error {
	current := &es.Kibana{}

	if err := clusterRequest.Get(cr.Name, current); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get kibana CR: %s", err.Error())
	}

	if isKibanaCRDDifferent(current, cr) {
		if err := clusterRequest.Update(current); err != nil {
			return err
		}
	}
	return nil
}

// Check if CRD is not equal to desired. If it's not, returns true and
// updates the CURRENT crd to be equal to DESIRED
func isKibanaCRDDifferent(current *es.Kibana, desired *es.Kibana) bool {
	different := false

	if current.Spec.ManagementState != desired.Spec.ManagementState {
		current.Spec.ManagementState = desired.Spec.ManagementState
		different = true
	}

	if current.Spec.Replicas != desired.Spec.Replicas {
		current.Spec.Replicas = desired.Spec.Replicas
		different = true
	}

	if !utils.AreMapsSame(current.Spec.NodeSelector, desired.Spec.NodeSelector) {
		current.Spec.NodeSelector = desired.Spec.NodeSelector
		different = true
	}

	if !utils.AreTolerationsSame(current.Spec.Tolerations, desired.Spec.Tolerations) {
		current.Spec.Tolerations = desired.Spec.Tolerations
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Resources, desired.Spec.Resources) {
		current.Spec.Resources = desired.Spec.Resources
		different = true
	}

	if !reflect.DeepEqual(current.Spec.ProxySpec, desired.Spec.ProxySpec) {
		current.Spec.ProxySpec = desired.Spec.ProxySpec
		different = true
	}

	return different
}

func (clusterRequest *ClusterLoggingRequest) removeKibanaCR() error {
	cr := &es.Kibana{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KibanaName,
			Namespace: clusterRequest.Cluster.Namespace,
		},
	}

	err := clusterRequest.Delete(cr)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("unable to delete kibana cr. E: %s", err.Error())
	}

	return nil
}
