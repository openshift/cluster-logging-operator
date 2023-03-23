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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clolog "github.com/ViaQ/logerr/v2/log/static"
)

const (
	InfraIndexPrefix   = "infra-"
	ProjectIndexPrefix = "app-"
	AuditIndexPrefix   = "audit-"
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

// HasInfraStructureLogs returns true if there are any indices that begin with InfraIndexPrefix and also contains documents
func (indices *Indices) HasInfraStructureLogs() bool {
	for _, index := range *indices {
		if strings.HasPrefix(index.Name, InfraIndexPrefix) && index.DocCount() > 0 {
			return true
		}
	}
	return false
}

// HasApplicationLogs returns true if there are any indices that begin with ProjectIndexPrefix and also contains documents
func (indices *Indices) HasApplicationLogs() bool {
	for _, index := range *indices {
		if strings.HasPrefix(index.Name, ProjectIndexPrefix) && index.DocCount() > 0 {
			return true
		}
	}
	return false
}

// HasAuditLogs returns true if there are any indices that begin with AuditIndexPrefix and also contains documents
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
			clolog.V(2).Info("Error retrieving indices from elasticsearch")
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
			clolog.Error(err, "Error retrieving indices from elasticsearch")
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
			clolog.Error(err, "Error retrieving indices from elasticsearch")
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

// Indices fetches the list of indices stored by Elasticsearch
func (es *ElasticLogStore) Indices() (Indices, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=elasticsearch",
	}

	pods, err := es.Framework.KubeClient.CoreV1().Pods(constants.WatchNamespace).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, errors.New("No pods found for elasticsearch")
	}
	clolog.V(3).Info("Pod ", "PodName", pods.Items[0].Name)
	indices := []Index{}
	stdout, err := es.Framework.PodExec(constants.WatchNamespace, pods.Items[0].Name, "elasticsearch", []string{"es_util", "--query=_cat/indices?format=json"})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(stdout), &indices)
	if err != nil {
		return nil, err
	}
	return indices, nil
}
