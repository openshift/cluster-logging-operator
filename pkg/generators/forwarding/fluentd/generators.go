package fluentd

import (
	"fmt"
	"sort"
	"text/template"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/util/sets"
)

//ConfigGenerator is a config generator for fluentd
type ConfigGenerator struct {
	*generators.Generator
	includeLegacyForwardConfig    bool
	includeLegacySyslogConfig     bool
	useOldRemoteSyslogPlugin      bool
	storeTemplate, outputTemplate string
}

//NewConfigGenerator creates an instance of FluentdConfigGenerator
func NewConfigGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin bool) (*ConfigGenerator, error) {
	engine, err := generators.New("OutputLabelConf",
		&template.FuncMap{
			"labelName":           labelName,
			"sourceTypelabelName": sourceTypeLabelName,
		},
		templateRegistry...)
	if err != nil {
		return nil, err
	}
	return &ConfigGenerator{
			Generator:                  engine,
			includeLegacyForwardConfig: includeLegacyForwardConfig,
			includeLegacySyslogConfig:  includeLegacySyslogConfig,
			useOldRemoteSyslogPlugin:   useOldRemoteSyslogPlugin,
		},
		nil
}

//Generate the fluent.conf file using the forwarding information
func (engine *ConfigGenerator) Generate(forwarder *logging.ClusterLogForwarderSpec) (string, error) {
	logger.DebugObject("Generating fluent.conf using %s", forwarder)
	//sanitize here
	var sourceInputLabels []string
	var sourceToPipelineLabels []string
	var pipelineToOutputLabels []string
	var outputLabels []string
	var err error

	inputs, namespaces := gatherSources(forwarder)

	if sourceInputLabels, err = engine.generateSource(inputs, namespaces); err != nil {
		logger.Tracef("Error generating source blocks: %v", err)
		return "", err
	}
	if sourceToPipelineLabels, err = engine.generateSourceToPipelineLabels(inputsToPipelines(forwarder.Pipelines)); err != nil {
		logger.Tracef("Error generating source to pipeline blocks: %v", err)
		return "", err
	}
	sort.Strings(sourceToPipelineLabels)
	if pipelineToOutputLabels, err = engine.generatePipelineToOutputLabels(forwarder.Pipelines); err != nil {
		logger.Tracef("Error generating pipeline to output labels blocks: %v", err)
		return "", err
	}
	if outputLabels, err = engine.generateOutputLabelBlocks(forwarder.Outputs); err != nil {
		logger.Tracef("Error generating to output label blocks: %v", err)
		return "", err
	}

	data := struct {
		IncludeLegacySecureForward bool
		IncludeLegacySyslog        bool
		CollectInfraLogs           bool
		CollectAppLogs             bool
		CollectAuditLogs           bool
		SourceInputLabels          []string
		SourceToPipelineLabels     []string
		PipelinesToOutputLabels    []string
		OutputLabels               []string
	}{
		engine.includeLegacyForwardConfig,
		engine.includeLegacySyslogConfig,
		inputs.Has(string(logging.InputNameInfrastructure)),
		inputs.Has(string(logging.InputNameApplication)),
		inputs.Has(string(logging.InputNameAudit)),
		sourceInputLabels,
		sourceToPipelineLabels,
		pipelineToOutputLabels,
		outputLabels,
	}
	result, err := engine.Execute("fluentConf", data)
	if err != nil {
		logger.Tracef("Error generating fluentConf")
		return "", fmt.Errorf("Error processing fluentConf template: %v", err)
	}
	logger.Tracef("Successfully generated fluent.conf: %v", result)
	return result, nil
}

