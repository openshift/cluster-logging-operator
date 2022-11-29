package loki

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	logType                          = "log_type"
	lokiLabelKubernetesNamespaceName = "kubernetes.namespace_name"
	lokiLabelKubernetesPodName       = "kubernetes.pod_name"
	lokiLabelKubernetesHost          = "kubernetes.host"
	lokiLabelKubernetesContainerName = "kubernetes.container_name"
	podNamespace                     = "kubernetes.namespace_name"
)

var (
	defaultLabelKeys = []string{
		logType,

		//container labels
		lokiLabelKubernetesNamespaceName,
		lokiLabelKubernetesPodName,
		lokiLabelKubernetesContainerName,
	}
	requiredLabelKeys = []string{
		lokiLabelKubernetesHost,
	}
	lokiEncodingJson = fmt.Sprintf("%q", "json")
)

type Loki struct {
	ComponentID string
	Inputs      string
	TenantID    Element
	Endpoint    string
	LokiLabel   []string
}

func (l Loki) Name() string {
	return "lokiVectorTemplate"
}

func (l Loki) Template() string {
	return `{{define "` + l.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "loki"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
out_of_order_action = "accept"
healthcheck.enabled = false
{{kv .TenantID -}}
{{end}}`
}

type LokiEncoding struct {
	ComponentID string
	Codec       string
}

func (le LokiEncoding) Name() string {
	return "lokiEncoding"
}

func (le LokiEncoding) Template() string {
	return `{{define "` + le.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = {{.Codec}}
{{end}}`
}

type Label struct {
	Name  string
	Value string
}

type LokiLabels struct {
	ComponentID string
	Labels      []Label
}

func (l LokiLabels) Name() string {
	return "lokiLabels"
}

func (l LokiLabels) Template() string {
	return `{{define "` + l.Name() + `" -}}
[sinks.{{.ComponentID}}.labels]
{{range $i, $label := .Labels -}}
{{$label.Name}} = "{{$label.Value}}"
{{end -}}
{{end}}
`
}

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	outputName := helpers.FormatComponentID(o.Name)
	componentID := fmt.Sprintf("%s_%s", outputName, "remap")
	return MergeElements(
		[]Element{
			CleanupFields(componentID, inputs),
			Output(o, []string{componentID}),
			Encoding(o),
			Labels(o),
		},
		TLSConf(o, secret),
		BasicAuth(o, secret),
		BearerTokenAuth(o, secret),
	)
}

func Output(o logging.OutputSpec, inputs []string) Element {
	return Loki{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Endpoint:    o.URL,
		TenantID:    Tenant(o.Loki),
	}
}

func Encoding(o logging.OutputSpec) Element {
	return LokiEncoding{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Codec:       lokiEncodingJson,
	}
}

func lokiLabelKeys(l *logging.Loki) []string {
	var keys sets.String
	if l != nil && len(l.LabelKeys) != 0 {
		keys = sets.NewString(l.LabelKeys...)
	} else {
		keys = sets.NewString(defaultLabelKeys...)
	}
	// Ensure required tags for serialization
	keys.Insert(requiredLabelKeys...)
	return keys.List()
}

func lokiLabels(lo *logging.Loki) []Label {
	ls := []Label{}
	for _, k := range lokiLabelKeys(lo) {
		l := Label{
			Name:  strings.ReplaceAll(k, ".", "_"),
			Value: fmt.Sprintf("{{%s}}", k),
		}
		if k == lokiLabelKubernetesHost {
			l.Value = "${VECTOR_SELF_NODE_NAME}"
		}
		if k == lokiLabelKubernetesNamespaceName {
			l.Value = fmt.Sprintf("{{%s}}", podNamespace)
		}
		ls = append(ls, l)
	}
	return ls
}

func Labels(o logging.OutputSpec) Element {
	return LokiLabels{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Labels:      lokiLabels(o.Loki),
	}
}

func Tenant(l *logging.Loki) Element {
	if l == nil || l.TenantKey == "" {
		return Nil
	}
	return KV("tenant_id", fmt.Sprintf("%q", fmt.Sprintf("{{%s}}", l.TenantKey)))
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret != nil {
		hasTLS := false
		conf = append(conf, security.TLSConf{
			ComponentID:        strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
			InsecureSkipVerify: o.TLS != nil && o.TLS.InsecureSkipVerify,
		})

		if o.Name == logging.OutputNameDefault || security.HasTLSCertAndKey(secret) {
			hasTLS = true
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if o.Name == logging.OutputNameDefault || security.HasCABundle(secret) {
			hasTLS = true
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
		if !hasTLS {
			return []Element{}
		}
	} else if secret != nil {
		// Set CA from logcollector ServiceAccount for internal Loki
		return []Element{
			security.TLSConf{
				ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
			},
			CAFile{
				CAFilePath: `"/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"`,
			},
		}
	}
	return conf
}

func BasicAuth(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		})
		if security.HasUsernamePassword(secret) {
			hasBasicAuth = true
			up := UserNamePass{
				Username: security.GetFromSecret(secret, constants.ClientUsername),
				Password: security.GetFromSecret(secret, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}

func BearerTokenAuth(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if secret != nil {
		// Inject token from secret, either provided by user using a custom secret
		// or from the default logcollector service account.
		if security.HasBearerTokenFileKey(secret) {
			conf = append(conf, BasicAuthConf{
				Desc:        "Bearer Auth Config",
				ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
			}, BearerToken{
				Token: security.GetFromSecret(secret, constants.BearerTokenFileKey),
			})
		}
	}
	return conf
}

func CleanupFields(id string, inputs []string) Element {
	return Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         "del(.tag)",
	}
}
