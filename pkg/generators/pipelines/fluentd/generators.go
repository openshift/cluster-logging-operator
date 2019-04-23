package fluentd

import (
	"fmt"
	"text/template"

	"github.com/sirupsen/logrus"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators"
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
		outputLabelConfTemplate,
		storeElasticsearchTemplate,
		outputLabelMatchTemplate,
		sourceLabelCopyTemplate)
	if err != nil {
		return nil, err
	}
	return &ConfigGenerator{engine}, nil
}

//GenerateSourceLabelCopyStanzas generates fluentd label stanzas for sources to copy to target labels
func (engine *ConfigGenerator) GenerateSourceLabelCopyStanzas(source string, targets []logging.PipelineTargetSpec) ([]string, error) {
	configs := []string{}
	conf := newSourceLabelCopyConf(source, targets)
	result, err := engine.Execute("sourceLabelCopy", conf)
	if err != nil {
		return nil, fmt.Errorf("Error generating fluentd config Processing template %s: %v", engine.Template.Name(), err)
	}
	configs = append(configs, result)
	return configs, nil
}

//GenerateOutputLabelMatchStanzas generates fluentd match stanzas for sources to a given set of tags
func (engine *ConfigGenerator) GenerateOutputLabelMatchStanzas(sourceTags sourceTags) ([]string, error) {
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

//GenerateOutputLabelConf generates fluentd label stanzas for sources to a given destination
func (engine *ConfigGenerator) GenerateOutputLabelConf(source string, targets []logging.PipelineTargetSpec) ([]string, error) {
	configs := []string{}
	counters := newTargetTypeCounterMap()
	for _, target := range targets {
		if counter, ok := counters.bump(target.Type); ok {
			conf := newOutputLabelConf(engine.Template, source, target, counter)
			result, err := engine.Execute("outputLabelConf", conf)
			if err != nil {
				return nil, fmt.Errorf("Error generating fluentd config Processing template %s: %v", engine.Template.Name(), err)
			}
			configs = append(configs, result)
		} else {
			logrus.Warnf("Pipline targets include an unrecognized type: %s", target.Type)
		}
	}
	return configs, nil
}
