package fluentd

import (
	"fmt"
	"strconv"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/util/sets"
)

// FIXME(alanconway) generateSource scrapes files for all requested namespaces.
// We need to also filter them per-user-SourceSpec since different SourceSpecs
// might request different namespaces.

func (engine *ConfigGenerator) generateSource(sources sets.String, tunings *logging.FluentdInFileSpec) (results []string, err error) {
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
	data := sourceConfData{
		LoggingNamespace:           constants.OpenshiftNS,
		CollectorPodNamePrefix:     constants.FluentdName,
		LogStorePodNamePrefix:      constants.ElasticsearchName,
		VisualizationPodNamePrefix: constants.KibanaName,
		Tunings:                    tunings,
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

type sourceConfData struct {
	LoggingNamespace           string
	CollectorPodNamePrefix     string
	LogStorePodNamePrefix      string
	VisualizationPodNamePrefix string
	Tunings                    *logging.FluentdInFileSpec
}

func (cl sourceConfData) ReadLinesLimit() string {
	if cl.Tunings == nil || cl.Tunings.ReadLinesLimit <= 0 {
		return ""
	}
	return "\n  read_lines_limit " + strconv.Itoa(cl.Tunings.ReadLinesLimit)
}
