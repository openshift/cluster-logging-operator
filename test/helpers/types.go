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
	"encoding/json"
	"strings"
)

type logs []log

type docker struct {
	ContainerID string `json:"container_id"`
}

type k8s struct {
	ContainerName    string            `json:"container_name"`
	ContainerImage   string            `json:"container_image"`
	ContainerImageID string            `json:"container_image_id"`
	PodName          string            `json:"pod_name"`
	PodID            string            `json:"pod_id"`
	Host             string            `json:"host"`
	Labels           map[string]string `json:"labels"`
	FlatLabels       []string          `json:"flat_labels"`
	MasterURL        string            `json:"master_url"`
	NamespaceName    string            `json:"namespace_name"`
	NamespaceID      string            `json:"namespace_id"`
}

type pipelineMetadata struct {
	Collector *struct {
		IPaddr4    string `json:"ipaddr4"`
		IPaddr6    string `json:"ipaddr6"`
		InputName  string `json:"inputname"`
		Name       string `json:"name"`
		ReceivedAt string `json:"received_at"`
		Version    string `json:"version"`
	} `json:"collector"`
}

type log struct {
	Docker           *docker           `json:"docker"`
	Kubernetes       *k8s              `json:"kubernetes"`
	Message          string            `json:"message"`
	Level            string            `json:"level"`
	Hostname         string            `json:"hostname"`
	PipelineMetadata *pipelineMetadata `json:"pipeline_metadata"`
	Timestamp        string            `json:"@timestamp"`
	IndexName        string            `json:"viaq_index_name"`
	MessageID        string            `json:"viaq_msg_id"`
	OpenshiftLabels  string            `json:"openshift"`
}

func ParseLogs(in string) (logs, error) {
	logs := []log{}
	if in == "" {
		return logs, nil
	}

	err := json.Unmarshal([]byte(in), &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (l logs) ByIndex(prefix string) logs {
	filtered := []log{}
	for _, entry := range l {
		if strings.HasPrefix(entry.IndexName, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (l logs) ByOpenshiftLabel(label string) logs {
	filtered := []log{}
	for _, entry := range l {
		if len(entry.OpenshiftLabels) == 0 {
			continue
		}
		if strings.Contains(entry.OpenshiftLabels, label) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (l logs) ByPod(name string) logs {
	filtered := []log{}
	for _, entry := range l {
		if entry.Kubernetes == nil {
			continue
		}
		if entry.Kubernetes.PodName == name {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (l logs) NonEmpty() bool {
	if l == nil {
		return false
	}
	return len(l) > 0
}
