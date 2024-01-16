package gcl

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"

	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	corev1 "k8s.io/api/core/v1"
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

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	id := vectorhelpers.FormatComponentID(o.Name)
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	if o.GoogleCloudLogging == nil {
		return []Element{}
	}
	g := o.GoogleCloudLogging
	dedottedID := normalize.ID(id, "dedot")
	gcl := GoogleCloudLogging{
		ComponentID:     id,
		Inputs:          helpers.MakeInputs(inputs...),
		LogDestination:  LogDestination(g),
		LogID:           g.LogID,
		SeverityKey:     SeverityKey(g),
		CredentialsPath: security.SecretPath(o.Secret.Name, GoogleApplicationCredentialsKey),
	}
	setInput(&gcl, []string{dedottedID})
	return MergeElements(
		[]Element{
			normalize.DedotLabels(dedottedID, inputs),
			gcl,
			output.NewBuffer(id),
			output.NewRequest(id),
		},
		security.TLS(o, secret, op),
	)
}

func setInput(gcl *GoogleCloudLogging, inputs []string) Element {
	gcl.Inputs = helpers.MakeInputs(inputs...)
	return gcl
}

// LogDestination is one of BillingAccountID, OrganizationID, FolderID, or ProjectID in that order
func LogDestination(g *logging.GoogleCloudLogging) Element {
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

func SeverityKey(g *logging.GoogleCloudLogging) string {
	return DefaultSeverityKey
}
