package k8shandler

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler/indexmanagement"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	elasticsearchResourceName       = "elasticsearch"
	maximumElasticsearchMasterCount = int32(3)
)

func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateLogStore() (err error) {
	if clusterRequest.Cluster.Spec.LogStore == nil || clusterRequest.Cluster.Spec.LogStore.Type == "" {
		return nil
	}
	if clusterRequest.Cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {

		cluster := clusterRequest.Cluster

		if err = clusterRequest.removeElasticsearchIfSecretOwnedByCLO(); err != nil {
			log.NewLogger("").Error(err, "Can't fully clean up old secret created by CLO")
			return err
		}

		if err = clusterRequest.createOrUpdateElasticsearchCR(); err != nil {
			return err
		}

		elasticsearchStatus, err := clusterRequest.getElasticsearchStatus()

		if err != nil {
			return fmt.Errorf("Failed to get Elasticsearch status for %q: %v", cluster.Name, err)
		}

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if !compareElasticsearchStatus(elasticsearchStatus, cluster.Status.LogStore.ElasticsearchStatus) {
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

// need for smooth upgrade CLO to the 5.4 version, after moving certificates generation to the EO side
// see details: https://issues.redhat.com/browse/LOG-1923
func (clusterRequest *ClusterLoggingRequest) removeElasticsearchIfSecretOwnedByCLO() (err error) {
	secret, err := clusterRequest.GetSecret(constants.ElasticsearchName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if utils.IsOwnedBy(secret.GetOwnerReferences(), utils.AsOwner(clusterRequest.Cluster)) {
		if err = clusterRequest.removeElasticsearch(); err != nil {
			return err
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
		for index := range lhs {
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
	var results = map[string][]byte{}
	_ = Syncronize(func() error {
		results = map[string][]byte{
			"elasticsearch.key": utils.GetWorkingDirFileContents("elasticsearch.key"),
			"elasticsearch.crt": utils.GetWorkingDirFileContents("elasticsearch.crt"),
			"logging-es.key":    utils.GetWorkingDirFileContents("logging-es.key"),
			"logging-es.crt":    utils.GetWorkingDirFileContents("logging-es.crt"),
			"admin-key":         utils.GetWorkingDirFileContents("system.admin.key"),
			"admin-cert":        utils.GetWorkingDirFileContents("system.admin.crt"),
			"admin-ca":          utils.GetWorkingDirFileContents("ca.crt"),
		}
		return nil
	})
	return results
}

var mutex sync.Mutex

//Syncronize blocks single threads access using the certificate mutex
func Syncronize(action func() error) error {
	mutex.Lock()
	defer mutex.Unlock()
	return action()
}

func (cr *ClusterLoggingRequest) emptyElasticsearchCR(elasticsearchName string) *elasticsearch.Elasticsearch {
	return &elasticsearch.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      elasticsearchName,
			Namespace: cr.Cluster.Namespace,
			Annotations: map[string]string{
				"logging.openshift.io/elasticsearch-cert-management": "true",
				"logging.openshift.io/elasticsearch-cert.collector":  "system.logging.fluentd",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.GroupVersion.String(),
		},
		Spec: elasticsearch.ElasticsearchSpec{},
	}
}

func (cr *ClusterLoggingRequest) newElasticsearchCR(elasticsearchName string, existing *elasticsearch.Elasticsearch) *elasticsearch.Elasticsearch {

	var esNodes []elasticsearch.ElasticsearchNode
	logStoreSpec := logging.LogStoreSpec{}
	if cr.Cluster.Spec.LogStore != nil {
		logStoreSpec = *cr.Cluster.Spec.LogStore
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
	var proxyResources = logStoreSpec.ProxySpec.Resources
	if proxyResources == nil {
		proxyResources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceMemory: defaultEsProxyMemory,
			},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultEsProxyMemory,
				v1.ResourceCPU:    defaultEsProxyCpuRequest,
			},
		}
	}

	esNode := elasticsearch.ElasticsearchNode{
		Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
		NodeCount: logStoreSpec.NodeCount,
		Storage:   logStoreSpec.ElasticsearchSpec.Storage,
	}

	// build Nodes
	esNodes = append(esNodes, esNode)

	// if we had more than 1 es node before, we also want to enter this condition
	if logStoreSpec.NodeCount > maximumElasticsearchMasterCount || len(existing.Spec.Nodes) > 1 {

		// we need to check this because if we scaled down we can enter this block
		if logStoreSpec.NodeCount > maximumElasticsearchMasterCount {
			esNodes[0].NodeCount = maximumElasticsearchMasterCount
		}

		remainder := logStoreSpec.NodeCount - maximumElasticsearchMasterCount
		if remainder < 0 {
			remainder = 0
		}

		dataNode := elasticsearch.ElasticsearchNode{
			Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data"},
			NodeCount: remainder,
			Storage:   logStoreSpec.ElasticsearchSpec.Storage,
		}

		esNodes = append(esNodes, dataNode)

	}

	redundancyPolicy := logStoreSpec.ElasticsearchSpec.RedundancyPolicy
	if redundancyPolicy == "" {
		redundancyPolicy = elasticsearch.ZeroRedundancy
	}

	indexManagementSpec := indexmanagement.NewSpec(logStoreSpec.RetentionPolicy)

	es := cr.emptyElasticsearchCR(elasticsearchName)
	es.Spec = elasticsearch.ElasticsearchSpec{
		Spec: elasticsearch.ElasticsearchNodeSpec{
			Resources:      *resources,
			NodeSelector:   logStoreSpec.NodeSelector,
			Tolerations:    logStoreSpec.Tolerations,
			ProxyResources: *proxyResources,
		},
		Nodes:            esNodes,
		ManagementState:  elasticsearch.ManagementStateManaged,
		RedundancyPolicy: redundancyPolicy,
		IndexManagement:  indexManagementSpec,
	}

	utils.AddOwnerRefToObject(es, utils.AsOwner(cr.Cluster))

	return es
}

func (clusterRequest *ClusterLoggingRequest) removeElasticsearchCR(elasticsearchName string) error {

	esCr := clusterRequest.emptyElasticsearchCR(elasticsearchName)

	err := clusterRequest.Delete(esCr)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v elasticsearch CR for %q: %v", elasticsearchName, clusterRequest.Cluster.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateElasticsearchCR() (err error) {

	if !clusterRequest.isManaged() {
		return nil
	}

	// get existing CR first
	existingCR := &elasticsearch.Elasticsearch{}
	if err = clusterRequest.Get(elasticsearchResourceName, existingCR); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("Failed to get Elasticsearch CR: %v", err)
		}
	}

	esCR := clusterRequest.newElasticsearchCR(elasticsearchResourceName, existingCR)

	err = clusterRequest.Create(esCR)
	if err == nil {
		return nil
	}

	if !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Elasticsearch CR: %v", err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return clusterRequest.updateElasticsearchCRIfRequired(esCR)
	})

	return retryErr
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

	logger := log.NewLogger("logstore")

	if !utils.AreMapsSame(current.Spec.Spec.NodeSelector, desired.Spec.Spec.NodeSelector) {
		logger.Info("Elasticsearch nodeSelector change found, updating", "currentName", current.Name)
		current.Spec.Spec.NodeSelector = desired.Spec.Spec.NodeSelector
		different = true
	}

	if !utils.AreTolerationsSame(current.Spec.Spec.Tolerations, desired.Spec.Spec.Tolerations) {
		logger.Info("Elasticsearch tolerations change found, updating", "currentName", current.Name)
		current.Spec.Spec.Tolerations = desired.Spec.Spec.Tolerations
		different = true
	}

	if current.Spec.Spec.Image != desired.Spec.Spec.Image {
		logger.Info("Elasticsearch image change found, updating", "currentName", current.Name)
		current.Spec.Spec.Image = desired.Spec.Spec.Image
		different = true
	}

	if current.Spec.RedundancyPolicy != desired.Spec.RedundancyPolicy {
		logger.Info("Elasticsearch redundancy policy change found, updating", "currentName", current.Name)
		current.Spec.RedundancyPolicy = desired.Spec.RedundancyPolicy
		different = true
	}

	if !reflect.DeepEqual(current.ObjectMeta.Annotations, desired.ObjectMeta.Annotations) {
		logger.Info("Elasticsearch resources change found in Annotations, updating", "currentName", current.Name)
		current.Annotations = desired.Annotations
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Spec.Resources, desired.Spec.Spec.Resources) {
		logger.Info("Elasticsearch resources change found, updating", "currentName", current.Name)
		current.Spec.Spec.Resources = desired.Spec.Spec.Resources
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Spec.ProxyResources, desired.Spec.Spec.ProxyResources) {
		logger.Info("Elasticsearch Proxy resources change found, updating", "currentName", current.Name)
		current.Spec.Spec.ProxyResources = desired.Spec.Spec.ProxyResources
		different = true
	}

	if nodes, ok := areNodesDifferent(current.Spec.Nodes, desired.Spec.Nodes); ok {
		logger.Info("Elasticsearch node configuration change found, updating", "currentName", current.Name)
		current.Spec.Nodes = nodes
		different = true
	}

	if !reflect.DeepEqual(current.Spec.IndexManagement, desired.Spec.IndexManagement) {
		logger.Info("Elasticsearch IndexManagement change found, updating", "currentName", current.Name)
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
				} else if desired[nodeIndex].GenUUID == nil {
					// ensure that we are setting the GenUUID if it existed
					desired[nodeIndex].GenUUID = updatedNode.GenUUID
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
