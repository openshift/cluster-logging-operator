package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clolog "github.com/ViaQ/logerr/v2/log"
	k8shandler "github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler/indexmanagement"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
)

const (
	InfraIndexPrefix          = "infra-"
	ProjectIndexPrefix        = "app-"
	AuditIndexPrefix          = "audit-"
	elasticsearchesLoggingURI = "apis/logging.openshift.io/v1/namespaces/openshift-logging/elasticsearches"
)

type Indices []Index

type Index struct {
	Health           string `json:"health"`
	Status           string `json:"status"`
	Name             string `json:"index"`
	UUID             string `json:"uuid"`
	Primary          string `json:"pri"`
	Replicas         string `json:"rep"`
	DocsCount        string `json:"docs.count"`
	DocsDeleted      string `json:"docs.deleted"`
	StoreSize        string `json:"store.size"`
	PrimaryStoreSize string `json:"pri.store.size"`
}

func (index *Index) DocCount() int {
	if index.DocsCount == "" {
		return 0
	}
	value, err := strconv.Atoi(index.DocsCount)
	if err != nil {
		panic(err)
	}
	return value
}

//HasInfraStructureLogs returns true if there are any indices that begin with InfraIndexPrefix and also contains documents
func (indices *Indices) HasInfraStructureLogs() bool {
	for _, index := range *indices {
		if strings.HasPrefix(index.Name, InfraIndexPrefix) && index.DocCount() > 0 {
			return true
		}
	}
	return false
}

//HasApplicationLogs returns true if there are any indices that begin with ProjectIndexPrefix and also contains documents
func (indices *Indices) HasApplicationLogs() bool {
	for _, index := range *indices {
		if strings.HasPrefix(index.Name, ProjectIndexPrefix) && index.DocCount() > 0 {
			return true
		}
	}
	return false
}

//HasAuditLogs returns true if there are any indices that begin with AuditIndexPrefix and also contains documents
func (indices *Indices) HasAuditLogs() bool {
	for _, index := range *indices {
		if strings.HasPrefix(index.Name, AuditIndexPrefix) && index.DocCount() > 0 {
			return true
		}
	}
	return false
}

type ElasticLogStore struct {
	Framework *E2ETestFramework
}

func (es *ElasticLogStore) ApplicationLogs(timeToWait time.Duration) (types.Logs, error) {
	panic("Method not implemented")
}

func (es *ElasticLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	err := wait.PollImmediate(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		indices, err := es.Indices()
		if err != nil {
			//accept arbitrary errors like 'etcd leader change'
			clolog.NewLogger("e2e-elasticsearch-testing").V(2).Info("Error retrieving indices from elasticsearch")
			return false, nil
		}
		return indices.HasInfraStructureLogs(), nil
	})
	return true, err
}

func (es *ElasticLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	err := wait.PollImmediate(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		indices, err := es.Indices()
		if err != nil {
			//accept arbitrary errors like 'etcd leader change'
			clolog.NewLogger("e2e-elasticsearch-testing").Error(err, "Error retrieving indices from elasticsearch")
			return false, nil
		}
		return indices.HasApplicationLogs(), nil
	})
	return true, err
}

func (es *ElasticLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	err := wait.PollImmediate(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		indices, err := es.Indices()
		if err != nil {
			//accept arbitrary errors like 'etcd leader change'
			clolog.NewLogger("e2e-elasticsearch-testing").Error(err, "Error retrieving indices from elasticsearch")
			return false, nil
		}
		return indices.HasAuditLogs(), nil
	})
	return true, err
}

func (es *ElasticLogStore) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return "Not Found", fmt.Errorf("Not implemented")
}

func (es *ElasticLogStore) RetrieveLogs() (map[string]string, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (es *ElasticLogStore) ClusterLocalEndpoint() string {
	panic("Not implemented")
}

//Indices fetches the list of indices stored by Elasticsearch
func (es *ElasticLogStore) Indices() (Indices, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=elasticsearch",
	}

	pods, err := es.Framework.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, errors.New("No pods found for elasticsearch")
	}
	clolog.NewLogger("e2e-elasticsearch-testing").V(3).Info("Pod ", "PodName", pods.Items[0].Name)
	indices := []Index{}
	stdout, err := es.Framework.PodExec(constants.OpenshiftNS, pods.Items[0].Name, "elasticsearch", []string{"es_util", "--query=_cat/indices?format=json"})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(stdout), &indices)
	if err != nil {
		return nil, err
	}
	return indices, nil
}

func (tc *E2ETestFramework) DeployAnElasticsearchCluster(pwd string) (cr *elasticsearch.Elasticsearch, pipelineSecret *corev1.Secret, err error) {
	logStoreName := "test-elastic-cluster"
	if pipelineSecret, err = tc.CreatePipelineSecret(pwd, logStoreName, "test-pipeline-to-elastic", map[string][]byte{}); err != nil {
		return nil, nil, err
	}

	opts := metav1.CreateOptions{}
	esSecret := k8shandler.NewSecret(
		logStoreName,
		constants.OpenshiftNS,
		k8shandler.LoadElasticsearchSecretMap(),
	)
	logger := clolog.NewLogger("e2e-elasticsearch-testing")
	logger.V(3).Info("Creating secret for an elasticsearch cluster: ", "secret", esSecret.Name)
	_, err = tc.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Create(context.TODO(), esSecret, opts)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			sOpts := metav1.UpdateOptions{}
			_, err := tc.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Update(context.TODO(), esSecret, sOpts)
			if err != nil {
				return nil, nil, nil
			}
		} else {
			return nil, nil, nil
		}
	}
	pvcSize := resource.MustParse("200G")
	node := elasticsearch.ElasticsearchNode{
		Roles:     []elasticsearch.ElasticsearchNodeRole{"client", "data", "master"},
		NodeCount: 1,
		Storage: elasticsearch.ElasticsearchStorageSpec{
			Size: &pvcSize,
		},
	}
	cr = &elasticsearch.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      esSecret.Name,
			Namespace: constants.OpenshiftNS,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.GroupVersion.String(),
		},
		Spec: elasticsearch.ElasticsearchSpec{
			Spec: elasticsearch.ElasticsearchNodeSpec{
				Image: utils.GetComponentImage("elasticsearch"),
				Resources: corev1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("3Gi"),
					},
				},
				ProxyResources: corev1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("128Mi"),
					},
				},
			},
			Nodes:            []elasticsearch.ElasticsearchNode{node},
			ManagementState:  elasticsearch.ManagementStateManaged,
			RedundancyPolicy: elasticsearch.ZeroRedundancy,
			IndexManagement:  indexmanagement.NewSpec(nil),
		},
	}
	tc.AddCleanup(func() error {
		result := tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/%s", elasticsearchesLoggingURI, cr.Name)).
			SetHeader("Content-Type", "application/json").
			Do(context.TODO())
		return result.Error()
	})
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		for _, name := range []string{esSecret.Name, pipelineSecret.Name} {
			if err := tc.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Delete(context.TODO(), name, opts); err != nil {
				return err
			}
		}
		return nil
	})

	logger.V(3).Info("Creating an elasticsearch cluster ", "cluster", cr)
	var body []byte
	if body, err = json.Marshal(cr); err != nil {
		return nil, nil, err
	}
	result := tc.KubeClient.RESTClient().Post().
		RequestURI(elasticsearchesLoggingURI).
		SetHeader("Content-Type", "application/json").
		Body(body).
		Do(context.TODO())

	name := cr.GetName()
	tc.LogStores[name] = &ElasticLogStore{
		Framework: tc,
	}
	return cr, pipelineSecret, result.Error()
}
