package k8shandler

import (
	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-logging-operator/pkg/utils"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateOrUpdateLogStore(logging *logging.ClusterLogging) error {
	createOrUpdateElasticsearchSecret(logging)
	return createOrUpdateElasticsearchCR(logging)
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
		})

	utils.AddOwnerRefToObject(esSecret, utils.AsOwner(logging))

	err := sdk.Create(esSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Fatalf("Failure constructing Elasticsearch secret: %v", err)
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
			APIVersion: "elasticsearch.redhat.com/v1alpha1",
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

func createOrUpdateElasticsearchCR(logging *logging.ClusterLogging) error {

	if utils.AllInOne(logging) {
		esCR := getElasticsearchCR(logging, "elasticsearch")

		err := sdk.Create(esCR)
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Fatalf("Failure creating Elasticsearch CR: %v", err)
		}
	} else {
		esCR := getElasticsearchCR(logging, "elasticsearch-app")

		err := sdk.Create(esCR)
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Fatalf("Failure creating Elasticsearch CR: %v", err)
		}

		esInfraCR := getElasticsearchCR(logging, "elasticsearch-infra")

		err = sdk.Create(esInfraCR)
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Fatalf("Failure creating Elasticsearch Infra CR: %v", err)
		}
	} /*else if errors.IsAlreadyExists(err) {
	  // Get existing configMap to check if it is same as what we want
	  existingCR := &v1alpha1.Elasticsearch{
	    ObjectMeta: metav1.ObjectMeta{
	      Name: "elasticsearch",
	      Namespace: logging.Namespace,
	    },
	    TypeMeta: metav1.TypeMeta{
	      Kind: "Elasticsearch",
	      APIVersion: "elasticsearch.redhat.com/v1alpha1",
	    },
	  }

	  err = sdk.Get(existingCR)
	  if err != nil {
	    logrus.Fatalf("Unable to get Elasticsearch CR: %v", err)
	  }

	  logrus.Infof("Found existing CR: %v", existingCR)

	  // TODO: Compare existing CR labels, selectors and port
	}*/

	return nil
}
