package gcl

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
)

const (
	OrganizationID   = "organization_id"
	ProjectID        = "project_id"
	BillingAccountID = "billing_account_id"
	FolderID         = "folder_id"

	DefaultSeverityKey = "level"

	GoogleApplicationCredentialsKey = "google-application-credentials.json"
)

type GoogleCloudLogging struct {
	Desc        string
	ComponentID string
	Inputs      string

	LogDestination Element

	LogID       string
	SeverityKey string

	CredentialsPath string
	common.RootMixin
}

func (g GoogleCloudLogging) Name() string {
	return "vectorGCL"
}

func (g GoogleCloudLogging) Template() string {
	return `{{define "` + g.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "gcp_stackdriver_logs"
inputs = {{.Inputs}}
{{kv .LogDestination -}}
credentials_path = {{.CredentialsPath}}
log_id = "{{"{{"}} _internal.{{.LogID}} {{"}}"}}"
severity_key = "{{.SeverityKey}}"

[sinks.{{.ComponentID}}.resource]
type = "k8s_node"
node_name = "{{"{{hostname}}"}}"
{{end}}`
}

func (g *GoogleCloudLogging) SetCompression(algo string) {
	g.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	if o.GoogleCloudLogging == nil {
		return []Element{}
	}
	componentID := vectorhelpers.MakeID(id, "log_id")
	gclSeverityID := vectorhelpers.MakeID(id, "normalize_severity")
	g := o.GoogleCloudLogging
	gcl := &GoogleCloudLogging{
		ComponentID:     id,
		Inputs:          helpers.MakeInputs(gclSeverityID),
		LogDestination:  LogDestination(g),
		LogID:           componentID,
		SeverityKey:     SeverityKey(g),
		CredentialsPath: auth(g.Authentication, secrets),
		RootMixin:       common.NewRootMixin(nil),
	}

	if strategy != nil {
		strategy.VisitSink(gcl)
	}
	return []Element{
		commontemplate.TemplateRemap(componentID, inputs, o.GoogleCloudLogging.LogId, componentID, "GoogleCloudLogging LogId"),
		NormalizeSeverity(gclSeverityID, []string{componentID}),
		gcl,
		common.NewEncoding(id, ""),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func auth(spec *obs.GoogleCloudLoggingAuthentication, secrets observability.Secrets) string {
	if spec == nil {
		return ""
	}
	return secrets.Path(spec.Credentials)
}

// LogDestination is one of BillingAccountID, OrganizationID, FolderID, or ProjectID in that order
func LogDestination(g *obs.GoogleCloudLogging) Element {
	var key string
	switch g.ID.Type {
	case obs.GoogleCloudLoggingIdTypeFolder:
		key = FolderID
	case obs.GoogleCloudLoggingIdTypeProject:
		key = ProjectID
	case obs.GoogleCloudLoggingIdTypeBillingAccount:
		key = BillingAccountID
	case obs.GoogleCloudLoggingIdTypeOrganization:
		key = OrganizationID
	}
	return KV(key, fmt.Sprintf("%q", g.ID.Value))
}

func SeverityKey(g *obs.GoogleCloudLogging) string {
	return DefaultSeverityKey
}

// NormalizeSeverity normalizes log severity to conform to GCL's standard
// Accepted Severity: DEFAULT, EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG
// Ref: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
func NormalizeSeverity(componentID string, inputs []string) Element {
	var vrl = `
# Set audit log level to 'INFO'
if .log_type == "audit" {
	.level = "INFO"
} else if !exists(.level) {
  	.level = "DEFAULT"
} else if .level == "warn" {
	.level = "WARNING"
} else if .level == "trace" {
	.level = "DEBUG"
} else {
	.level = upcase!(.level) 
}
`
	return Remap{
		Desc:        "Normalize GCL severity levels",
		ComponentID: componentID,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
}
