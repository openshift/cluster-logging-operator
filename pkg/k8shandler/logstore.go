package k8shandler

import (
	"fmt"

	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/util/retry"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	//OOTB should be 3
	ElasticsearchNodeCountDefault int32 = 1
)

func (cluster *ClusterLogging) CreateOrUpdateLogStore(stack *logging.StackSpec) (err error) {
	if stack.Store == nil {
		logrus.Debugf("Store is not spec'd for stack '%s'", stack.Name)
		return
	}

	if err = createOrUpdateElasticsearchSecret(cluster); err != nil {
		return
	}

	if err = cluster.createOrUpdateElasticsearchCR(stack); err != nil {
		return
	}

	elasticsearchStatus, err := getElasticsearchStatus(cluster.Namespace)

	if err != nil {
		return fmt.Errorf("Failed to get status for Elasticsearch: %v", err)
	}

	printUpdateMessage := true
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if exists := utils.DoesClusterLoggingExist(cluster.ClusterLogging); exists {
			if !reflect.DeepEqual(elasticsearchStatus, cluster.ClusterLogging.Status.LogStore.ElasticsearchStatus) {
				if printUpdateMessage {
					logrus.Info("Updating status of Elasticsearch")
					printUpdateMessage = false
				}
				cluster.ClusterLogging.Status.LogStore.ElasticsearchStatus = elasticsearchStatus
				return sdk.Update(cluster)
			}
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging Elasticsearch status: %v", retryErr)
	}

	return nil
}

func createOrUpdateElasticsearchSecret(logging *ClusterLogging) error {

	esSecret := utils.NewSecret(
		Elasticsearch,
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

	logging.addOwnerRefTo(esSecret)

	err := sdk.Create(esSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Elasticsearch secret: %v", err)
	}

	return nil
}

func (logging *ClusterLogging) getElasticsearchCR(stack *logging.StackSpec) *v1alpha1.Elasticsearch {

	elasticsearchName := logging.getElasticsearchName(stack.Name)
	var esNodes []v1alpha1.ElasticsearchNode

	nodeCount := stack.Store.Replicas
	if nodeCount == 0 {
		nodeCount = ElasticsearchNodeCountDefault
	}
	esNode := v1alpha1.ElasticsearchNode{
		Roles:        []v1alpha1.ElasticsearchNodeRole{"client", "data", "master"},
		Replicas:     nodeCount,
		NodeSelector: stack.Store.NodeSelector,
		Spec: v1alpha1.ElasticsearchNodeSpec{
			Resources: stack.Store.Resources,
		},
		Storage: stack.Store.Storage,
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
				Image: utils.GetComponentImage(Elasticsearch),
			},
			Nodes:             esNodes,
			ManagementState:   v1alpha1.ManagementStateManaged,
			ReplicationPolicy: stack.Store.ReplicationPolicy,
		},
	}

	logging.addOwnerRefTo(cr)

	return cr
}

func (cluster *ClusterLogging) createOrUpdateElasticsearchCR(stack *logging.StackSpec) (err error) {

	if stack.Store == nil {
		return nil
	}

	esCR := cluster.getElasticsearchCR(stack)

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
		logrus.Infof("Elasticsearch image change found, updating %q", current.Name)
		current.Spec.Spec.Image = desired.Spec.Spec.Image
		different = true
	}

	return current, different
}
