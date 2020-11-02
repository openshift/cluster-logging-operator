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

package fluentd

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/util/sets"
)

// FIXME(alanconway) generateSource scrapes files for all requested namespaces.
// We need to also filter them per-user-SourceSpec since different SourceSpecs
// might request different namespaces.

func (engine *ConfigGenerator) generateSource(sources sets.String) (results []string, err error) {
	// Order of templates matters.
	var templates []string
	if sources.Has(logging.InputNameInfrastructure) {
		templates = append(templates, "inputSourceJournalTemplate")
	}
	if sources.Has(logging.InputNameApplication) || sources.Has(logging.InputNameInfrastructure) {
		templates = append(templates, "inputSourceContainerTemplate")
	}
	if sources.Has(logging.InputNameAudit) {
		templates = append(templates, "inputSourceHostAuditTemplate")
		templates = append(templates, "inputSourceK8sAuditTemplate")
		templates = append(templates, "inputSourceOpenShiftAuditTemplate")
	}
	if len(templates) == 0 {
		return results, fmt.Errorf("No recognized input types: %v", sources.List())
	}
	data := struct {
		LoggingNamespace           string
		CollectorPodNamePrefix     string
		LogStorePodNamePrefix      string
		VisualizationPodNamePrefix string
	}{
		constants.OpenshiftNS,
		constants.FluentdName,
		constants.ElasticsearchName,
		constants.KibanaName,
	}
	for _, template := range templates {
		result, err := engine.Execute(template, data)
		if err != nil {
			return results, fmt.Errorf("Error processing template %s: %v", template, err)
		}
		results = append(results, result)
	}
	return results, nil
}
