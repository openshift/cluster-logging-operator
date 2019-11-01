package fluentd

import (
	"fmt"

	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (engine *ConfigGenerator) generateSource(sources sets.String) (results []string, err error) {
	//looking to control order
	templates := []string{}
	if sources.Has(string(logforward.LogSourceTypeInfra)) {
		templates = append(templates, "inputSourceJournalTemplate")
	}
	if sources.Has(string(logforward.LogSourceTypeApp)) {
		templates = append(templates, "inputSourceContainerTemplate")
	}
	if sources.Has(string(logforward.LogSourceTypeAudit)) {
		templates = append(templates, "inputSourceHostAuditTemplate")
		templates = append(templates, "inputSourceK8sAuditTemplate")
		templates = append(templates, "inputSourceOpenShiftAuditTemplate")
	}
	if len(templates) == 0 {
		return results, fmt.Errorf("Unable to generate source configs for supported source types: %v", sources.List())
	}
	for _, template := range templates {
		result, err := engine.Execute(template, "")
		if err != nil {
			return results, fmt.Errorf("Error processing template %s: %v", template, err)
		}
		results = append(results, result)
	}
	return results, nil
}
