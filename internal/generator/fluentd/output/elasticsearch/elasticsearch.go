package elasticsearch

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
)

const (
	defaultElasticsearchPort = "9200"
	KeyStructured            = "structured"
)

type Elasticsearch struct {
	Desc            string
	StoreID         string
	Host            string
	Port            string
	RetryTag        Element
	SecurityConfig  []Element
	BufferConfig    []Element
	SuppressVersion []Element
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
{{compose .SecurityConfig}}
target_index_key viaq_index_name
id_key viaq_msg_id
remove_keys viaq_index_name
{{ compose  .SuppressVersion }}
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
				ViaqDataModel(bufspec, secret, o, op),
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
	u, _ := url.Parse(o.URL)
	port := u.Port()
	if port == "" {
		port = defaultElasticsearchPort
	}
	storeID := helpers.StoreID("", o.Name, "")
	es := Elasticsearch{
		StoreID:         storeID,
		Host:            u.Hostname(),
		Port:            port,
		SecurityConfig:  SecurityConfig(o, secret),
		BufferConfig:    output.Buffer(output.NOKEYS, bufspec, storeID, &o),
		SuppressVersion: SuppressVersion(o),
	}

	return es
}

func SuppressVersion(o logging.OutputSpec) []Element {
	if o.Elasticsearch != nil && o.Elasticsearch.Version >= logging.FirstESVersionWithoutType {
		return []Element{KV("suppress_type_name", "true")}
	} else {
		return []Element{
			KV("verify_es_version_at_startup", "false"),
			KV("type_name", "_doc"),
		}
	}
}

func RetryOutput(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Elasticsearch {
	es := Output(bufspec, secret, o, op)
	es.StoreID = helpers.StoreID("retry_", o.Name, "")
	es.BufferConfig = output.Buffer(output.NOKEYS, bufspec, es.StoreID, &o)
	return es
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	// URL is passable, checked at input sanitization
	u, _ := url.Parse(o.URL)
	isHttps := url.IsTLSScheme(u.Scheme)
	isSSLVerify := !(o.TLS != nil && o.TLS.InsecureSkipVerify)
	esTls := EsTLS{
		security.TLS(isHttps),
		isSSLVerify,
	}
	conf := append([]Element{}, esTls)
	if o.Secret != nil {
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
	}
	return conf
}
