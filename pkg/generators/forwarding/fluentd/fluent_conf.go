package fluentd

import (
	"fmt"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/url"
)

var replacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")

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
}

func newOutputLabelConf(t *template.Template, storeTemplate string, target logging.OutputSpec, config *logging.ForwarderSpec, fluentTags ...string) (*outputLabelConf, error) {
	u, err := url.ParseAbsoluteOrEmpty(target.URL)
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

// Protocol returns the insecure base protocol name used in fluentd configuration.
func (conf *outputLabelConf) Protocol() string {
	protocol := strings.ToLower(conf.URL.Scheme)
	switch protocol {
	case "tls":
		return "tcp" // Fluentd uses "tcp" for TLS connections, TLS is dealt with elsewhere.
	case "udps":
		return "udp"
	default:
		return protocol
	}
}

func (conf *outputLabelConf) LogRetentionDays() int {
	if conf.Target.Type == logging.OutputTypeCloudwatch {
		return conf.Target.Cloudwatch.LogGroupStrategy.RetentionInDays
	}
	return 0
}
func (conf *outputLabelConf) LogGroupName() string {
	if conf.Target.Type == logging.OutputTypeCloudwatch {
		switch conf.Target.Cloudwatch.LogGroupStrategy.Name {
		case logging.LogGroupStrategyTypeNamespace:
			return `${record["kubernetes"]["namespace_name"]}`
		default:
			return ""
		}
		return fmt.Sprintf("/var/lib/fluentd/%s", conf.StoreID())
	}
	return ""
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
	return strings.Join(conf.fluentTags.List(), " ")
}
