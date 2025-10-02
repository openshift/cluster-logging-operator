package cloudwatch

import (
	_ "embed"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
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
	EndpointConfig Element
	AuthConfig     Element
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

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	componentID := vectorhelpers.MakeID(id, "normalize_streams")
	groupNameID := vectorhelpers.MakeID(id, "group_name")
	if genhelper.IsDebugOutput(op) {
		return []Element{
			NormalizeStreamName(componentID, inputs),
			Debug(id, vectorhelpers.MakeInputs([]string{componentID}...)),
		}
	}
	cwSink := sink(id, o, []string{groupNameID}, secrets, op, o.Cloudwatch.Region, groupNameID)
	if strategy != nil {
		strategy.VisitSink(cwSink)
	}

	return []Element{
		NormalizeStreamName(componentID, inputs),
		commontemplate.TemplateRemap(groupNameID, []string{componentID}, o.Cloudwatch.GroupName, groupNameID, "Cloudwatch Groupname"),
		cwSink,
		common.NewEncoding(id, common.CodecJSON),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, op Options, region, groupName string) *CloudWatch {
	return &CloudWatch{
		Desc:           "Cloudwatch Logs",
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		Region:         region,
		GroupName:      groupName,
		AuthConfig:     authConfig(o.Name, o.Cloudwatch.Authentication, op, secrets),
		EndpointConfig: endpointConfig(o.Cloudwatch),
		RootMixin:      common.NewRootMixin("none"),
	}
}

// authConfig returns the templated VRL containing cloudwatch auth configuration
func authConfig(outputName string, auth *obs.CloudwatchAuthentication, options Options, secrets observability.Secrets) Element {
	authConfig := NewAuth()
	if auth == nil {
		return authConfig
	}
	switch auth.Type {
	case obs.CloudwatchAuthTypeAccessKey:
		authConfig.KeyID.Value = vectorhelpers.SecretFrom(&auth.AWSAccessKey.KeyId)
		authConfig.KeySecret.Value = vectorhelpers.SecretFrom(&auth.AWSAccessKey.KeySecret)
		// New assumeRole works with static keys as well
		if auth.AssumeRole != nil {
			authConfig.AssumeRole.Value = vectorhelpers.SecretFrom(&auth.AssumeRole.RoleARN)
			// Optional externalID
			if hasExtID, extID := AssumeRoleHasExternalId(auth.AssumeRole); hasExtID {
				authConfig.ExternalID.Value = extID
			}
		}
	case obs.CloudwatchAuthTypeIAMRole:
		if forwarderName, found := utils.GetOption(options, OptionForwarderName, ""); found {
			// For OIDC roles we mount a configMap containing a credentials file
			authConfig.CredentialsPath.Value = strings.Trim(vectorhelpers.ConfigPath(forwarderName+"-"+constants.AWSCredentialsConfigMapName, constants.AWSCredentialsKey), `"`)
			authConfig.Profile.Value = "output_" + outputName
		}
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
	return Remap{
		Desc:        "Cloudwatch Stream Names",
		ComponentID: componentID,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
}
