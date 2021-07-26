package elasticsearch

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	. "github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
	genhelper "github.com/openshift/cluster-logging-operator/pkg/generator/helpers"
	"github.com/openshift/cluster-logging-operator/pkg/generator/url"
	urlhelper "github.com/openshift/cluster-logging-operator/pkg/generator/url"
)

const (
	defaultElasticsearchPort = "9200"
	KeyStructured            = "structured"
)

type Elasticsearch struct {
	Desc           string
	StoreID        string
	Host           string
	Port           string
	RetryTag       Element
	SecurityConfig []Element
	BufferConfig   []Element
}

func (e Elasticsearch) Name() string {
	return "elasticsearchTemplate"
}

func (e Elasticsearch) Template() string {
	return `{{define "` + e.Name() + `" -}}
{{if .Desc -}}
# {{.Desc}}
{{ end -}}
@type elasticsearch
@id {{.StoreID}}
host {{.Host}}
port {{.Port}}
verify_es_version_at_startup false
{{compose .SecurityConfig}}
target_index_key viaq_index_name
id_key viaq_msg_id
remove_keys viaq_index_name
type_name _doc
{{ kv .RetryTag -}}
http_backend typhoeus
write_operation create
reload_connections 'true'
# https://github.com/uken/fluent-plugin-elasticsearch#reload-after
reload_after '200'
# https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
reload_on_failure false
# 2 ^ 31
request_timeout 2147483648
{{compose .BufferConfig}}
{{- end}}
`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: MergeElements(
				ChangeESIndex(bufspec, secret, o, op),
				FlattenLabels(bufspec, secret, o, op),
				OutputConf(bufspec, secret, o, op),
			),
		},
	}
}

func OutputConf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			genhelper.DebugOutput,
		}
	}
	es := Output(bufspec, secret, o, op)
	retryTag := helpers.StoreID("retry_", o.Name, "")
	es.RetryTag = KV("retry_tag", retryTag)
	return []Element{
		Match{
			MatchTags:    retryTag,
			MatchElement: RetryOutput(bufspec, secret, o, op),
		},
		Match{
			MatchTags:    "**",
			MatchElement: es,
		},
	}
}

func Output(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Elasticsearch {
	// URL is parasable, checked at input sanitization
	u, _ := urlhelper.Parse(o.URL)
	port := u.Port()
	if port == "" {
		port = defaultElasticsearchPort
	}
	storeID := helpers.StoreID("", o.Name, "")
	return Elasticsearch{
		StoreID:        storeID,
		Host:           u.Hostname(),
		Port:           port,
		SecurityConfig: SecurityConfig(o, secret),
		BufferConfig:   output.Buffer(output.NOKEYS, bufspec, storeID, &o),
	}
}

func RetryOutput(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Elasticsearch {
	es := Output(bufspec, secret, o, op)
	es.StoreID = helpers.StoreID("retry_", o.Name, "")
	es.BufferConfig = output.Buffer(output.NOKEYS, bufspec, es.StoreID, &o)
	return es
}

func ChangeESIndex(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	if o.Elasticsearch != nil && (o.Elasticsearch.StructuredTypeKey != "" || o.Elasticsearch.StructuredTypeName != "") {
		return []Element{
			Filter{
				MatchTags: "**",
				Element: RecordModifier{
					Records: []Record{
						{
							Key:        "typeFromKey",
							Expression: (fmt.Sprintf("${record.dig(%s)}", generateRubyDigArgs(o.Elasticsearch.StructuredTypeKey))),
						},
						{
							Key:        "hasStructuredTypeName",
							Expression: o.Elasticsearch.StructuredTypeName,
						},
						{
							Key:        "viaq_index_name",
							Expression: `${ if !record['structured'].nil? && record['structured'] != {}; if !record['typeFromKey'].nil?; "app-"+record['typeFromKey']+"-write"; elsif record['hasStructuredTypeName'] != ""; "app-"+record['hasStructuredTypeName']+"-write"; else record['viaq_index_name']; end; else record['viaq_index_name']; end;}`,
						},
					},
					RemoveKeys: []string{"typeFromKey", "hasStructuredTypeName"},
				},
			},
		}
	}

	return []Element{
		Filter{
			Desc:      "remove structured field if present",
			MatchTags: "**",
			Element: RecordModifier{
				RemoveKeys: []string{KeyStructured},
			},
		},
	}
}

func FlattenLabels(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	return []Element{
		Filter{
			Desc:      "flatten labels to prevent field explosion in ES",
			MatchTags: "**",
			Element: RecordTransformer{
				Records: []Record{
					{
						Key:        "kubernetes",
						Expression: `${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }`,
					},
				},
				RemoveKeys: []string{"$.kubernetes.labels"},
			},
		},
	}
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	// URL is parasable, checked at input sanitization
	u, _ := urlhelper.Parse(o.URL)
	conf := []Element{}
	if o.Secret != nil {
		conf = append(conf, TLS(url.IsTLSScheme(u.Scheme)))
		if security.HasUsernamePassword(secret) {
			up := UserNamePass{
				UsernamePath: security.SecretPath(o.Secret.Name, constants.ClientUsername),
				PasswordPath: security.SecretPath(o.Secret.Name, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if o.Name == logging.OutputNameDefault || security.HasTLSCertAndKey(secret) {
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if o.Name == logging.OutputNameDefault || security.HasCABundle(secret) {
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
	} else {
		conf = append(conf, TLS(false))
	}
	return conf
}

func generateRubyDigArgs(path string) string {
	var args []string
	for _, s := range strings.Split(path, ".") {
		args = append(args, fmt.Sprintf("%q", s))
	}
	return strings.Join(args, ",")
}
