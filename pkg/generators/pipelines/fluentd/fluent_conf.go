package fluentd

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

type sourceLabelCopyConf struct {
	Source       string
	TargetLabels []string
}

type outputLabelConf struct {
	Source          string
	Target          logging.PipelineTargetSpec
	Counter         int
	fluentTags      sets.String
	TemplateContext *template.Template
	hints           sets.String
}

func newOutputLabelConf(t *template.Template, source string, target logging.PipelineTargetSpec, counter int, fluentTags ...string) *outputLabelConf {
	return &outputLabelConf{
		Source:          source,
		Target:          target,
		Counter:         counter,
		TemplateContext: t,
		fluentTags:      sets.NewString(fluentTags...),
	}
}

func (conf *outputLabelConf) SetHints(hints []string) {
	conf.hints = sets.NewString(hints...)
}
func (conf *outputLabelConf) Hints() sets.String {
	return conf.hints
}
func (conf *outputLabelConf) Template() *template.Template {
	return conf.TemplateContext
}
func (conf *outputLabelConf) Host() string {
	return strings.Split(conf.Target.Endpoint, ":")[0]
}

func (conf *outputLabelConf) Port() string {
	parts := strings.Split(conf.Target.Endpoint, ":")
	if len(parts) == 1 {
		return "9200"
	}
	return parts[1]
}

func (conf *outputLabelConf) BufferPath() string {
	return fmt.Sprintf("/var/lib/fluentd/%s", conf.StoreID())
}
func (conf *outputLabelConf) SecretPath(file string) string {
	return fmt.Sprintf("/var/run/ocp-collector/secrets/%s/%s", conf.Target.Certificates.SecretName, file)
}

func (conf *outputLabelConf) LabelName() string {
	return labelName(conf.Source)
}

func labelName(source string) string {
	return strings.ToUpper(fmt.Sprintf("@%v", strings.Join(strings.Split(source, "."), "_")))
}

func (conf *outputLabelConf) ReLabelName() string {
	return reLabelName(conf.Source, conf.Target.Type, conf.Counter)
}

func reLabelName(source string, targetType logging.PipelineTargetType, counter int) string {
	return strings.ToUpper(fmt.Sprintf("@%v_%v%v", strings.Join(strings.Split(source, "."), "_"), targetType, counter))
}

func (conf *outputLabelConf) StoreID() string {
	prefix := ""
	if conf.Hints().Has("prefix_as_retry") {
		prefix = "retry_"
	}
	return strings.ToLower(fmt.Sprintf("%v%v_%v%v", prefix, strings.Join(strings.Split(conf.Source, "."), "_"), conf.Target.Type, conf.Counter))
}

func (conf *outputLabelConf) RetryTag() string {
	source := conf.Source
	return "retry_" + strings.ToLower(strings.Join(strings.Split(source, "."), "_"))
}
func (conf *outputLabelConf) Tags() string {
	return fmt.Sprintf("%s", strings.Join(conf.fluentTags.List(), " "))
}

type targetTypeCounterMap map[logging.PipelineTargetType]int

func newTargetTypeCounterMap() *targetTypeCounterMap {
	counters := make(targetTypeCounterMap)
	for _, t := range []logging.PipelineTargetType{logging.PipelineTargetTypeElasticsearch} {
		counters[t] = 0
	}
	return &counters
}

//bump the counter map if the type is recognized
func (t targetTypeCounterMap) bump(targetType logging.PipelineTargetType) (int, bool) {
	if counter, ok := t[targetType]; ok {
		t[targetType]++
		return counter, true
	}
	return 0, false
}

func newSourceLabelCopyConf(source string, targets []logging.PipelineTargetSpec) *sourceLabelCopyConf {

	counters := newTargetTypeCounterMap()
	targetLabels := []string{}
	for _, target := range targets {
		if counter, ok := counters.bump(target.Type); ok {
			targetLabels = append(targetLabels, reLabelName(source, target.Type, counter))
		} else {
			logrus.Warnf("Pipeline targets include an unrecognized type: %s", target.Type)
		}
	}
	return &sourceLabelCopyConf{
		source,
		targetLabels,
	}
}
