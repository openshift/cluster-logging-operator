package k8shandler

import (
	"fmt"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateLogStore() (err error) {

	if clusterRequest.cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {

		cluster := clusterRequest.cluster

		if err = clusterRequest.createOrUpdateElasticsearchSecret(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateElasticsearchCR(); err != nil {
			return
		}

		elasticsearchStatus, err := clusterRequest.getElasticsearchStatus()

		if err != nil {
			return fmt.Errorf("Failed to get Elasticsearch status for %q: %v", cluster.Name, err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if !reflect.DeepEqual(elasticsearchStatus, cluster.Status.LogStore.ElasticsearchStatus) {
				if printUpdateMessage {
					logrus.Info("Updating status of Elasticsearch")
					printUpdateMessage = false
				}
				cluster.Status.LogStore.ElasticsearchStatus = elasticsearchStatus
				return clusterRequest.Update(cluster)
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Elasticsearch status: %v", retryErr)
		}
	} else {
		clusterRequest.removeElasticsearch()
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeElasticsearch() (err error) {
	// is this even required here anymore?
	if clusterRequest.isManaged() {
		if err = clusterRequest.RemoveSecret("elasticsearch"); err != nil {
			return
		}

		if err = clusterRequest.removeElasticsearchCR("elasticsearch"); err != nil {
			return
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateElasticsearchSecret() error {

	esSecret := NewSecret(
		"elasticsearch",
		clusterRequest.cluster.Namespace,
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

	utils.AddOwnerRefToObject(esSecret, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.CreateOrUpdateSecret(esSecret)
	if err != nil {
		return err
	}

	return nil
}

func newElasticsearchCR(cluster *logging.ClusterLogging, elasticsearchName string) *elasticsearch.Elasticsearch {

	var esNodes []elasticsearch.ElasticsearchNode

	var resources = cluster.Spec.LogStore.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: defaultEsMemory,
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultEsMemory,
				v1.ResourceCPU:    defaultEsCpuRequest,
			},
		}
	}

	if cluster.Spec.LogStore.NodeCount > 3 {

		masterNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
			NodeCount: 3,
			Storage:   cluster.Spec.LogStore.ElasticsearchSpec.Storage,
		}

		esNodes = append(esNodes, masterNode)

		dataNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data"},
			NodeCount: cluster.Spec.LogStore.NodeCount - 3,
			Storage:   cluster.Spec.LogStore.ElasticsearchSpec.Storage,
		}

		esNodes = append(esNodes, dataNode)

	} else {

		esNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
			NodeCount: cluster.Spec.LogStore.NodeCount,
			Storage:   cluster.Spec.LogStore.ElasticsearchSpec.Storage,
		}

		// build Nodes
		esNodes = append(esNodes, esNode)
	}

	redundancyPolicy := cluster.Spec.LogStore.ElasticsearchSpec.RedundancyPolicy
	if redundancyPolicy == "" {
		redundancyPolicy = elasticsearch.ZeroRedundancy
	}

	cr := &elasticsearch.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      elasticsearchName,
			Namespace: cluster.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.SchemeGroupVersion.String(),
		},
		Spec: elasticsearch.ElasticsearchSpec{
			Spec: elasticsearch.ElasticsearchNodeSpec{
				Image:        utils.GetComponentImage("elasticsearch"),
				Resources:    *resources,
				NodeSelector: cluster.Spec.LogStore.NodeSelector,
			},
			Nodes:            esNodes,
			ManagementState:  elasticsearch.ManagementStateManaged,
			RedundancyPolicy: redundancyPolicy,
		},
	}

	utils.AddOwnerRefToObject(cr, utils.AsOwner(cluster))

	return cr
}

func (clusterRequest *ClusterLoggingRequest) removeElasticsearchCR(elasticsearchName string) error {

	esCr := newElasticsearchCR(clusterRequest.cluster, elasticsearchName)

	err := clusterRequest.Delete(esCr)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v elasticsearch CR for %q: %v", elasticsearchName, clusterRequest.cluster.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateElasticsearchCR() (err error) {

	esCR := newElasticsearchCR(clusterRequest.cluster, "elasticsearch")

	err = clusterRequest.Create(esCR)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Elasticsearch CR: %v", err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateElasticsearchCRIfRequired(esCR)
		})
		if retryErr != nil {
			return retryErr
		}
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) updateElasticsearchCRIfRequired(desired *elasticsearch.Elasticsearch) (err error) {
	current := &elasticsearch.Elasticsearch{}

	if err = clusterRequest.Get(desired.Name, current); err != nil {
		if apierrors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Elasticsearch CR: %v", err)
	}

	if current, different := isElasticsearchCRDifferent(current, desired); different {
		if err = clusterRequest.Update(current); err != nil {
			return err
		}
	}

	return nil
}

func isElasticsearchCRDifferent(current *elasticsearch.Elasticsearch, desired *elasticsearch.Elasticsearch) (*elasticsearch.Elasticsearch, bool) {

	different := false

	if !utils.AreSelectorsSame(current.Spec.Spec.NodeSelector, desired.Spec.Spec.NodeSelector) {
		logrus.Infof("Elasticsearch nodeSelector change found, updating '%s'", current.Name)
		current.Spec.Spec.NodeSelector = desired.Spec.Spec.NodeSelector
		different = true
	}

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

	if nodes, ok := areNodesDifferent(current.Spec.Nodes, desired.Spec.Nodes); ok {
		logrus.Infof("Elasticsearch node configuration change found, updating %v", current.Name)
		current.Spec.Nodes = nodes
		different = true
	}

	return current, different
}

func areNodesDifferent(current, desired []elasticsearch.ElasticsearchNode) ([]elasticsearch.ElasticsearchNode, bool) {

	different := false

	// nodes were removed
	if len(current) == 0 {
		return desired, true
	}

	if len(current) != len(desired) {
		return desired, true
	}

	foundRoleMatch := false
	for nodeIndex := 0; nodeIndex < len(desired); nodeIndex++ {
		for _, node := range current {
			if areNodeRolesSame(node, desired[nodeIndex]) {
				if updatedNode, isDifferent := isNodeDifferent(node, desired[nodeIndex]); isDifferent {
					desired[nodeIndex] = updatedNode
					different = true
				}
				foundRoleMatch = true
			}
		}
	}

	// if we didn't find a role match, then that means changes were made
	if !foundRoleMatch {
		different = true
	}

	return desired, different
}

func areNodeRolesSame(lhs, rhs elasticsearch.ElasticsearchNode) bool {

	if len(lhs.Roles) != len(rhs.Roles) {
		return false
	}

	lhsClient := false
	lhsData := false
	lhsMaster := false

	rhsClient := false
	rhsData := false
	rhsMaster := false

	for _, role := range lhs.Roles {
		if role == elasticsearch.ElasticsearchRoleClient {
			lhsClient = true
		}

		if role == elasticsearch.ElasticsearchRoleData {
			lhsData = true
		}

		if role == elasticsearch.ElasticsearchRoleMaster {
			lhsMaster = true
		}
	}

	for _, role := range rhs.Roles {
		if role == elasticsearch.ElasticsearchRoleClient {
			rhsClient = true
		}

		if role == elasticsearch.ElasticsearchRoleData {
			rhsData = true
		}

		if role == elasticsearch.ElasticsearchRoleMaster {
			rhsMaster = true
		}
	}

	return (lhsClient == rhsClient) && (lhsData == rhsData) && (lhsMaster == rhsMaster)
}

func isNodeDifferent(current, desired elasticsearch.ElasticsearchNode) (elasticsearch.ElasticsearchNode, bool) {

	different := false

	// check the different components that we normally set instead of using reflect
	// ignore the GenUUID if we aren't setting it.
	if desired.GenUUID == nil {
		desired.GenUUID = current.GenUUID
	}

	if !reflect.DeepEqual(current, desired) {
		current = desired
		different = true
	}

	return current, different
}
