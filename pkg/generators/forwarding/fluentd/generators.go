package fluentd

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	defaultInfraProjectPatterns = []string{"default", "openshift", "openshift-", "kube-"}
)

//ConfigGenerator is a config generator for fluentd
type ConfigGenerator struct {
	*generators.Generator
}

//NewConfigGenerator creates an instance of FluentdConfigGenerator
func NewConfigGenerator() (*ConfigGenerator, error) {
	engine, err := generators.New("OutputLabelConf",
		&template.FuncMap{
			"labelName":           labelName,
			"sourceTypelabelName": sourceTypeLabelName,
		},
		templateRegistry...)
	if err != nil {
		return nil, err
	}
	return &ConfigGenerator{engine}, nil
}

//Generate the fluent.conf file using the forwarding information
func (engine *ConfigGenerator) Generate(forwarding *logging.ForwardingSpec) (string, error) {

	//sanitize here
	sourceInputLabels := []string{}
	sourceToPipelineLabels := []string{}
	pipelineToOutputLabels := []string{}
	var outputLabels []string
	var err error

	logTypes := gatherLogSourceTypes(forwarding.Pipelines)

	if sourceInputLabels, err = engine.generateSource(logTypes); err != nil {
		return "", err
	}
	if sourceToPipelineLabels, err = engine.generateSourceToPipelineLabels(mapSourceTypesToPipelineNames(forwarding.Pipelines)); err != nil {
		return "", err
	}
	sort.Strings(sourceToPipelineLabels)
	if pipelineToOutputLabels, err = engine.generatePipelineToOutputLabels(forwarding.Pipelines); err != nil {
		return "", err
	}
	if outputLabels, err = engine.generateOutputLabelBlocks(forwarding.Outputs); err != nil {
		return "", err
	}

	data := struct {
		CollectInfraLogs        bool
		CollectAppLogs          bool
		SourceInputLabels       []string
		SourceToPipelineLabels  []string
		PipelinesToOutputLabels []string
		OutputLabels            []string
	}{
		logTypes.Has(string(logging.LogSourceTypeApp)),
		logTypes.Has(string(logging.LogSourceTypeInfra)),
		sourceInputLabels,
		sourceToPipelineLabels,
		pipelineToOutputLabels,
		outputLabels,
	}
	result, err := engine.Execute("fluentConf", data)
	if err != nil {
		return "", fmt.Errorf("Error processing fluentConf template: %v", err)
	}
	return pretty(result), nil
}

func gatherLogSourceTypes(pipelines []logging.PipelineSpec) sets.String {
	types := sets.NewString()
	for _, pipeline := range pipelines {
		types.Insert(string(pipeline.SourceType))
	}
	return types
}

func mapSourceTypesToPipelineNames(pipelines []logging.PipelineSpec) map[logging.LogSourceType][]string {
	result := map[logging.LogSourceType][]string{}
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

func pretty(in string) string {
	stack := -1
	out := bytes.NewBufferString("")
	for _, line := range strings.Split(in, "\n") {
		stack = prettyLine(out, line, stack)
	}
	return out.String()
}
func prettyLine(out *bytes.Buffer, in string, levelIn int) int {
	levelOut := levelIn
	trimmed := strings.Trim(in, " \t")
	if strings.HasPrefix(trimmed, "</") {
		levelOut = levelIn - 1
	} else if strings.HasPrefix(trimmed, "<") && !strings.HasPrefix(trimmed, "</") {
		levelOut = levelIn + 1
	}

	for i := 0; i < levelOut; i++ {
		out.WriteString("\t")
	}
	out.WriteString(trimmed)
	out.WriteString("\n")
	return levelOut
}

//generateSourceToPipelineLabels generates fluentd label stanzas to fan source types to multiple pipelines
func (engine *ConfigGenerator) generateSourceToPipelineLabels(sourcesToPipelines map[logging.LogSourceType][]string) ([]string, error) {
	configs := []string{}
	for sourceType, pipelineNames := range sourcesToPipelines {
		data := struct {
			Source        string
			PipelineNames []string
		}{
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
//	  @type elasticsearch
//  </store>
//  </match>
//  <match **>
//    @type copy
//  </match>
// </label>
func (engine *ConfigGenerator) generateOutputLabelBlocks(outputs []logging.OutputSpec) ([]string, error) {
	configs := []string{}
	logger.Debugf("Evaluating %v forwarding outputs...", len(outputs))
	for _, output := range outputs {
		logger.Debugf("Generate output type %v", output.Type)
		outputTemplateName := "outputLabelConf"
		var storeTemplateName string
		switch output.Type {
		case logging.OutputTypeElasticsearch:
			storeTemplateName = "storeElasticsearch"
		case logging.OutputTypeForward:
			storeTemplateName = "forward"
			outputTemplateName = "outputLabelConfNoCopy"
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
