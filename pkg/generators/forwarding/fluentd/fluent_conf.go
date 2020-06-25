package fluentd

import (
	"fmt"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var replacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")
var protocolSeparator = "://"

type outputLabelConf struct {
	Name            string
	Target          logging.OutputSpec
	Counter         int
	fluentTags      sets.String
	TemplateContext *template.Template
	hints           sets.String
	storeTemplate   string
}

func newOutputLabelConf(t *template.Template, storeTemplate string, target logging.OutputSpec, fluentTags ...string) *outputLabelConf {
	return &outputLabelConf{
		Name:            target.Name,
		Target:          target,
		TemplateContext: t,
		fluentTags:      sets.NewString(fluentTags...),
		storeTemplate:   storeTemplate,
	}
}
func (conf *outputLabelConf) StoreTemplate() string {
	return conf.storeTemplate
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
	endpoint := stripProtocol(conf.Target.URL)
	return strings.Split(endpoint, ":")[0]
}

func (conf *outputLabelConf) Port() string {
	endpoint := stripProtocol(conf.Target.URL)
	parts := strings.Split(endpoint, ":")
	if len(parts) == 1 {
		return "9200"
	}
	return parts[1]
}

func (conf *outputLabelConf) Protocol() string {
	endpoint := conf.Target.URL
	if index := strings.Index(endpoint, protocolSeparator); index != -1 {
		return endpoint[:index]
	}
	return "tcp"
}

func stripProtocol(endpoint string) string {
	if index := strings.Index(endpoint, protocolSeparator); index != -1 {
		endpoint = endpoint[index+len(protocolSeparator):]
	}
	return endpoint
}

func (conf *outputLabelConf) BufferPath() string {
	return fmt.Sprintf("/var/lib/fluentd/%s", conf.StoreID())
}
func (conf *outputLabelConf) SecretPath(file string) string {
	return fmt.Sprintf("/var/run/ocp-collector/secrets/%s/%s", conf.Target.Secret.Name, file)
}

func (conf *outputLabelConf) LabelName() string {
	return labelName(conf.Name)
}

func labelName(name string) string {
	return strings.ToUpper(fmt.Sprintf("@%s", replacer.Replace(name)))
}

func sourceTypeLabelName(name string) string {
	return strings.ToUpper(fmt.Sprintf("@_%s", replacer.Replace(name)))
}

func (conf *outputLabelConf) StoreID() string {
	prefix := ""
	if conf.Hints().Has("prefix_as_retry") {
		prefix = "retry_"
	}
	return strings.ToLower(fmt.Sprintf("%v%v", prefix, replacer.Replace(conf.Name)))
}

func (conf *outputLabelConf) RetryTag() string {
	return "retry_" + strings.ToLower(replacer.Replace(conf.Name))
}
func (conf *outputLabelConf) Tags() string {
	return fmt.Sprintf("%s", strings.Join(conf.fluentTags.List(), " "))
}
