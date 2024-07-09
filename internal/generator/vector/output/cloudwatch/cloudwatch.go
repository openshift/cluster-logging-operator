package cloudwatch

import (
	"regexp"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
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
	GroupName      string
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
group_name = "{{.GroupName}}"
stream_name = "{{"{{ stream_name }}"}}"
{{compose_one .SecurityConfig}}
healthcheck.enabled = false
{{compose_one .EndpointConfig}}
{{- end}}
`
}

func (e *CloudWatch) SetCompression(algo string) {
	e.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	componentID := vectorhelpers.MakeID(id, "normalize_streams")
	dedottedID := vectorhelpers.MakeID(id, "dedot")
	if genhelper.IsDebugOutput(op) {
		return []Element{
			NormalizeStreamName(componentID, inputs),
			Debug(id, vectorhelpers.MakeInputs([]string{componentID}...)),
		}
	}
	sink := sink(id, o, []string{dedottedID}, secrets, op, o.Cloudwatch.Region, o.Cloudwatch.GroupName)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	return []Element{
		NormalizeStreamName(componentID, inputs),
		viaq.DedotLabels(dedottedID, []string{componentID}),
		sink,
		common.NewEncoding(id, common.CodecJSON),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, op Options, region, groupName string) *CloudWatch {
	return &CloudWatch{
		Desc:           "Cloudwatch Logs",
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		Region:         region,
		GroupName:      groupName,
		SecurityConfig: authConfig(o.Cloudwatch.Authentication, secrets),
		EndpointConfig: endpointConfig(o.Cloudwatch),
		RootMixin:      common.NewRootMixin("none"),
	}
}

func authConfig(auth *obs.CloudwatchAuthentication, secrets vectorhelpers.Secrets) Element {
	authConfig := NewAuth()
	if auth != nil && auth.Type == obs.CloudwatchAuthTypeAccessKey {
		authConfig.KeyID.Value = vectorhelpers.SecretFrom(auth.AWSAccessKey.KeyID)
		authConfig.KeySecret.Value = vectorhelpers.SecretFrom(auth.AWSAccessKey.KeySecret)
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

func NormalizeStreamName(componentID string, inputs []string) Element {
	vrl := strings.TrimSpace(`
.stream_name = "default"

if (.file != null) {
 .file = "kubernetes" + replace!(.file, "/", ".")
 .stream_name = del(.file)
}

if ( .log_type == "audit" ) {
 .stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
}
if ( .log_type == "infrastructure" ) {
 .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
}
if ( .tag == ".journal.system" ) {
 .stream_name =  ( .hostname + .tag ) ?? .stream_name
}
del(.tag)
del(.source_type)
	`)
	return Remap{
		Desc:        "Cloudwatch Stream Names",
		ComponentID: componentID,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
}

// ParseRoleArn search for matching valid ARN
func ParseRoleArn(auth *obs.CloudwatchAuthentication, secrets vectorhelpers.Secrets) string {
	if auth.Type == obs.CloudwatchAuthTypeIAMRole {
		roleArnString := secrets.AsString(auth.IAMRole.RoleARN)

		if roleArnString != "" {
			reg := regexp.MustCompile(`(arn:aws(.*)?:(iam|sts)::\d{12}:role\/\S+)\s?`)
			roleArn := reg.FindStringSubmatch(roleArnString)
			if roleArn != nil {
				return roleArn[1] // the capturing group is index 1
			}
		}
	}
	return ""
}
