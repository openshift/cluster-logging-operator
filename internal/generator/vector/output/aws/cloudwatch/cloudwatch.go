package cloudwatch

import (
	_ "embed"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/aws"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/aws"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
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
	EndpointConfig framework.Element
	AuthConfig     framework.Element
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
group_name = "{{"{{"}} _internal.{{.GroupName}} {{"}}"}}"
stream_name = "{{"{{ stream_name }}"}}"
{{compose_one .AuthConfig}}
healthcheck.enabled = false
{{compose_one .EndpointConfig}}
{{- end}}
`
}

func (e *CloudWatch) SetCompression(algo string) {
	e.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, adapter observability.TunableOutput, op utils.Options) []Element {
	componentID := vectorhelpers.MakeID(id, "normalize_streams")
	groupNameID := vectorhelpers.MakeID(id, "group_name")
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			NormalizeStreamName(componentID, inputs),
			elements.Debug(id, vectorhelpers.MakeInputs([]string{componentID}...)),
		}
	}

	newSink := sinks.NewAwsCloudwatchLogs(func(s *sinks.AwsCloudwatchLogs) {
		s.Region = o.Cloudwatch.Region
		s.GroupName = "{{ _internal.cw_group_name }}"
		s.StreamName = "{{ stream_name }}"
		s.Endpoint = o.Cloudwatch.URL
		s.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
		if adapter.GetTuning().Compression == "" {
			s.Compression = sinks.CompressionTypeNone
		} else {
			s.Compression = sinks.CompressionType(adapter.GetTuning().Compression)
		}
		s.Batch = common.NewApiBatch(adapter)
		s.Buffer = common.NewApiBuffer(adapter)
		s.Request = common.NewApiRequest(adapter)
		s.TLS = tls.NewTls(o.TLS, secrets, op)
		s.Auth = aws.NewAuthConfig(o.Name, o.Cloudwatch.Authentication, op)
	}, groupNameID)

	return []framework.Element{
		NormalizeStreamName(componentID, inputs),
		commontemplate.TemplateRemap(groupNameID, []string{componentID}, o.Cloudwatch.GroupName, groupNameID, "Cloudwatch Groupname"),
		api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = newSink
		}),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, op utils.Options, region, groupName string) *CloudWatch {
	return &CloudWatch{
		Desc:           "Cloudwatch Logs",
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		Region:         region,
		GroupName:      groupName,
		AuthConfig:     aws.AuthConfig(o.Name, o.Cloudwatch.Authentication, op, secrets),
		EndpointConfig: endpointConfig(o.Cloudwatch),
		RootMixin:      common.NewRootMixin("none"),
	}
}

func endpointConfig(cw *obs.Cloudwatch) framework.Element {
	if cw == nil {
		return Endpoint{}
	}
	return Endpoint{
		URL: cw.URL,
	}
}

func NormalizeStreamName(componentID string, inputs []string) framework.Element {
	vrl := strings.TrimSpace(`
.stream_name = "default"
if ( .log_type == "audit" ) {
 .stream_name = (.hostname +"."+ downcase(.log_source)) ?? .stream_name
}
if ( .log_source == "container" ) {
  k = .kubernetes
  .stream_name = (k.namespace_name+"_"+k.pod_name+"_"+k.container_name) ?? .stream_name
}
if ( .log_type == "infrastructure" ) {
 .stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
}
if ( .log_source == "node" ) {
 .stream_name =  ( .hostname + ".journal.system" ) ?? .stream_name
}
del(.tag)
del(.source_type)
	`)
	return elements.Remap{
		Desc:        "Cloudwatch Stream Names",
		ComponentID: componentID,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
}
