package loki

import (
	"fmt"
	"slices"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
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
	// DefaultViaqLabels contains the log entry keys that are used as Loki stream labels by default.
	DefaultViaqLabels = []string{
		logType,
	}

	DefaultOtelLabels = []string{
		otellogType,
	}

	viaqContainerLabels = []string{
		lokiLabelKubernetesNamespaceName,
		lokiLabelKubernetesPodName,
		lokiLabelKubernetesContainerName,
	}

	otelContainerLabels = []string{
		otellokiLabelKubernetesNamespaceName,
		otellokiLabelKubernetesPodName,
		otellokiLabelKubernetesContainerName,
	}

	RequiredViaqLabels = []string{
		lokiLabelKubernetesHost,
	}

	RequiredOtelLabels = []string{
		otellokiLabelKubernetesNodeName,
	}

	LokistackContainerLabels = slices.Concat(viaqContainerLabels, otelContainerLabels)

	viaqOtelLabelMap = map[string]string{
		logType:                          otellogType,
		lokiLabelKubernetesNamespaceName: otellokiLabelKubernetesNamespaceName,
		lokiLabelKubernetesPodName:       otellokiLabelKubernetesPodName,
		lokiLabelKubernetesContainerName: otellokiLabelKubernetesContainerName,
		lokiLabelKubernetesHost:          otellokiLabelKubernetesNodeName,
	}
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

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	componentID := vectorhelpers.MakeID(id, "remap")
	elements := []Element{
		CleanupFields(componentID, inputs),
	}

	// Add remap labels based on input type
	shouldRemapLabels := false
	// Used to determine if OTel labels are needed for lokistack outputs
	isLokistackOutput := false

	// InputTypeOption is set for Lokistack outputs
	if input, ok := op[constants.InputTypeOption]; ok {
		// Only add remapLabels for application or infrastructure for Lokistack outputs
		if input == string(obs.InputTypeApplication) || input == string(obs.InputTypeInfrastructure) {
			shouldRemapLabels = true
		}
		// Always add OTel labels for Lokistack outputs
		isLokistackOutput = true
	} else {
		// For regular Loki outputs, always add remapLabels
		shouldRemapLabels = true
	}

	if shouldRemapLabels {
		remapLabelID := vectorhelpers.MakeID(id, "remap_label")
		remapLabels := RemapLabels(remapLabelID, o, []string{componentID})
		elements = append(elements, remapLabels)
		componentID = remapLabelID
	}

	// Add tenant template if tenant key is configured
	var lokiTenantID string
	if hasTenantKey(o.Loki) {
		lokiTenantID = vectorhelpers.MakeID(id, "loki_tenant")
		tenantTemplate := commontemplate.TemplateRemap(lokiTenantID, []string{componentID}, o.Loki.TenantKey, lokiTenantID, "Loki Tenant")
		elements = append(elements, tenantTemplate)
		componentID = lokiTenantID
	}

	sink := Output(id, o, []string{componentID}, lokiTenantID)

	if strategy != nil {
		strategy.VisitSink(sink)
	}

	// Add remaining elements
	elements = append(elements,
		sink,
		common.NewEncoding(id, common.CodecJSON),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		NewLabels(id, o, isLokistackOutput),
		tls.New(id, o.TLS, secrets, op),
		auth.HTTPAuth(id, o.Loki.Authentication, secrets, op),
	)

	return elements
}

func Output(id string, o obs.OutputSpec, inputs []string, tenant string) *Loki {
	return &Loki{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Endpoint:    o.Loki.URLSpec.URL,
		TenantID:    Tenant(o.Loki, tenant),
		RootMixin:   common.NewRootMixin(nil),
	}
}

func lokiLabelKeys(l *obs.Loki, isLokistackOutput bool) []string {
	keys := sets.NewString(RequiredViaqLabels...)
	if l != nil && len(l.LabelKeys) != 0 {
		keys.Insert(l.LabelKeys...)
		// Determine which of the OTel labels need to also be added based on spec'd custom labels
		// Only required for Lokistack outputs
		if isLokistackOutput {
			keys.Insert(addOtelEquivalentLabels(l.LabelKeys)...)
		}
	} else {
		// Add default labels and container labels if none specified for regular Loki outputs
		keys.Insert(DefaultViaqLabels...).
			Insert(viaqContainerLabels...)
	}
	return keys.List()
}

func lokiLabels(lo *obs.Loki, isLokistackOutput bool) []LokiLabel {
	ls := []LokiLabel{}
	for _, k := range lokiLabelKeys(lo, isLokistackOutput) {
		r := strings.NewReplacer(".", "_", "/", "_", "\\", "_", "-", "_")
		name := r.Replace(k)
		l := LokiLabel{
			Name:  name,
			Value: formatLokiLabelValue(k),
		}
		// some labels need custom values. e.g. host, otel labels
		if val := generateCustomLabelValues(k); val != "" {
			l.Value = val
		}

		ls = append(ls, l)
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

func remapContainerLabelsVrl(labels []string) string {
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

func RemapLabels(id string, o obs.OutputSpec, inputs []string) Element {
	return Remap{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         remapContainerLabelsVrl(viaqContainerLabels),
	}
}

func NewLabels(id string, o obs.OutputSpec, isLokistackOutput bool) Element {
	return LokiLabels{
		ComponentID: id,
		Labels:      lokiLabels(o.Loki, isLokistackOutput),
	}
}

func hasTenantKey(l *obs.Loki) bool {
	return l != nil && l.TenantKey != ""
}

func Tenant(l *obs.Loki, tenant string) Element {
	if !hasTenantKey(l) {
		return Nil
	}
	return KV("tenant_id", fmt.Sprintf(`"{{ _internal.%s }}"`, tenant))
}

func CleanupFields(id string, inputs []string) Element {
	return Remap{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         "del(.tag)",
	}
}
