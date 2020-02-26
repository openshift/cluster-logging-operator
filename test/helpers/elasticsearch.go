package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	k8shandler "github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/sirupsen/logrus"
)

const (
	InfraIndexPrefix          = "infra-"
	ProjectIndexPrefix        = "app-"
	AuditIndexPrefix          = "audit-infra-"
	AppWriteIndex             = "app-write"
	elasticsearchesLoggingURI = "apis/logging.openshift.io/v1/namespaces/openshift-logging/elasticsearches"
)

type SearchResult struct {
	Hits *struct {
		Total int `json:"total"`
		Hits  []*struct {
			Index  string `json:"_index"`
			Source *struct {
				Message    string `json:"message"`
				Kubernetes *struct {
					Name      string            `json:"container_name"`
					Namespace string            `json:"namespace_name"`
					Labels    map[string]string `json:"labels"`
				} `json:"kubernetes"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (sr *SearchResult) Total() int {
	if sr.Hits == nil {
		return 0
	}
	return sr.Hits.Total
}

func (sr *SearchResult) HasMessage(s string) (bool, error) {
	logrus.Printf("%#v", sr)

	if sr.Hits == nil {
		return false, errors.New("Empty search result")
	}

	if len(sr.Hits.Hits) == 0 {
		return false, nil
	}

	found := false
	for _, hit := range sr.Hits.Hits {
		if s == hit.Source.Message {
			found = true
			break
		}
	}

	return found, nil
}

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

func (es *ElasticLogStore) HasAppLogEntry(msg string, timeToWait time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		errorCount := 0
		sr, err := es.Search(AppWriteIndex, msg)
		if err != nil {
			logger.Errorf("Error searching %s index on elasticsearch %v", AppWriteIndex, err)
			errorCount++
			if errorCount > 5 { //accept arbitrary errors like 'etcd leader change'
				return false, err
			}
			return false, nil
		}
		return sr.HasMessage(msg)
	})
	return true, err
}

func (es *ElasticLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		errorCount := 0
		indices, err := es.Indices()
		if err != nil {
			logger.Errorf("Error retrieving indices from elasticsearch %v", err)
			errorCount++
			if errorCount > 5 { //accept arbitrary errors like 'etcd leader change'
				return false, err
			}
			return false, nil
		}
		return indices.HasInfraStructureLogs(), nil
	})
	return true, err
}

func (es *ElasticLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		errorCount := 0
		indices, err := es.Indices()
		if err != nil {
			logger.Errorf("Error retrieving indices from elasticsearch %v", err)
			errorCount++
			if errorCount > 5 {
				return false, err
			}
			return false, nil
		}
		return indices.HasApplicationLogs(), nil
	})
	return true, err
}

func (es *ElasticLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		indices, err := es.Indices()
		if err != nil {
			return false, err
		}
		return indices.HasAuditLogs(), nil
	})
	return true, err
}

//Indices fetches the list of indices stored by Elasticsearch
func (es *ElasticLogStore) Indices() (Indices, error) {
	indices := []Index{}
	stdout, err := es.esUtil("--query=_cat/indices?format=json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(stdout), &indices)
	if err != nil {
		return nil, err
	}
	return indices, nil
}

func (es *ElasticLogStore) Search(index, s string) (*SearchResult, error) {
	results := SearchResult{}
	stdout, err := es.esUtil(fmt.Sprintf("--query=%s/_search/?q=%s", index, url.QueryEscape(s)))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(stdout), &results)
	if err != nil {
		return nil, err
	}
	return &results, nil
}

func (tc *E2ETestFramework) DeployAnElasticsearchCluster(pwd string) (cr *elasticsearch.Elasticsearch, pipelineSecret *corev1.Secret, err error) {
	logger.Debug("DeployAnElasticsearchCluster")
	logStoreName := "test-elastic-cluster"
	if pipelineSecret, err = tc.CreatePipelineSecret(pwd, logStoreName, "test-pipeline-to-elastic", map[string][]byte{}); err != nil {
		return nil, nil, err
	}

	esSecret := k8shandler.NewSecret(
		logStoreName,
		OpenshiftLoggingNS,
		k8shandler.LoadElasticsearchSecretMap(),
	)
	logger.Debugf("Creating secret for an elasticsearch cluster: %s", esSecret.Name)
	if esSecret, err = tc.KubeClient.Core().Secrets(OpenshiftLoggingNS).Create(esSecret); err != nil {
		return nil, nil, err
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
			Namespace: OpenshiftLoggingNS,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.SchemeGroupVersion.String(),
		},
		Spec: elasticsearch.ElasticsearchSpec{
			Spec: elasticsearch.ElasticsearchNodeSpec{
				Image: utils.GetComponentImage("elasticsearch"),
				Resources: corev1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("3Gi"),
					},
				},
			},
			Nodes:            []elasticsearch.ElasticsearchNode{node},
			ManagementState:  elasticsearch.ManagementStateManaged,
			RedundancyPolicy: elasticsearch.ZeroRedundancy,
		},
	}
	tc.AddCleanup(func() error {
		result := tc.KubeClient.RESTClient().Delete().
			RequestURI(fmt.Sprintf("%s/%s", elasticsearchesLoggingURI, cr.Name)).
			SetHeader("Content-Type", "application/json").
			Do()
		return result.Error()
	})
	tc.AddCleanup(func() error {
		for _, name := range []string{esSecret.Name, pipelineSecret.Name} {
			if err := tc.KubeClient.Core().Secrets(OpenshiftLoggingNS).Delete(name, nil); err != nil {
				return err
			}
		}
		return nil
	})

	logger.Debugf("Creating an elasticsearch cluster %v:", cr)
	var body []byte
	if body, err = json.Marshal(cr); err != nil {
		return nil, nil, err
	}
	result := tc.KubeClient.RESTClient().Post().
		RequestURI(elasticsearchesLoggingURI).
		SetHeader("Content-Type", "application/json").
		Body(body).
		Do()
	tc.LogStore = &ElasticLogStore{
		Framework: tc,
	}
	return cr, pipelineSecret, result.Error()
}

func (es *ElasticLogStore) esUtil(args ...string) (string, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=elasticsearch",
	}
	pods, err := es.Framework.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(options)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", errors.New("No pods found for elasticsearch")
	}
	logger.Debugf("Pod %s", pods.Items[0].Name)

	command := []string{"es_util"}
	command = append(command, args...)

	stdout, err := es.Framework.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, "elasticsearch", command)
	if err != nil {
		return "", err
	}
	return stdout, nil
}
