package gcl

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
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
log_id = "{{.LogID}}"
severity_key = "{{.SeverityKey}}"

[sinks.{{.ComponentID}}.resource]
type = "k8s_node"
node_name = "{{"{{hostname}}"}}"
{{end}}`
}

func (g *GoogleCloudLogging) SetCompression(algo string) {
	g.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets helpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	if o.GoogleCloudLogging == nil {
		return []Element{}
	}
	g := o.GoogleCloudLogging
	gcl := &GoogleCloudLogging{
		ComponentID:     id,
		Inputs:          helpers.MakeInputs(inputs...),
		LogDestination:  LogDestination(g),
		LogID:           g.LogID,
		SeverityKey:     SeverityKey(g),
		CredentialsPath: auth(g.Authentication, secrets),
		RootMixin:       common.NewRootMixin(nil),
	}
	if strategy != nil {
		strategy.VisitSink(gcl)
	}
	return []Element{
		gcl,
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func auth(spec *obs.GoogleCloudLoggingAuthentication, secrets helpers.Secrets) string {
	if spec == nil {
		return ""
	}
	return secrets.Path(spec.Credentials)
}

// LogDestination is one of BillingAccountID, OrganizationID, FolderID, or ProjectID in that order
func LogDestination(g *obs.GoogleCloudLogging) Element {
	if g.BillingAccountID != "" {
		return KV(BillingAccountID, fmt.Sprintf("%q", g.BillingAccountID))
	}
	if g.OrganizationID != "" {
		return KV(OrganizationID, fmt.Sprintf("%q", g.OrganizationID))
	}
	if g.FolderID != "" {
		return KV(FolderID, fmt.Sprintf("%q", g.FolderID))
	}
	if g.ProjectID != "" {
		return KV(ProjectID, fmt.Sprintf("%q", g.ProjectID))
	}
	return Nil
}

func SeverityKey(g *obs.GoogleCloudLogging) string {
	return DefaultSeverityKey
}
