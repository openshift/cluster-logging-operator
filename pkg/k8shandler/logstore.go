package k8shandler

import (
	"fmt"

	"reflect"

	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cluster *ClusterLogging) CreateOrUpdateLogStore() (err error) {

	if cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {

		if err = cluster.createOrUpdateElasticsearchSecret(); err != nil {
			return
		}

		if err = cluster.createOrUpdateElasticsearchCR(); err != nil {
			return
		}

		elasticsearchStatus, err := cluster.getElasticsearchStatus()

		if err != nil {
			return fmt.Errorf("Failed to get Elasticsearch status for %q: %v", cluster.Name, err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if exists := cluster.Exists(); exists {
				if !reflect.DeepEqual(elasticsearchStatus, cluster.Status.LogStore.ElasticsearchStatus) {
					if printUpdateMessage {
						logrus.Info("Updating status of Elasticsearch")
						printUpdateMessage = false
					}
					cluster.Status.LogStore.ElasticsearchStatus = elasticsearchStatus
					return sdk.Update(cluster)
				}
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Elasticsearch status: %v", retryErr)
		}
	} else {
		cluster.removeElasticsearch()
	}

	return nil
}

func (cluster *ClusterLogging) removeElasticsearch() (err error) {
	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		if err = utils.RemoveSecret(cluster.Namespace, "elasticsearch"); err != nil {
			return
		}

		if err = cluster.removeElasticsearchCR("elasticsearch"); err != nil {
			return
		}
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateElasticsearchSecret() error {

	esSecret := utils.Secret(
		"elasticsearch",
		cluster.Namespace,
		map[string][]byte{
			"elasticsearch.key": utils.GetWorkingDirFileContents("elasticsearch.key"),
			"elasticsearch.crt": utils.GetWorkingDirFileContents("elasticsearch.crt"),
			"logging-es.key":    utils.GetWorkingDirFileContents("logging-es.key"),
			"logging-es.crt":    utils.GetWorkingDirFileContents("logging-es.crt"),
			"admin-key":         utils.GetWorkingDirFileContents("system.admin.key"),
			"admin-cert":        utils.GetWorkingDirFileContents("system.admin.crt"),
			"admin-ca":          utils.GetWorkingDirFileContents("ca.crt"),
		},
	)

	cluster.AddOwnerRefTo(esSecret)

	err := utils.CreateOrUpdateSecret(esSecret)
	if err != nil {
		return err
	}

	return nil
}

func (cluster *ClusterLogging) newElasticsearchCR(elasticsearchName string) *v1alpha1.Elasticsearch {

	var esNodes []v1alpha1.ElasticsearchNode

	var resources = cluster.Spec.LogStore.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultEsMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultEsMemory,
				v1.ResourceCPU:    defaultEsCpuRequest,
			},
		}
	}

	esNode := v1alpha1.ElasticsearchNode{
		Roles:        []v1alpha1.ElasticsearchNodeRole{"client", "data", "master"},
		NodeCount:    cluster.Spec.LogStore.NodeCount,
		NodeSelector: cluster.Spec.LogStore.NodeSelector,
		Resources:    *resources,
		Storage:      cluster.Spec.LogStore.ElasticsearchSpec.Storage,
	}

	// build Nodes
	esNodes = append(esNodes, esNode)

	cr := &v1alpha1.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      elasticsearchName,
			Namespace: cluster.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.ElasticsearchSpec{
			Spec: v1alpha1.ElasticsearchNodeSpec{
				Image:     utils.GetComponentImage("elasticsearch"),
				Resources: *resources,
			},
			Nodes:            esNodes,
			ManagementState:  v1alpha1.ManagementStateManaged,
			RedundancyPolicy: cluster.Spec.LogStore.ElasticsearchSpec.RedundancyPolicy,
		},
	}

	cluster.AddOwnerRefTo(cr)

	return cr
}

func (cluster *ClusterLogging) removeElasticsearchCR(elasticsearchName string) error {

	esCr := cluster.newElasticsearchCR(elasticsearchName)

	err := sdk.Delete(esCr)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v elasticsearch CR for %q: %v", elasticsearchName, cluster.Name, err)
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateElasticsearchCR() (err error) {

	esCR := cluster.newElasticsearchCR("elasticsearch")

	err = sdk.Create(esCR)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Elasticsearch CR: %v", err)
	}

	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return updateElasticsearchCRIfRequired(esCR)
		})
		if retryErr != nil {
			return retryErr
		}
	}
	return nil
}

func updateElasticsearchCRIfRequired(desired *v1alpha1.Elasticsearch) (err error) {
	current := desired.DeepCopy()

	if err = sdk.Get(current); err != nil {
		if apierrors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Elasticsearch CR: %v", err)
	}

	current, different := isElasticsearchCRDifferent(current, desired)

	if different {
		if err = sdk.Update(current); err != nil {
			return err
		}
	}

	return nil
}

func isElasticsearchCRDifferent(current *v1alpha1.Elasticsearch, desired *v1alpha1.Elasticsearch) (*v1alpha1.Elasticsearch, bool) {

	different := false

	if current.Spec.Spec.Image != desired.Spec.Spec.Image {
		logrus.Infof("Elasticsearch image change found, updating %v", current.Name)
		current.Spec.Spec.Image = desired.Spec.Spec.Image
		different = true
	}

	if current.Spec.RedundancyPolicy != desired.Spec.RedundancyPolicy {
		logrus.Infof("Elasticsearch redundancy policy change found, updating %v", current.Name)
		current.Spec.RedundancyPolicy = desired.Spec.RedundancyPolicy
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Spec.Resources, desired.Spec.Spec.Resources) {
		logrus.Infof("Elasticsearch resources change found, updating %v", current.Name)
		current.Spec.Spec.Resources = desired.Spec.Spec.Resources
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Nodes, desired.Spec.Nodes) {
		logrus.Infof("Elasticsearch node configuration change found, updating %v", current.Name)
		current.Spec.Nodes = desired.Spec.Nodes
		different = true
	}

	return current, different
}
