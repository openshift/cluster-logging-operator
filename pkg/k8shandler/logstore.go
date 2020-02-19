package k8shandler

import (
	"fmt"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler/indexmanagement"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
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

const (
	elasticsearchResourceName = "elasticsearch"
)

func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateLogStore() (err error) {
	if clusterRequest.cluster.Spec.LogStore == nil || clusterRequest.cluster.Spec.LogStore.Type == "" {
		if err = clusterRequest.removeElasticsearch(); err != nil {
			return
		}
		return nil
	}
	if clusterRequest.cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {

		cluster := clusterRequest.cluster

		if err = clusterRequest.createOrUpdateElasticsearchSecret(); err != nil {
			return nil
		}

		if err = clusterRequest.createOrUpdateElasticsearchCR(); err != nil {
			return nil
		}

		elasticsearchStatus, err := clusterRequest.getElasticsearchStatus()

		if err != nil {
			return fmt.Errorf("Failed to get Elasticsearch status for %q: %v", cluster.Name, err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if !compareElasticsearchStatus(elasticsearchStatus, cluster.Status.LogStore.ElasticsearchStatus) {
				if printUpdateMessage {
					logrus.Info("Updating status of Elasticsearch")
					printUpdateMessage = false
				}
				cluster.Status.LogStore.ElasticsearchStatus = elasticsearchStatus
				return clusterRequest.UpdateStatus(cluster)
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Elasticsearch status: %v", retryErr)
		}
	}

	return nil
}

func compareElasticsearchStatus(lhs, rhs []logging.ElasticsearchStatus) bool {
	// there should only ever be a single elasticsearch status object
	if len(lhs) != len(rhs) {
		return false
	}

	if len(lhs) > 0 {
		for index, _ := range lhs {
			if lhs[index].ClusterName != rhs[index].ClusterName {
				return false
			}

			if lhs[index].NodeCount != rhs[index].NodeCount {
				return false
			}

			if lhs[index].ClusterHealth != rhs[index].ClusterHealth {
				return false
			}

			if lhs[index].Cluster != rhs[index].Cluster {
				return false
			}

			if lhs[index].ShardAllocationEnabled != rhs[index].ShardAllocationEnabled {
				return false
			}

			if len(lhs[index].Pods) != len(rhs[index].Pods) {
				return false
			}

			if len(lhs[index].Pods) > 0 {
				if !reflect.DeepEqual(lhs[index].Pods, rhs[index].Pods) {
					return false
				}
			}

			if len(lhs[index].ClusterConditions) != len(rhs[index].ClusterConditions) {
				return false
			}

			if len(lhs[index].ClusterConditions) > 0 {
				if !reflect.DeepEqual(lhs[index].ClusterConditions, rhs[index].ClusterConditions) {
					return false
				}
			}

			if len(lhs[index].NodeConditions) != len(rhs[index].NodeConditions) {
				return false
			}

			if len(lhs[index].NodeConditions) > 0 {
				if !reflect.DeepEqual(lhs[index].NodeConditions, rhs[index].NodeConditions) {
					return false
				}
			}
		}
	}

	return true
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
func LoadElasticsearchSecretMap() map[string][]byte {
	return map[string][]byte{
		"elasticsearch.key": utils.GetWorkingDirFileContents("elasticsearch.key"),
		"elasticsearch.crt": utils.GetWorkingDirFileContents("elasticsearch.crt"),
		"logging-es.key":    utils.GetWorkingDirFileContents("logging-es.key"),
		"logging-es.crt":    utils.GetWorkingDirFileContents("logging-es.crt"),
		"admin-key":         utils.GetWorkingDirFileContents("system.admin.key"),
		"admin-cert":        utils.GetWorkingDirFileContents("system.admin.crt"),
		"admin-ca":          utils.GetWorkingDirFileContents("ca.crt"),
	}
}
func (clusterRequest *ClusterLoggingRequest) createOrUpdateElasticsearchSecret() error {

	esSecret := NewSecret(
		elasticsearchResourceName,
		clusterRequest.cluster.Namespace,
		LoadElasticsearchSecretMap(),
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
	logStoreSpec := logging.LogStoreSpec{}
	if cluster.Spec.LogStore != nil {
		logStoreSpec = *cluster.Spec.LogStore
	}
	var resources = logStoreSpec.Resources
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

	if logStoreSpec.NodeCount > 3 {

		masterNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
			NodeCount: 3,
			Storage:   logStoreSpec.ElasticsearchSpec.Storage,
		}

		esNodes = append(esNodes, masterNode)

		dataNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data"},
			NodeCount: logStoreSpec.NodeCount - 3,
			Storage:   logStoreSpec.ElasticsearchSpec.Storage,
		}

		esNodes = append(esNodes, dataNode)

	} else {

		esNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
			NodeCount: logStoreSpec.NodeCount,
			Storage:   logStoreSpec.ElasticsearchSpec.Storage,
		}

		// build Nodes
		esNodes = append(esNodes, esNode)
	}

	redundancyPolicy := logStoreSpec.ElasticsearchSpec.RedundancyPolicy
	if redundancyPolicy == "" {
		redundancyPolicy = elasticsearch.ZeroRedundancy
	}

	indexManagementSpec := indexmanagement.NewSpec(logStoreSpec.RetentionPolicy)

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
				NodeSelector: logStoreSpec.NodeSelector,
				Tolerations:  logStoreSpec.Tolerations,
			},
			Nodes:            esNodes,
			ManagementState:  elasticsearch.ManagementStateManaged,
			RedundancyPolicy: redundancyPolicy,
			IndexManagement:  indexManagementSpec,
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

	esCR := newElasticsearchCR(clusterRequest.cluster, elasticsearchResourceName)

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

	if !utils.AreMapsSame(current.Spec.Spec.NodeSelector, desired.Spec.Spec.NodeSelector) {
		logrus.Infof("Elasticsearch nodeSelector change found, updating '%s'", current.Name)
		current.Spec.Spec.NodeSelector = desired.Spec.Spec.NodeSelector
		different = true
	}

	if !utils.AreTolerationsSame(current.Spec.Spec.Tolerations, desired.Spec.Spec.Tolerations) {
		logrus.Infof("Elasticsearch tolerations change found, updating '%s'", current.Name)
		current.Spec.Spec.Tolerations = desired.Spec.Spec.Tolerations
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

	if !reflect.DeepEqual(current.Spec.IndexManagement, desired.Spec.IndexManagement) {
		logger.Infof("Elasticsearch IndexManagement change found, updating %v", current.Name)
		current.Spec.IndexManagement = desired.Spec.IndexManagement
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

	foundRoleMatch := false
	for nodeIndex := 0; nodeIndex < len(desired); nodeIndex++ {
		for _, node := range current {
			if areNodeRolesSame(node, desired[nodeIndex]) {
				updatedNode, isDifferent := isNodeDifferent(node, desired[nodeIndex])
				if isDifferent {
					desired[nodeIndex] = updatedNode
					different = true
				} else {
					// ensure that we are setting the GenUUID if it existed
					if desired[nodeIndex].GenUUID == nil {
						desired[nodeIndex].GenUUID = updatedNode.GenUUID
					}
				}
				foundRoleMatch = true
			}
		}
	}

	// if we didn't find a role match, then that means changes were made
	if !foundRoleMatch {
		different = true
	}

	// we don't use this to shortcut because the above loop will help to preserve
	// any generated UUIDs
	if len(current) != len(desired) {
		return desired, true
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
