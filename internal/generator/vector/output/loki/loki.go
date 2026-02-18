package loki

import (
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/remap"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

const (
	logType                          = "log_type"
	lokiLabelKubernetesNamespaceName = "kubernetes.namespace_name"
	lokiLabelKubernetesPodName       = "kubernetes.pod_name"
	lokiLabelKubernetesHost          = "kubernetes.host"
	lokiLabelKubernetesContainerName = "kubernetes.container_name"
	podNamespace                     = "kubernetes.namespace_name"

	// OTel
	otellogType                          = "openshift.log_type"
	otellokiLabelKubernetesNamespaceName = "k8s.namespace_name"
	otellokiLabelKubernetesPodName       = "k8s.pod_name"
	otellokiLabelKubernetesContainerName = "k8s.container_name"
	otellokiLabelKubernetesNodeName      = "k8s.node_name"
)

var (
	// DefaultLabelKeys contains the log entry keys that are used as Loki stream labels by default.
	DefaultLabelKeys = []string{
		logType,

		//container labels
		lokiLabelKubernetesNamespaceName,
		lokiLabelKubernetesPodName,
		lokiLabelKubernetesContainerName,

		// OTel labels
		otellogType,
		otellokiLabelKubernetesNamespaceName,
		otellokiLabelKubernetesPodName,
		otellokiLabelKubernetesContainerName,
	}

	containerLabels = []string{
		lokiLabelKubernetesNamespaceName,
		lokiLabelKubernetesPodName,
		lokiLabelKubernetesContainerName,
	}

	requiredLabelKeys = []string{
		otellokiLabelKubernetesNodeName,
		lokiLabelKubernetesHost,
	}

	viaqOtelLabelMap = map[string]string{
		logType:                          otellogType,
		lokiLabelKubernetesNamespaceName: otellokiLabelKubernetesNamespaceName,
		lokiLabelKubernetesPodName:       otellokiLabelKubernetesPodName,
		lokiLabelKubernetesContainerName: otellokiLabelKubernetesContainerName,
	}
)

<<<<<<< HEAD
type Loki struct {
	ComponentID string
	Inputs      string
	TenantID    framework.Element
	Endpoint    string
	Proxy       string
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
{{with .Proxy -}}
proxy.enabled = true
proxy.http = "{{.}}"
proxy.https = "{{.}}"
{{end -}}
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

type LokiLabel struct {
	Name  string
	Value string
}

type LokiLabels struct {
	ComponentID string
	Labels      []LokiLabel
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

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	componentID := vectorhelpers.MakeID(id, "remap")
	remapLabelID := vectorhelpers.MakeID(id, "remap_label")

	var tenantTemplate framework.Element
	sink := sinks.NewLoki(o.Loki.URL, func(s *sinks.Loki) {
		s.OutOfOrderAction = sinks.LokiOutOfOrderActionAccept
		s.Auth = auth.NewHttpAuth(o.Loki.Authentication, op)
		s.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
		s.Compression = sinks.CompressionType(o.GetTuning().Compression)
		s.Batch = common.NewApiBatch(o)
		s.Buffer = common.NewApiBuffer(o)
		s.Request = common.NewApiRequest(o)
		s.TLS = tls.NewTls(o, secrets, op)
		if o.Loki.ProxyURL != "" {
			s.Proxy = &sinks.Proxy{
				Enabled: true,
				Http:    o.Loki.ProxyURL,
				Https:   o.Loki.ProxyURL,
			}
		}
	}, remapLabelID)
	if hasTenantKey(o.Loki) {
		lokiTenantID := vectorhelpers.MakeID(id, "loki_tenant")
		tenantTemplate = commontemplate.TemplateRemap(lokiTenantID, []string{remapLabelID}, o.Loki.TenantKey, lokiTenantID, "Loki Tenant")
		sink.Inputs = []string{lokiTenantID}
		sink.TenantId = tenantId(o.Loki, lokiTenantID)
	}
	sink.Labels = lokiLabels(o.Loki)

	return []framework.Element{
		CleanupFields(componentID, inputs),
		RemapLabels(remapLabelID, []string{componentID}),
		tenantTemplate,
		api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = sink
		}),
	}
}

func lokiLabelKeys(l *obs.Loki) []string {
	var keys sets.String
	if l != nil && len(l.LabelKeys) != 0 {
		keys = *sets.NewString(l.LabelKeys...)
		// Determine which of the OTel labels need to also be added based on spec'd custom labels
		keys.Insert(addOtelEquivalentLabels(l.LabelKeys)...)
	} else {
		keys = *sets.NewString(DefaultLabelKeys...)
	}
	// Ensure required tags for serialization
	keys.Insert(requiredLabelKeys...)
	return keys.List()
}

func lokiLabels(lo *obs.Loki) map[string]string {
	ls := map[string]string{}
	for _, k := range lokiLabelKeys(lo) {
		r := strings.NewReplacer(".", "_", "/", "_", "\\", "_", "-", "_")
		name := r.Replace(k)
		ls[name] = formatLokiLabelValue(k)
		// some labels need custom values. e.g. host, otel labels
		if val := generateCustomLabelValues(k); val != "" {
			ls[name] = val
		}
	}
	return ls
}

// addOtelEquivalentLabels checks spec'd custom label keys to add matching otel labels
// e.g kubernetes.namespace_name = k8s.namespace_name
func addOtelEquivalentLabels(customLabelKeys []string) []string {
	matchingLabels := []string{}

	for _, label := range customLabelKeys {
		if val, ok := viaqOtelLabelMap[label]; ok {
			matchingLabels = append(matchingLabels, val)
		}
	}
	return matchingLabels
}

// generateCustomLabelValues generates custom values for specific labels like kubernetes.host, k8s_* labels
func generateCustomLabelValues(value string) string {
	var labelVal string

	switch value {
	case otellogType:
		labelVal = logType
	case otellokiLabelKubernetesContainerName:
		labelVal = lokiLabelKubernetesContainerName
	case lokiLabelKubernetesNamespaceName, otellokiLabelKubernetesNamespaceName:
		labelVal = podNamespace
	case otellokiLabelKubernetesPodName:
		labelVal = lokiLabelKubernetesPodName
	// Special case for the kubernetes node name (same as kubernetes.host)
	case lokiLabelKubernetesHost, otellokiLabelKubernetesNodeName:
		return "${VECTOR_SELF_NODE_NAME}"
	default:
		return ""
	}
	return fmt.Sprintf("{{%s}}", labelVal)
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
		key = fmt.Sprintf("%q", key)
		value = fmt.Sprintf("%s%s", parts[0], key)
	}
	return fmt.Sprintf("{{%s}}", value)
}

func RemapLabels(id string, inputs []string) framework.Element {
	return remap.New(id, remapLabelsVrl(containerLabels), inputs...)
}

func hasTenantKey(l *obs.Loki) bool {
	return l != nil && l.TenantKey != ""
}

func tenantId(l *obs.Loki, tenant string) string {
	if !hasTenantKey(l) {
		return ""
	}
	return fmt.Sprintf("{{ _internal.%s }}", tenant)
}

func CleanupFields(id string, inputs []string) framework.Element {
	return remap.New(id, "del(.tag)", inputs...)
}
