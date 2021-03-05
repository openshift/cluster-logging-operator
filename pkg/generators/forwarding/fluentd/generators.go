package fluentd

import (
	"encoding/json"
	"fmt"
	"sort"
	"text/template"

	"github.com/openshift/cluster-logging-operator/pkg/constants"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators"
	corev1 "k8s.io/api/core/v1"
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
func (engine *ConfigGenerator) Generate(clfSpec *logging.ClusterLogForwarderSpec, secrets map[string]*corev1.Secret, fwSpec *logging.ForwarderSpec) (string, error) {
	//sanitize here
	var (
		inputs                 sets.String
		namespaces             sets.String
		routeMap               logging.RouteMap
		sourceInputLabels      []string
		sourceToPipelineLabels []string
		pipelineToOutputLabels []string
		outputLabels           []string
		err                    error
	)

	inputs, namespaces = gatherSources(clfSpec)
	routeMap = inputsToPipelines(clfSpec)
	// Provide inputs and inputsPipelines for legacy forwarding protocols
	// w/o a user-provided ClusterLogFowarder instance to enable inputs and
	// inputs-to-pipelines a.k.a. route map template generation.
	if engine.includeLegacyForwardConfig || engine.includeLegacySyslogConfig {
		inputs.Insert(
			logging.InputNameInfrastructure,
			logging.InputNameApplication,
			logging.InputNameAudit,
		)
		for _, logType := range inputs.List() {
			if engine.includeLegacySyslogConfig {
				routeMap.Insert(logType, constants.LegacySyslog)
			}
			if engine.includeLegacyForwardConfig {
				routeMap.Insert(logType, constants.LegacySecureforward)
			}
		}
	}

	sourceInputLabels, err = engine.generateSource(inputs)
	if err != nil {

		log.V(3).Error(err, "Error generating source blocks")
		return "", err
	}

	sourceToPipelineLabels, err = engine.generateSourceToPipelineLabels(routeMap)
	if err != nil {
		log.V(3).Error(err, "Error generating source to pipeline blocks")
		return "", err
	}
	sort.Strings(sourceToPipelineLabels)

	// Omit generation for missing pipelines, i.e. legacy methods don't provide any
	if len(clfSpec.Pipelines) > 0 {
		pipelineToOutputLabels, err = engine.generatePipelineToOutputLabels(clfSpec.Pipelines)
		if err != nil {
			log.V(3).Error(err, "Error generating pipeline to output labels blocks")
			return "", err
		}
	}

	// Omit generation for missing outputs, i.e. legacy methods provide them via configmap
	if len(clfSpec.Outputs) > 0 {
		outputLabels, err = engine.generateOutputLabelBlocks(clfSpec.Outputs, secrets, fwSpec)
		if err != nil {
			log.V(3).Error(err, "Error generating to output label blocks")
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
		AppNamespaces              []string
	}{
		engine.includeLegacyForwardConfig,
		engine.includeLegacySyslogConfig,
		inputs.Has(logging.InputNameInfrastructure),
		inputs.Has(logging.InputNameApplication),
		inputs.Has(logging.InputNameAudit),
		sourceInputLabels,
		sourceToPipelineLabels,
		pipelineToOutputLabels,
		outputLabels,
		namespaces.List(),
	}
	result, err := engine.Execute("fluentConf", data)
	if err != nil {
		log.V(3).Info("Error generating fluentConf")
		return "", fmt.Errorf("Error processing fluentConf template: %v", err)
	}
	log.V(3).Info("Successfully generated fluent.conf", "fluent.conf", result)
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

func inputsToPipelines(fwdspec *logging.ClusterLogForwarderSpec) logging.RouteMap {
	result := logging.RouteMap{}
	inputs := fwdspec.InputMap()
	for _, pipeline := range fwdspec.Pipelines {
		for _, inRef := range pipeline.InputRefs {
			if input, ok := inputs[inRef]; ok {
				// User defined input spec, unwrap.
				for t := range input.Types() {
					result.Insert(t, pipeline.Name)
				}
			} else {
				// Not a user defined type, insert direct.
				result.Insert(inRef, pipeline.Name)
			}
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
			sourceType,
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
		var jsonLabels string

		if pipeline.Labels != nil {
			marshalledLabels, err := json.Marshal(pipeline.Labels)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal pipeline labels: %v", err)
			}
			jsonLabels = string(marshalledLabels)
		}

		data := struct {
			Name           string
			Outputs        []string
			PipelineLabels string
			Parse          string
		}{
			pipeline.Name,
			pipeline.OutputRefs,
			jsonLabels,
			pipeline.Parse,
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
func (engine *ConfigGenerator) generateOutputLabelBlocks(outputs []logging.OutputSpec, secrets map[string]*corev1.Secret, outputConf *logging.ForwarderSpec) ([]string, error) {
	configs := []string{}
	for _, output := range outputs {
		log.V(3).Info("Generate output type", "type", output.Type)
		engine.outputTemplate = "outputLabelConf" // Default
		switch output.Type {
		case logging.OutputTypeElasticsearch:
			engine.storeTemplate = "storeElasticsearch"
		case logging.OutputTypeFluentdForward:
			engine.storeTemplate, engine.outputTemplate = "forward", "outputLabelConfNoCopy"
		case logging.OutputTypeSyslog:
			if engine.useOldRemoteSyslogPlugin {
				engine.storeTemplate = "storeSyslogOld"
			} else {
				engine.storeTemplate = "storeSyslog"
			}
			engine.outputTemplate = "outputLabelConfJsonParseNoRetry"
		case logging.OutputTypeKafka:
			engine.storeTemplate = "storeKafka"
			engine.outputTemplate = "outputLabelConfNoCopy"
		default:
			return nil, fmt.Errorf("Unknown output type: %v", output.Type)
		}
		conf, err := newOutputLabelConf(engine.Template, engine.storeTemplate, output, secrets[output.Name], outputConf)
		if err != nil {
			return nil, fmt.Errorf("generating fluentd output label: %v", err)
		}
		result, err := engine.Execute(engine.outputTemplate, conf)
		if err != nil {
			return nil, fmt.Errorf("generating fluent output label: %v", err)
		}
		configs = append(configs, result)
	}
	log.V(3).Info("Generated output configurations", "configurations", configs)
	return configs, nil
}
