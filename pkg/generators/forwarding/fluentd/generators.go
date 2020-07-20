package fluentd

import (
	"fmt"
	"sort"
	"text/template"

	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/generators"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/util/sets"
)

//ConfigGenerator is a config generator for fluentd
type ConfigGenerator struct {
	*generators.Generator
	includeLegacyForwardConfig bool
	includeLegacySyslogConfig  bool
	useOldRemoteSyslogPlugin   bool
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
	return &ConfigGenerator{engine, includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin}, nil
}

//Generate the fluent.conf file using the forwarding information
func (engine *ConfigGenerator) Generate(forwarding *logforward.ForwardingSpec) (string, error) {
	logger.DebugObject("Generating fluent.conf using %s", forwarding)
	//sanitize here
	var (
		logTypes               sets.String
		appNs                  sets.String
		sourceInputLabels      []string
		pipelineNames          map[logforward.LogSourceType][]string
		sourceToPipelineLabels []string
		pipelineToOutputLabels []string
		outputLabels           []string
		err                    error
	)

	// Provide logTypes for legacy forwarding protocols w/o a user-provided
	// LogFowarding instance to enable logTypes and appNs for template generation.
	if engine.includeLegacyForwardConfig || engine.includeLegacySyslogConfig {
		logTypes = sets.NewString(
			string(logforward.LogSourceTypeApp),
			string(logforward.LogSourceTypeAudit),
			string(logforward.LogSourceTypeInfra),
		)
		pipelineNames = map[logforward.LogSourceType][]string{
			logforward.LogSourceTypeApp:   {},
			logforward.LogSourceTypeAudit: {},
			logforward.LogSourceTypeInfra: {},
		}
	} else {
		logTypes, appNs = gatherLogSourceTypes(forwarding.Pipelines)
		pipelineNames = mapSourceTypesToPipelineNames(forwarding.Pipelines)
	}

	sourceInputLabels, err = engine.generateSource(logTypes, appNs)
	if err != nil {
		logger.Tracef("Error generating source blocks: %v", err)
		return "", err
	}

	sourceToPipelineLabels, err = engine.generateSourceToPipelineLabels(pipelineNames)
	if err != nil {
		logger.Tracef("Error generating source to pipeline blocks: %v", err)
		return "", err
	}
	sort.Strings(sourceToPipelineLabels)

	// Omit generation for missing pipelines, i.e. legacy methods don't provide any
	if len(forwarding.Pipelines) > 0 {
		pipelineToOutputLabels, err = engine.generatePipelineToOutputLabels(forwarding.Pipelines)
		if err != nil {
			logger.Tracef("Error generating pipeline to output labels blocks: %v", err)
			return "", err
		}
	}

	// Omit generation for missing outputs, i.e. legacy methods provide them via configmap
	if len(forwarding.Outputs) > 0 {
		outputLabels, err = engine.generateOutputLabelBlocks(forwarding.Outputs)
		if err != nil {
			logger.Tracef("Error generating to output label blocks: %v", err)
			return "", err
		}
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
		logTypes.Has(string(logforward.LogSourceTypeInfra)),
		logTypes.Has(string(logforward.LogSourceTypeApp)),
		logTypes.Has(string(logforward.LogSourceTypeAudit)),
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

func gatherLogSourceTypes(pipelines []logforward.PipelineSpec) (sets.String, sets.String) {
	types := sets.NewString()
	appNamespaces := sets.NewString()
	for _, pipeline := range pipelines {
		types.Insert(string(pipeline.SourceType))
		if pipeline.SourceType == logforward.LogSourceTypeApp {
			appNamespaces.Insert(pipeline.Namespaces...)
		}
	}
	return types, appNamespaces
}

func mapSourceTypesToPipelineNames(pipelines []logforward.PipelineSpec) map[logforward.LogSourceType][]string {
	result := map[logforward.LogSourceType][]string{}
	for _, pipeline := range pipelines {
		names, ok := result[pipeline.SourceType]
		if !ok {
			names = []string{}
		}
		names = append(names, pipeline.Name)
		result[pipeline.SourceType] = names
	}
	return result
}

//generateSourceToPipelineLabels generates fluentd label stanzas to fan source types to multiple pipelines
func (engine *ConfigGenerator) generateSourceToPipelineLabels(sourcesToPipelines map[logforward.LogSourceType][]string) ([]string, error) {
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
			pipelineNames,
		}
		result, err := engine.Execute("sourceToPipelineCopyTemplate", data)
		if err != nil {
			return nil, fmt.Errorf("Error processing sourceToPipelineCopyTemplate template: %v", err)
		}
		configs = append(configs, result)
	}
	return configs, nil
}

func (engine *ConfigGenerator) generatePipelineToOutputLabels(pipelines []logforward.PipelineSpec) ([]string, error) {
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
//	  @type elasticsearch
//  </store>
//  </match>
//  <match **>
//    @type copy
//  </match>
// </label>
func (engine *ConfigGenerator) generateOutputLabelBlocks(outputs []logforward.OutputSpec) ([]string, error) {
	configs := []string{}
	logger.Debugf("Evaluating %v forwarding outputs...", len(outputs))
	for _, output := range outputs {
		logger.Debugf("Generate output type %v", output.Type)
		outputTemplateName := "outputLabelConf"
		var storeTemplateName string
		switch output.Type {
		case logforward.OutputTypeElasticsearch:
			storeTemplateName = "storeElasticsearch"
		case logforward.OutputTypeForward:
			storeTemplateName = "forward"
			outputTemplateName = "outputLabelConfNoCopy"
		case logforward.OutputTypeSyslog:
			if engine.useOldRemoteSyslogPlugin {
				storeTemplateName = "storeSyslogOld"
			} else {
				storeTemplateName = "storeSyslog"
			}
			outputTemplateName = "outputLabelConfNoRetry"
		default:
			logger.Warnf("Pipeline targets include an unrecognized type: %q", output.Type)
			continue
		}
		conf := newOutputLabelConf(engine.Template, storeTemplateName, output)
		result, err := engine.Execute(outputTemplateName, conf)
		if err != nil {
			return nil, fmt.Errorf("Error generating fluentd config Processing template outputLabelConf: %v", err)
		}
		configs = append(configs, result)
	}
	logger.Debugf("Generated output configurations: %v", configs)
	return configs, nil
}
