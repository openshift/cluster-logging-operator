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

package k8shandler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

func NewPrometheusRule(ruleName, namespace string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ruleName,
			Namespace: namespace,
		},
	}
}

func (clusterRequest *ClusterLoggingRequest) RemovePrometheusRule(ruleName string) error {

	promRule := NewPrometheusRule(ruleName, clusterRequest.Cluster.Namespace)

	err := clusterRequest.Delete(promRule)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v prometheus rule: %v", promRule, err)
	}

	return nil
}

func NewPrometheusRuleSpecFrom(filePath string) (*monitoringv1.PrometheusRuleSpec, error) {
	if err := utils.CheckFileExists(filePath); err != nil {
		return nil, err
	}
	fileContent, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("'%s' not readable", filePath)
	}
	ruleSpec := monitoringv1.PrometheusRuleSpec{}
	if err := k8sYAML.NewYAMLOrJSONDecoder(bytes.NewBufferString(string(fileContent)), 1000).Decode(&ruleSpec); err != nil {
		return nil, err
	}
	return &ruleSpec, nil
}
