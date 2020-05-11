package k8shandler

import (
	"context"
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	es "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateVisualization reconciles visualization component for cluster logging
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateVisualization(proxyConfig *configv1.Proxy) (err error) {
	if clusterRequest.cluster.Spec.Visualization == nil || clusterRequest.cluster.Spec.Visualization.Type == "" {
		if err = clusterRequest.removeKibana(); err != nil {
			return
		}
		return nil
	}

	//TODO: Remove this in the next release after removing old kibana code completely
	if err = clusterRequest.removeOldKibana(); err != nil {
		return
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
		logrus.Errorf("Failed to get Kibana status for %q: %v", clusterRequest.cluster.Name, err)
		return
	}

	printUpdateMessage := true
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if !compareKibanaStatus(kibanaStatus, clusterRequest.cluster.Status.Visualization.KibanaStatus) {
			if printUpdateMessage {
				logrus.Info("Updating status of Kibana")
				printUpdateMessage = false
			}
			clusterRequest.cluster.Status.Visualization.KibanaStatus = kibanaStatus
			return clusterRequest.UpdateStatus(clusterRequest.cluster)
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

//TODO: Remove this in the next release after removing old kibana code completely
// since kibana is now handled by the Elasticsearch Operator
func (clusterRequest *ClusterLoggingRequest) removeOldKibana() (err error) {
	if clusterRequest.isManaged() {
		name := "kibana"
		proxyName := "kibana-proxy"

		if err = clusterRequest.RemoveDeployment(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveOAuthClient(proxyName); err != nil {
			return
		}

		if err = clusterRequest.RemoveRoute(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("sharing-config"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(constants.KibanaTrustedCAName); err != nil {
			return
		}

		if err = clusterRequest.RemoveService(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveServiceAccount(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConsoleExternalLogLink(name); err != nil {
			return
		}

	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaSecret() error {

	kibanaSecret := NewSecret(
		"kibana",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"key":  utils.GetWorkingDirFileContents("system.logging.kibana.key"),
			"cert": utils.GetWorkingDirFileContents("system.logging.kibana.crt"),
		})

	utils.AddOwnerRefToObject(kibanaSecret, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.CreateOrUpdateSecret(kibanaSecret)
	if err != nil {
		return err
	}

	var sessionSecret []byte

	sessionSecret = utils.GetWorkingDirFileContents("kibana-session-secret")
	if sessionSecret == nil {
		sessionSecret = utils.GetRandomWord(32)
	}

	proxySecret := NewSecret(
		"kibana-proxy",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"session-secret": sessionSecret,
			"server-key":     utils.GetWorkingDirFileContents("kibana-internal.key"),
			"server-cert":    utils.GetWorkingDirFileContents("kibana-internal.crt"),
		})

	utils.AddOwnerRefToObject(proxySecret, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.CreateOrUpdateSecret(proxySecret)
	if err != nil {
		return err
	}

	return nil
}

func newKibanaCustomResource(cluster *logging.ClusterLogging, kibanaName string) *es.Kibana {
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
			Replicas:        1,
			Resources: &v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceMemory: defaultKibanaMemory,
				},
				Requests: v1.ResourceList{
					v1.ResourceMemory: defaultKibanaMemory,
					v1.ResourceCPU:    defaultKibanaCpuRequest,
				},
			},
		},
		Status: []es.KibanaStatus{},
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

	err := clusterRequest.client.Get(context.TODO(),
		client.ObjectKey{
			Namespace: clusterRequest.cluster.Namespace,
			Name:      constants.KibanaName,
		}, kb)

	if err != nil {
		return nil, err
	}
	return kb, nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaCR() error {
	cr := newKibanaCustomResource(clusterRequest.cluster, constants.KibanaName)

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

	if current.Spec.Image != desired.Spec.Image {
		current.Spec.Image = desired.Spec.Image
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Resources, desired.Spec.Resources) {
		current.Spec.Resources = desired.Spec.Resources
		different = true
	}

	return different
}

func (clusterRequest *ClusterLoggingRequest) removeKibanaCR() error {
	cr := newKibanaCustomResource(clusterRequest.cluster, constants.KibanaName)

	err := clusterRequest.Delete(cr)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("unable to delete kibana cr. E: %s", err.Error())
	}

	return nil
}

func HasCLORef(object metav1.Object, request *ClusterLoggingRequest) bool {
	refs := object.GetOwnerReferences()
	for _, r := range refs {
		bref := utils.AsOwner(request.cluster)
		if AreRefsEqual(&r, &bref) {
			return true
		}
	}
	return false
}

func AreRefsEqual(l *metav1.OwnerReference, r *metav1.OwnerReference) bool {
	if l.Name == r.Name &&
		l.APIVersion == r.APIVersion &&
		l.Kind == r.Kind &&
		*l.Controller == *r.Controller {
		return true
	}
	return false
}
