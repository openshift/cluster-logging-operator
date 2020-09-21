package fluentbit

import (
	"fmt"
	"text/template"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/sources"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/util/sets"
)

//ConfigGenerator is a config generator for fluentd
type ConfigGenerator struct {
	*generators.Generator
}

//NewConfigGenerator creates an instance of Fluentbit ConfigGenerator
func NewConfigGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin bool) (*ConfigGenerator, error) {
	engine, err := generators.New("fluentbitConfTemplate",
		&template.FuncMap{},
		templateRegistry...)
	if err != nil {
		return nil, err
	}
	generator := &ConfigGenerator{
		Generator: engine,
	}
	return generator, nil
}

func (engine *ConfigGenerator) Generate(clfSpec *logging.ClusterLogForwarderSpec, fwSpec *logging.ForwarderSpec) (string, error) {
	logger.DebugObject("Generating fluent-bit.conf using %s", clfSpec)
	logger.DebugObject("Ignoring buffer config: %s", fwSpec)

	//sanitize here
	var (
		inputs      sets.String
		sourcesConf []string
		err         error
	)

	inputs, _ = sources.GatherSources(clfSpec)

	sourcesConf, err = engine.generateSource(inputs)
	if err != nil {

		logger.Tracef("Error generating source blocks: %v", err)
		return "", err
	}
	data := struct {
		Sources []string
	}{
		sourcesConf,
	}
	result, err := engine.Execute("fluentbitConfTemplate", data)
	if err != nil {
		logger.Tracef("Error generating fluentBitConf")
		return "", fmt.Errorf("Error processing fluentBitConf template: %v", err)
	}
	logger.Tracef("Successfully generated fluent-bit.conf: %v", result)
	return result, nil
}

func (engine *ConfigGenerator) generateSource(sources sets.String) (results []string, err error) {
	// Order of templates matters.
	var templates []string
	if sources.Has(logging.InputNameInfrastructure) {
		templates = append(templates, "systemdTemplate", "containerInfraTemplate")
	}
	if sources.Has(logging.InputNameApplication) {
		templates = append(templates, "containerAppTemplate")
	}
	if sources.Has(logging.InputNameAudit) {
		templates = append(templates, "auditTemplate")
	}
	if len(templates) == 0 {
		return results, fmt.Errorf("No recognized input types: %v", sources.List())
	}
	data := struct{}{}
	for _, template := range templates {
		result, err := engine.Execute(template, data)
		if err != nil {
			return results, fmt.Errorf("Error processing template %s: %v", template, err)
		}
		results = append(results, result)
	}
	return results, nil
}
