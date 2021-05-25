package fluentd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/url"
)

var replacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")

type inputSelectorConf struct {
	Pipeline   string
	Namespaces string
	Labels     string
}

func newInputSelectorConf(pipeline string, namespaces []string, labelSelector *metav1.LabelSelector) (*inputSelectorConf, error) {
	labelList := ""
	var names []string

	labelMap, err := metav1.LabelSelectorAsMap(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("LabelSelector: %v", err)
	}
	for name := range labelMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, k := range names {
		if labelList != "" {
			labelList += ","
		}
		labelList += fmt.Sprintf("%s:%s", k, labelMap[k])
	}

	return &inputSelectorConf{
		Pipeline:   pipeline,
		Namespaces: strings.Join(namespaces, ","),
		Labels:     labelList,
	}, nil
}

type outputLabelConf struct {
	Name            string
	Target          logging.OutputSpec
	forwarder       *logging.ForwarderSpec
	Counter         int
	fluentTags      sets.String
	TemplateContext *template.Template
	hints           sets.String
	storeTemplate   string
	URL             *url.URL
	Secret          *corev1.Secret
}

func newOutputLabelConf(t *template.Template, storeTemplate string, target logging.OutputSpec, secret *corev1.Secret, config *logging.ForwarderSpec, fluentTags ...string) (*outputLabelConf, error) {
	u, err := url.Parse(target.URL)
	if err != nil {
		return nil, fmt.Errorf("url field: %v", err)
	}
	if target.Type == logging.OutputTypeSyslog && target.Syslog == nil {
		target.Syslog = &logging.Syslog{RFC: "RFC5424"}
	}
	return &outputLabelConf{
		Name:            target.Name,
		Target:          target,
		TemplateContext: t,
		forwarder:       config,
		fluentTags:      sets.NewString(fluentTags...),
		storeTemplate:   storeTemplate,
		URL:             u,
		Secret:          secret,
	}, nil
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
func (conf *outputLabelConf) Host() string { return conf.URL.Hostname() }

func (conf *outputLabelConf) Port() string {
	p := conf.URL.Port()
	if p == "" {
		return "9200"
	}
	return p
}

// Protocol returns the insecure protocol name used in fluentd configuration.
func (conf *outputLabelConf) Protocol() string { return url.PlainScheme(conf.URL.Scheme) }

func (conf *outputLabelConf) LogGroupName() string {
	if conf.Target.Type == logging.OutputTypeCloudwatch {
		switch conf.Target.Cloudwatch.GroupBy {
		case logging.LogGroupByNamespaceName:
			return "${record['kubernetes']['namespace_name']}"
		case logging.LogGroupByNamespaceUUID:
			return "${record['kubernetes']['namespace_id']}"
		default:
			return logging.InputNameApplication
		}
	}
	return ""
}
func (conf *outputLabelConf) LogGroupPrefix() string {
	if conf.Target.Type == logging.OutputTypeCloudwatch {
		prefix := conf.Target.Cloudwatch.GroupPrefix
		if prefix != nil && strings.TrimSpace(*prefix) != "" {
			return fmt.Sprintf("%s.", *prefix)
		}
	}
	return ""
}

func (conf *outputLabelConf) BufferPath() string {
	return fmt.Sprintf("/var/lib/fluentd/%s", conf.StoreID())
}

func (conf *outputLabelConf) IsTLS() bool {
	return url.IsTLSScheme(conf.URL.Scheme) || conf.Secret != nil
}

func (conf *outputLabelConf) SecretPath(file string) string {
	if conf.Target.Secret != nil {
		return filepath.Join(constants.CollectorSecretsDir, conf.Target.Secret.Name, file)
	}
	return ""
}

func (conf *outputLabelConf) SecretPathIfFound(file string) string {
	if conf.Secret != nil {
		if _, ok := conf.Secret.Data[file]; ok {
			return conf.SecretPath(file)
		}
	}
	return ""
}

func (conf *outputLabelConf) GetSecret(key string) string {
	if conf.Secret != nil {
		return string(conf.Secret.Data[key])
	}
	return ""
}

func (conf *outputLabelConf) LabelName() string {
	return labelName(conf.Name)
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
	return strings.Join(conf.fluentTags.List(), " ")
}

func (conf *outputLabelConf) IsElasticSearchOutput() bool {
	return conf.Target.Type == logging.OutputTypeElasticsearch
}

func (conf *outputLabelConf) NeedChangeElasticsearchStructuredIndexName() bool {
	return conf.Target.Type == logging.OutputTypeElasticsearch &&
		conf.Target.OutputTypeSpec.Elasticsearch != nil &&
		(conf.Target.OutputTypeSpec.Elasticsearch.StructuredIndexKey != "" || conf.Target.OutputTypeSpec.Elasticsearch.StructuredIndexName != "")
}

func generateRubyDigArgs(path string) string {
	var args []string
	for _, s := range strings.Split(path, ".") {
		args = append(args, fmt.Sprintf("%q", s))
	}
	return strings.Join(args, ",")
}

func (conf *outputLabelConf) GetKeyVal(path string) string {
	return generateRubyDigArgs(path)
}
