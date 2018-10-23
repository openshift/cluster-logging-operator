package k8shandler

import (
	"fmt"
	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"reflect"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateOrUpdateLogStore(cluster *logging.ClusterLogging) (err error) {

	if cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {

		if err = createOrUpdateElasticsearchSecret(cluster); err != nil {
			return
		}

		if err = createOrUpdateElasticsearchCR(cluster); err != nil {
			return
		}

		elasticsearchStatus, err := getElasticsearchStatus(cluster.Namespace)

		if err != nil {
			return fmt.Errorf("Failed to get status for Elasticsearch: %v", err)
		}

		if !reflect.DeepEqual(elasticsearchStatus, cluster.Status.LogStore.ElasticsearchStatus) {
			logrus.Infof("Updating status of Elasticsearch")
			cluster.Status.LogStore.ElasticsearchStatus = elasticsearchStatus

			if err = sdk.Update(cluster); err != nil {
				return fmt.Errorf("Failed to update Cluster Logging Elasticsearch status: %v", err)
			}
		}
	}

	return nil
}

func createOrUpdateElasticsearchSecret(logging *logging.ClusterLogging) error {

	esSecret := utils.Secret(
		"elasticsearch",
		logging.Namespace,
		map[string][]byte{
			"elasticsearch.key": utils.GetFileContents("/tmp/_working_dir/elasticsearch.key"),
			"elasticsearch.crt": utils.GetFileContents("/tmp/_working_dir/elasticsearch.crt"),
			"logging-es.key":    utils.GetFileContents("/tmp/_working_dir/logging-es.key"),
			"logging-es.crt":    utils.GetFileContents("/tmp/_working_dir/logging-es.crt"),
			"admin-key":         utils.GetFileContents("/tmp/_working_dir/system.admin.key"),
			"admin-cert":        utils.GetFileContents("/tmp/_working_dir/system.admin.crt"),
			"admin-ca":          utils.GetFileContents("/tmp/_working_dir/ca.crt"),
		},
	)

	utils.AddOwnerRefToObject(esSecret, utils.AsOwner(logging))

	err := sdk.Create(esSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Elasticsearch secret: %v", err)
	}

	return nil
}

func getElasticsearchCR(logging *logging.ClusterLogging, elasticsearchName string) *v1alpha1.Elasticsearch {

	var esNodes []v1alpha1.ElasticsearchNode

	esNode := v1alpha1.ElasticsearchNode{
		Roles:        []v1alpha1.ElasticsearchNodeRole{"client", "data", "master"},
		Replicas:     logging.Spec.LogStore.Replicas,
		NodeSelector: logging.Spec.LogStore.NodeSelector,
		Spec: v1alpha1.ElasticsearchNodeSpec{
			Resources: logging.Spec.LogStore.Resources,
		},
		Storage: logging.Spec.LogStore.ElasticsearchSpec.Storage,
	}

	if esNode.Storage.VolumeClaimTemplate != nil {
		esNode.Storage.VolumeClaimTemplate.ObjectMeta = metav1.ObjectMeta{
			Name:      elasticsearchName,
			Namespace: logging.Namespace,
			Labels: map[string]string{
				"logging-infra": "support",
			},
		}
	}

	// build Nodes
	esNodes = append(esNodes, esNode)

	cr := &v1alpha1.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      elasticsearchName,
			Namespace: logging.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.ElasticsearchSpec{
			Spec: v1alpha1.ElasticsearchNodeSpec{
				Image: utils.GetComponentImage("elasticsearch"),
			},
			Nodes:              esNodes,
			ServiceAccountName: "elasticsearch",
			ConfigMapName:      "elasticsearch",
			//SecretName: "elasticsearch",
		},
	}

	utils.AddOwnerRefToObject(cr, utils.AsOwner(logging))

	return cr
}

func createOrUpdateElasticsearchCR(logging *logging.ClusterLogging) (err error) {

	if utils.AllInOne(logging) {
		esCR := getElasticsearchCR(logging, "elasticsearch")

		err = sdk.Create(esCR)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Elasticsearch CR: %v", err)
		}

		if err = updateElasticsearchCRIfRequired(esCR); err != nil {
			return
		}
	} else {
		esCR := getElasticsearchCR(logging, "elasticsearch-app")

		err = sdk.Create(esCR)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Elasticsearch CR: %v", err)
		}

		if err = updateElasticsearchCRIfRequired(esCR); err != nil {
			return
		}

		esInfraCR := getElasticsearchCR(logging, "elasticsearch-infra")

		err = sdk.Create(esInfraCR)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Elasticsearch Infra CR: %v", err)
		}

		if err = updateElasticsearchCRIfRequired(esInfraCR); err != nil {
			return
		}
	}

	return nil
}

func updateElasticsearchCRIfRequired(desired *v1alpha1.Elasticsearch) (err error) {
	current := desired.DeepCopy()

	if err = sdk.Get(current); err != nil {
		return fmt.Errorf("Failed to get Elasticsearch CR: %v", err)
	}

	current, different := isElasticsearchCRDifferent(current, desired)

	if different {
		if err = sdk.Update(current); err != nil {
			return fmt.Errorf("Failed to update Elasticsearch CR: %v", err)
		}
	}

	return nil
}

func isElasticsearchCRDifferent(current *v1alpha1.Elasticsearch, desired *v1alpha1.Elasticsearch) (*v1alpha1.Elasticsearch, bool) {

	different := false

	//TODO: populate this

	return current, different
}
