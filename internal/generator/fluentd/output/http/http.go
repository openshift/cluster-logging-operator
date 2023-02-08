package http

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/internal/generator/utils"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
)

type Http struct {
	StoreID        string
	URI            string
	Method         string
	Headers        Element
	SecurityConfig []Element
	BufferConfig   []Element

	// Encoding is set by plugin according to
}

func (h Http) Name() string {
	return "fluentdHttpTemplate"
}

func (h Http) Template() string {
	return `{{define "` + h.Name() + `" -}}
@type http
endpoint {{.URI}}
http_method {{.Method}}
encoding "application/x-ndjson"
{{kv  .Headers -}}
{{compose .SecurityConfig}}
{{compose .BufferConfig}}
{{end}}`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []Element{
				Output(bufspec, secret, o, op),
			},
		},
	}
}

func Output(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	storeID := helpers.StoreID("", o.Name, "")
	return Match{
		MatchTags: "**",
		MatchElement: Http{
			StoreID:        strings.ToLower(helpers.Replacer.Replace(o.Name)),
			URI:            o.URL,
			Method:         Method(o.Http),
			Headers:        Headers(o),
			SecurityConfig: SecurityConfig(o, secret),
			BufferConfig:   output.Buffer(output.NOKEYS, bufspec, storeID, &o),
		},
	}
}

func Headers(o logging.OutputSpec) Element {
	if o.Http == nil || len(o.Http.Headers) == 0 {
		return Nil
	}
	return KV("headers", utils.ToHeaderStr(o.Http.Headers, "%q:%q"))
}

func Method(h *logging.Http) string {
	if h == nil {
		return "post"
	}
	switch h.Method {
	case "GET":
		return "get"
	case "HEAD":
		return "head"
	case "POST":
		return "post"
	case "PUT":
		return "put"
	case "DELETE":
		return "delete"
	case "OPTIONS":
		return "options"
	case "TRACE":
		return "trace"
	case "PATCH":
		return "patch"
	}
	return "post"
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret != nil {
		if security.HasUsernamePassword(secret) {
			up := UserNamePass{
				UsernamePath: security.SecretPath(o.Secret.Name, constants.ClientUsername),
				PasswordPath: security.SecretPath(o.Secret.Name, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if security.HasTLSCertAndKey(secret) {
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if security.HasCABundle(secret) {
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
		if security.HasBearerTokenFileKey(secret) {
			bt := BearerTokenFile{
				BearerTokenFilePath: security.SecretPath(o.Secret.Name, constants.BearerTokenFileKey),
			}
			conf = append(conf, bt)
		}
	}
	return conf
}
