package fluentd

import (
	"fmt"
	"text/template"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators"
)

var (
	defaultInfraProjectPatterns = []string{"default", "openshift", "openshift-", "kube-"}
)

//sourceTags are the tags used by the fluentd pipeline to match log messages (e.g. **_foo_bar**) and
// are used in a <match **_foo_bar_**> block
type sourceTags map[string][]string

func newSourceTags(source string, tags ...string) sourceTags {
	sourceTags := make(map[string][]string)
	sourceTags[source] = tags
	return sourceTags
}

func (st sourceTags) insert(source string, tags ...string) sourceTags {
	st[source] = tags
	return st
}

func formatFluentTags() sourceTags {
	sourceTags := newSourceTags(string(logging.PipelineSourceTypeLogsApp), "**")
	infraTags := sets.NewString()
	for _, pattern := range defaultInfraProjectPatterns {
		if pattern[len(pattern)-1] == '-' {
			infraTags.Insert(fmt.Sprintf("**_%s_**", pattern))
		} else {
			infraTags.Insert(pattern)
		}
	}
	return sourceTags
}

//ConfigGenerator is a config generator for fluentd
type ConfigGenerator struct {
	*generators.Generator
}

//NewConfigGenerator creates an instance of FluentdConfigGenerator
func NewConfigGenerator() (*ConfigGenerator, error) {
	engine, err := generators.New("OutputLabelConf",
		&template.FuncMap{
			"labelName":   labelName,
			"relabelName": reLabelName,
		},
		templateRegistry...)
	if err != nil {
		return nil, err
	}
	return &ConfigGenerator{engine}, nil
}

//Generate the fluent.conf file using the pipeline information
func (engine *ConfigGenerator) Generate(pipeline *logging.PipelinesSpec) (string, error) {

	//sanitize here
	sourceMatchers := []string{}
	copyLabels := []string{}
	storeLabels := []string{}
	for source, targets := range pipeline.Map() {
		if targets == nil {
			logrus.Infof("Skipping pipeline source %s, there are no targets", source)
			continue
		}
		if matchers, err := engine.generateSourceMatchBlocks(formatFluentTags()); err == nil {
			sourceMatchers = append(sourceMatchers, matchers...)
		} else {
			return "", err
		}
		if copies, err := engine.generateLabelCopyBlocks(source, targets.Targets); err == nil {
			copyLabels = append(copyLabels, copies...)
		} else {
			return "", err
		}
		if stores, err := engine.generateStoreLabelBlocks(source, targets.Targets); err == nil {
			storeLabels = append(storeLabels, stores...)
		} else {
			return "", err
		}
	}
	data := struct {
		SourceMatchers []string
		CopyLabels     []string
		StoreLabels    []string
	}{
		sourceMatchers,
		copyLabels,
		storeLabels,
	}
	result, err := engine.Execute("fluentConf", data)
	if err != nil {
		return "", fmt.Errorf("Error processing fluentConf template: %v", err)
	}
	return result, nil
}

//generateLabelCopyBlocks generates fluentd label stanzas for sources to copy to target labels
func (engine *ConfigGenerator) generateLabelCopyBlocks(source string, targets []logging.PipelineTargetSpec) ([]string, error) {
	configs := []string{}
	conf := newSourceLabelCopyConf(source, targets)
	result, err := engine.Execute("sourceLabelCopy", conf)
	if err != nil {
		return nil, fmt.Errorf("Error processing sourceLabelCopy template: %v", err)
	}
	configs = append(configs, result)
	return configs, nil
}

//generateSourceMatchBlocks generates fluentd match stanzas for sources to a given set of tags
func (engine *ConfigGenerator) generateSourceMatchBlocks(sourceTags sourceTags) ([]string, error) {
	configs := []string{}
	for source, fluentTags := range sourceTags {
		// reuse outputLabelConf where counter (e.g. 0) is not evaluated
		conf := newOutputLabelConf(engine.Template, source, logging.PipelineTargetSpec{}, 0, fluentTags...)
		result, err := engine.Execute("outputLabelMatch", conf)
		if err != nil {
			return nil, fmt.Errorf("Error generating fluentd config Processing template %s: %v", engine.Template.Name(), err)
		}
		configs = append(configs, result)
	}
	return configs, nil
}

//generateStoreLabelBlocks generates fluentd label stanzas for sources to specific store destinations
func (engine *ConfigGenerator) generateStoreLabelBlocks(source string, targets []logging.PipelineTargetSpec) ([]string, error) {
	configs := []string{}
	counters := newTargetTypeCounterMap()
	for _, target := range targets {
		if counter, ok := counters.bump(target.Type); ok {
			conf := newOutputLabelConf(engine.Template, source, target, counter)
			result, err := engine.Execute("outputLabelConf", conf)
			if err != nil {
				return nil, fmt.Errorf("Error generating fluentd config Processing template outputLabelConf: %v", err)
			}
			configs = append(configs, result)
		} else {
			logrus.Warnf("Pipeline targets include an unrecognized type: %s", target.Type)
		}
	}
	return configs, nil
}