func gatherSources(forwarder *logging.ClusterLogForwarderSpec) (types sets.String, namespaces sets.String) {
	types, namespaces = sets.NewString(), sets.NewString()
	specs := forwarder.InputMap()
	for inputName := range logging.NewRoutes(forwarder.Pipelines).ByInput {
		if logging.ReservedInputNames.Has(inputName) {
			types.Insert(inputName) // Use name as type.
		} else if spec, ok := specs[inputName]; ok {
			if app := spec.Application; app != nil {
				types.Insert(logging.InputNameApplication)
				namespaces.Insert(app.Namespaces...)
			}
			if spec.Infrastructure != nil {
				types.Insert(logging.InputNameInfrastructure)
			}
			if spec.Audit != nil {
				types.Insert(logging.InputNameAudit)
			}
		}
	}
	return types, namespaces
}

func inputsToPipelines(pipelines []logging.PipelineSpec) logging.RouteMap {
	result := logging.RouteMap{}
	for _, pipeline := range pipelines {
		for _, inRef := range pipeline.InputRefs {
			result.Insert(inRef, pipeline.Name)
		}
	}
	return result
}

//generateSourceToPipelineLabels generates fluentd label stanzas to fan source types to multiple pipelines
func (engine *ConfigGenerator) generateSourceToPipelineLabels(sourcesToPipelines logging.RouteMap) ([]string, error) {
	configs := []string{}
	for sourceType, pipelineNames := range sourcesToPipelines {
		data := struct {
			IncludeLegacySecureForward bool
			IncludeLegacySyslog        bool
			Source                     string
			PipelineNames              []string
		}{
			engine.includeLegacyForwardConfig,
			engine.includeLegacySyslogConfig,
			string(sourceType),
			pipelineNames.List(),
		}
		result, err := engine.Execute("sourceToPipelineCopyTemplate", data)
		if err != nil {
			return nil, fmt.Errorf("Error processing sourceToPipelineCopyTemplate template: %v", err)
		}
		configs = append(configs, result)
	}
	return configs, nil
}

func (engine *ConfigGenerator) generatePipelineToOutputLabels(pipelines []logging.PipelineSpec) ([]string, error) {
	configs := []string{}
	for _, pipeline := range pipelines {
		data := struct {
			Name    string
			Outputs []string
		}{
			pipeline.Name,
			pipeline.OutputRefs,
		}
		result, err := engine.Execute("pipelineToOutputCopyTemplate", data)
		if err != nil {
			return nil, fmt.Errorf("Error processing pipelineToOutputCopyTemplate template: %v", err)
		}
		configs = append(configs, result)
	}
	return configs, nil
}

//generateStoreLabelBlocks generates fluentd label stanzas for sources to specific store destinations
// <label @ELASTICSEARCH_OFFCLUSTER>
//  <match retry_elasticsearch_offcluster>
//  @type copy
//  <store>
//		@type elasticsearch
//  </store>
//  </match>
//  <match **>
//    @type copy
//  </match>
// </label>
func (engine *ConfigGenerator) generateOutputLabelBlocks(outputs []logging.OutputSpec) ([]string, error) {
	configs := []string{}
	logger.Debugf("Evaluating %v forwarder logging...", len(outputs))
	for _, output := range outputs {
		logger.Debugf("Generate output type %v", output.Type)
		engine.outputTemplate = "outputLabelConf" // Default
		switch output.Type {
		case logging.OutputTypeElasticsearch:
			engine.storeTemplate = "storeElasticsearch"
		case logging.OutputTypeFluentForward:
			engine.storeTemplate, engine.outputTemplate = "forward", "outputLabelConfNoCopy"
		case logging.OutputTypeSyslog:
			if engine.useOldRemoteSyslogPlugin {
				engine.storeTemplate = "storeSyslogOld"
			} else {
				engine.storeTemplate = "storeSyslog"
			}
			engine.outputTemplate = "outputLabelConfNoRetry"
		default:
			return nil, fmt.Errorf("Unknown outpt type: %v", output.Type)
		}
		conf := newOutputLabelConf(engine.Template, engine.storeTemplate, output)
		result, err := engine.Execute(engine.outputTemplate, conf)
		if err != nil {
			return nil, fmt.Errorf("Error generating fluentd config Processing template outputLabelConf: %v", err)
		}
		configs = append(configs, result)
	}
	logger.Debugf("Generated output configurations: %v", configs)
	return configs, nil
}
