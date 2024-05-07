package cloudwatch

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"regexp"
	"strings"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type Endpoint struct {
	URL string
}

func (e Endpoint) Name() string {
	return "awsEndpointTemplate"
}

func (e Endpoint) Template() (ret string) {
	ret = `{{define "` + e.Name() + `" -}}`
	if e.URL != "" {
		ret += `endpoint = "{{ .URL }}"`
	}
	ret += `{{end}}`
	return
}

type CloudWatch struct {
	Desc           string
	ComponentID    string
	Inputs         string
	Region         string
	EndpointConfig Element
	SecurityConfig Element
	common.RootMixin
}

func (e CloudWatch) Name() string {
	return "cloudwatchTemplate"
}

func (e CloudWatch) Template() string {

	return `{{define "` + e.Name() + `" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[sinks.{{.ComponentID}}]
type = "aws_cloudwatch_logs"
inputs = {{.Inputs}}
region = "{{.Region}}"
{{.Compression}}
group_name = "{{"{{ group_name }}"}}"
stream_name = "{{"{{ stream_name }}"}}"
{{compose_one .SecurityConfig}}
encoding.codec = "json"
healthcheck.enabled = false
{{compose_one .EndpointConfig}}
{{- end}}
`
}

func (e *CloudWatch) SetCompression(algo string) {
	e.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	componentID := helpers.MakeID(id, "normalize_group_and_streams")
	dedottedID := helpers.MakeID(id, "dedot")
	if genhelper.IsDebugOutput(op) {
		return []Element{
			NormalizeGroupAndStreamName(LogGroupNameField(o), LogGroupPrefix(o), componentID, inputs),
			Debug(id, helpers.MakeInputs([]string{componentID}...)),
		}
	}
	sink := sink(id, o, []string{dedottedID}, secrets, op, o.Cloudwatch.Region)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	return []Element{
		NormalizeGroupAndStreamName(LogGroupNameField(o), LogGroupPrefix(o), componentID, inputs),
		viaq.DedotLabels(dedottedID, []string{componentID}),
		sink,
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, op Options, region string) *CloudWatch {
	return &CloudWatch{
		Desc:           "Cloudwatch Logs",
		ComponentID:    id,
		Inputs:         helpers.MakeInputs(inputs...),
		Region:         region,
		SecurityConfig: authConfig(o.Cloudwatch.Authentication, secrets),
		EndpointConfig: endpointConfig(o.Cloudwatch),
		RootMixin:      common.NewRootMixin("none"),
	}
}

func authConfig(auth *obs.CloudwatchAuthentication, secrets vectorhelpers.Secrets) Element {
	authConfig := NewAuth()
	if auth != nil {
		if auth.RoleARN != nil || auth.Credentials != nil {
			return authConfig
		}
		// Otherwise use ID and Secret
		authConfig.KeyID.Value = strings.TrimSpace(secrets.AsString(auth.AccessKeyID))
		authConfig.KeySecret.Value = strings.TrimSpace(secrets.AsString(auth.AccessKeySecret))
	}
	return authConfig
}

func endpointConfig(cw *obs.Cloudwatch) Element {
	if cw == nil {
		return Endpoint{}
	}
	return Endpoint{
		URL: cw.URL,
	}
}

func NormalizeGroupAndStreamName(logGroupNameField string, logGroupPrefix string, componentID string, inputs []string) Element {
	appGroupName := fmt.Sprintf("%q + %s", logGroupPrefix, logGroupNameField)
	if logGroupPrefix == "" {
		appGroupName = logGroupNameField
	}
	auditGroupName := fmt.Sprintf("%s%s", logGroupPrefix, "audit")
	infraGroupName := fmt.Sprintf("%s%s", logGroupPrefix, "infrastructure")
	vrl := strings.TrimSpace(`
.group_name = "default"
.stream_name = "default"

if (.file != null) {
 .file = "kubernetes" + replace!(.file, "/", ".")
 .stream_name = del(.file)
}

if ( .log_type == "application" ) {
 .group_name = ( ` + appGroupName + ` ) ?? "application"
}
if ( .log_type == "audit" ) {
 .group_name = "` + auditGroupName + `"
 .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
}
if ( .log_type == "infrastructure" ) {
 .group_name = "` + infraGroupName + `"
 .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
}
if ( .tag == ".journal.system" ) {
 .stream_name =  ( .hostname + .tag ) ?? .stream_name
}
del(.tag)
del(.source_type)
	`)
	return Remap{
		Desc:        "Cloudwatch Group and Stream Names",
		ComponentID: componentID,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
}

func LogGroupPrefix(o obs.OutputSpec) string {
	if o.Cloudwatch != nil {
		prefix := o.Cloudwatch.GroupPrefix
		if strings.TrimSpace(prefix) != "" {
			return fmt.Sprintf("%s.", prefix)
		}
	}
	return ""
}

// LogGroupNameField Return the field used for grouping the application logs
func LogGroupNameField(o obs.OutputSpec) string {
	if o.Cloudwatch != nil {
		switch o.Cloudwatch.GroupBy {
		case obs.LogGroupByNamespaceName:
			return ".kubernetes.namespace_name"
		case obs.LogGroupByNamespaceUUID:
			return ".kubernetes.namespace_id"
		default:
			return ".log_type"
		}
	}
	return ""
}

// ParseRoleArn search for matching valid arn within the 'credentials' or 'role_arn' key
func ParseRoleArn(auth *obs.CloudwatchAuthentication, secrets vectorhelpers.Secrets) string {
	roleArnString := secrets.AsString(auth.Credentials)
	if roleArnString == "" {
		roleArnString = secrets.AsString(auth.RoleARN)
	}

	if roleArnString != "" {
		reg := regexp.MustCompile(`(arn:aws(.*)?:(iam|sts)::\d{12}:role\/\S+)\s?`)
		roleArn := reg.FindStringSubmatch(roleArnString)
		if roleArn != nil {
			return roleArn[1] // the capturing group is index 1
		}
	}
	return ""
}
