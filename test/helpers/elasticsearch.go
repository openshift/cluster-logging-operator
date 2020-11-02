// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clolog "github.com/ViaQ/logerr/log"
	k8shandler "github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler/indexmanagement"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
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

func (es *ElasticLogStore) ApplicationLogs(timeToWait time.Duration) (logs, error) {
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

//Indices fetches the list of indices stored by Elasticsearch
func (es *ElasticLogStore) Indices() (Indices, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=elasticsearch",
	}

	pods, err := es.Framework.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, errors.New("No pods found for elasticsearch")
	}
	clolog.V(3).Info("Pod ", "PodName", pods.Items[0].Name)
	indices := []Index{}
	stdout, err := es.Framework.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, "elasticsearch", []string{"es_util", "--query=_cat/indices?format=json"})
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
	clolog.V(3).Info("DeployAnElasticsearchCluster")
	logStoreName := "test-elastic-cluster"
	if pipelineSecret, err = tc.CreatePipelineSecret(pwd, logStoreName, "test-pipeline-to-elastic", map[string][]byte{}); err != nil {
		return nil, nil, err
	}

	opts := metav1.CreateOptions{}
	esSecret := k8shandler.NewSecret(
		logStoreName,
		OpenshiftLoggingNS,
		k8shandler.LoadElasticsearchSecretMap(),
	)
	clolog.V(3).Info("Creating secret for an elasticsearch cluster: ", "secret", esSecret.Name)
	if esSecret, err = tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Create(context.TODO(), esSecret, opts); err != nil {
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
			if err := tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Delete(context.TODO(), name, opts); err != nil {
				return err
			}
		}
		return nil
	})

	clolog.V(3).Info("Creating an elasticsearch cluster ", "cluster", cr)
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
