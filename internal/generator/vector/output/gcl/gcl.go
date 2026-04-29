package gcl

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	DefaultSeverityKey              = "level"
	GoogleApplicationCredentialsKey = "google-application-credentials.json"
)

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	tfs = api.Transforms{}
	if o.GoogleCloudLogging == nil {
		return "", nil, nil
	}
	componentID := helpers.MakeID(id, "log_id")
	gclSeverityID := helpers.MakeID(id, "normalize_severity")
	tfs[gclSeverityID] = NormalizeSeverity(componentID)
	tfs[componentID] = commontemplate.NewTemplateRemap(inputs, o.GoogleCloudLogging.LogId, componentID)
	g := o.GoogleCloudLogging
	sink = sinks.NewGcpStackdriverLogs(func(s *sinks.GcpStackdriverLogs) {
		LogDestination(s, o.GoogleCloudLogging)
		s.LogId = fmt.Sprintf("{{ _internal.%s }}", componentID)
		s.SeverityKey = DefaultSeverityKey
		s.CredentialsPath = auth(g.Authentication)
		s.Encoding = common.NewApiEncoding("")
		s.Batch = common.NewApiBatch(o)
		s.Buffer = common.NewApiBuffer(o)
		s.Request = common.NewApiRequest(o)
		s.TLS = tls.NewTls(o, secrets, op)
		s.Resource = map[string]string{
			"type":      "k8s_node",
			"node_name": "{{hostname}}",
		}
	}, gclSeverityID)
	return id, sink, tfs
}

func auth(spec *obs.GoogleCloudLoggingAuthentication) string {
	if spec == nil || spec.Credentials == nil {
		return ""
	}
	return helpers.SecretPath(spec.Credentials.SecretName, spec.Credentials.Key, "%s")
}

// LogDestination is one of BillingAccountID, OrganizationID, FolderID, or ProjectID in that order
func LogDestination(sink *sinks.GcpStackdriverLogs, g *obs.GoogleCloudLogging) {
	value := g.ID.Value
	switch g.ID.Type {
	case obs.GoogleCloudLoggingIdTypeFolder:
		sink.FolderId = value
	case obs.GoogleCloudLoggingIdTypeProject:
		sink.ProjectId = value
	case obs.GoogleCloudLoggingIdTypeBillingAccount:
		sink.BillingAccountId = value
	case obs.GoogleCloudLoggingIdTypeOrganization:
		sink.OrganizationId = value
	}
}

// NormalizeSeverity normalizes log severity to conform to GCL's standard
// Accepted Severity: DEFAULT, EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG
// Ref: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
func NormalizeSeverity(inputs ...string) types.Transform {
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
	return transforms.NewRemap(vrl, inputs...)
}
