package fluentd

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/util/sets"
)

// FIXME(alanconway) generateSource scrapes files for all requested namespaces.
// We need to also filter them per-user-SourceSpec since different SourceSpecs
// might request different namespaces.

func (engine *ConfigGenerator) generateSource(sources sets.String, appNs sets.String) (results []string, err error) {
	// Order of templates matters.
	templates := []string{}
	nsPaths := []string{}
	if sources.Has(string(logging.InputNameInfrastructure)) {
		templates = append(templates, "inputSourceJournalTemplate")
	}
	if sources.Has(string(logging.InputNameApplication)) {
		templates = append(templates, "inputSourceContainerTemplate")
		for _, ns := range appNs.List() {
			nsPaths = append(nsPaths, fmt.Sprintf("\"/var/log/containers/*_%s_*.log\"", ns))
		}
	}
	if sources.Has(string(logging.InputNameAudit)) {
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
		AppNsPaths                 string
	}{
		constants.OpenshiftNS,
		constants.FluentdName,
		constants.ElasticsearchName,
		constants.KibanaName,
		strings.Join(nsPaths, ", "),
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
