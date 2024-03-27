package loki

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"strings"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	corev1 "k8s.io/api/core/v1"
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

	containerLabels = []string{
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
	common.RootMixin
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
{{.Compression}}
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

func (e *Loki) SetCompression(algo string) {
	e.Compression.Value = algo
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	componentID := vectorhelpers.MakeID(id, "remap")
	dedottedID := vectorhelpers.MakeID(id, "dedot")
	remapLabelID := vectorhelpers.MakeID(id, "remap_label")
	sink := Output(id, o, []string{remapLabelID})
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(
		[]Element{
			CleanupFields(componentID, inputs),
			viaq.DedotLabels(dedottedID, []string{componentID}),
			RemapLabels(remapLabelID, o, []string{dedottedID}),
			sink,
			Encoding(id, o),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
			common.NewRequest(id, strategy),
			Labels(id, o),
		},
		TLSConf(id, o, secret, op),
		BasicAuth(id, o, secret),
		BearerTokenAuth(id, o, secret),
	)
}

func Output(id string, o logging.OutputSpec, inputs []string) *Loki {
	return &Loki{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Endpoint:    o.URL,
		TenantID:    Tenant(o.Loki),
		RootMixin:   common.NewRootMixin(nil),
	}
}

func Encoding(id string, o logging.OutputSpec) Element {
	return LokiEncoding{
		ComponentID: id,
		Codec:       lokiEncodingJson,
	}
}

func lokiLabelKeys(l *logging.Loki) []string {
	var keys sets.String
	if l != nil && len(l.LabelKeys) != 0 {
		keys = *sets.NewString(l.LabelKeys...)
	} else {
		keys = *sets.NewString(defaultLabelKeys...)
	}
	// Ensure required tags for serialization
	keys.Insert(requiredLabelKeys...)
	return keys.List()
}

func lokiLabels(lo *logging.Loki) []Label {
	ls := []Label{}
	for _, k := range lokiLabelKeys(lo) {
		r := strings.NewReplacer(".", "_", "/", "_", "\\", "_", "-", "_")
		name := r.Replace(k)
		l := Label{
			Name:  name,
			Value: formatLokiLabelValue(k),
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

func remapLabelsVrl(labels []string) string {
	k8sEventLabel := `
if !exists(.%s) {
  .%s = ""
}`
	sb := strings.Builder{}
	for _, v := range labels {
		sb.WriteString(fmt.Sprintf(k8sEventLabel, v, v))
	}
	return sb.String()
}

func formatLokiLabelValue(value string) string {
	if strings.HasPrefix(value, "kubernetes.labels.") || strings.HasPrefix(value, "kubernetes.namespace_labels.") {
		parts := strings.SplitAfterN(value, "labels.", 2)
		r := strings.NewReplacer("/", "_", ".", "_")
		key := r.Replace(parts[1])
		key = fmt.Sprintf(`\"%s\"`, key)
		value = fmt.Sprintf("%s%s", parts[0], key)
	}
	return fmt.Sprintf("{{%s}}", value)
}

func RemapLabels(id string, o logging.OutputSpec, inputs []string) Element {
	return Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         remapLabelsVrl(containerLabels),
	}
}

func Labels(id string, o logging.OutputSpec) Element {
	return LokiLabels{
		ComponentID: id,
		Labels:      lokiLabels(o.Loki),
	}
}

func Tenant(l *logging.Loki) Element {
	if l == nil || l.TenantKey == "" {
		return Nil
	}
	return KV("tenant_id", fmt.Sprintf("%q", fmt.Sprintf("{{%s}}", l.TenantKey)))
}

func TLSConf(id string, o logging.OutputSpec, secret *corev1.Secret, op Options) []Element {
	conf := []Element{}
	if isDefaultOutput(o.Name) {
		// Set CA from logcollector ServiceAccount for internal Loki
		tlsConf := common.TLSConf{
			ComponentID: id,
			CAFilePath:  `"/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"`,
		}
		tlsConf.SetTLSProfileFromOptions(op)
		return append(conf, tlsConf)
	}
	if o.Secret != nil || (o.TLS != nil && o.TLS.InsecureSkipVerify) {

		if tlsConf := common.GenerateTLSConfWithID(id, o, secret, op, false); tlsConf != nil {
			tlsConf.NeedsEnabled = false
			conf = append(conf, tlsConf)
		}

	} else if secret != nil {
		// Use secret of logcollector service account as backup
		tlsConf := common.TLSConf{
			ComponentID: id,
			CAFilePath:  `"/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"`,
		}
		tlsConf.SetTLSProfileFromOptions(op)
		conf = append(conf, tlsConf)
	}

	return conf
}

func isDefaultOutput(name string) bool {
	return strings.HasPrefix(name, "default-")
}

func BasicAuth(id string, o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: id,
		})
		if common.HasUsernamePassword(secret) {
			hasBasicAuth = true
			up := UserNamePass{
				Username: common.GetFromSecret(secret, constants.ClientUsername),
				Password: common.GetFromSecret(secret, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}

func BearerTokenAuth(id string, o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if secret != nil {
		// Inject token from secret, either provided by user using a custom secret
		// or from the default logcollector service account.
		if common.HasBearerTokenFileKey(secret) {
			conf = append(conf, BasicAuthConf{
				Desc:        "Bearer Auth Config",
				ComponentID: id,
			}, BearerToken{
				Token: common.GetFromSecret(secret, constants.BearerTokenFileKey),
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
