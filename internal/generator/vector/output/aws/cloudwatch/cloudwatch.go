package cloudwatch

import (
	_ "embed"

	"strings"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/remap"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func New(id string, o *observability.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
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
		s.Endpoint = o.Cloudwatch.URL
		s.Auth = auth.New(o.Name, o.Cloudwatch.Authentication, op)
		s.GroupName = "{{ _internal.cw_group_name }}"
		s.StreamName = "{{ stream_name }}"
		s.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
		if o.GetTuning() != nil && o.GetTuning().Compression == "" {
			s.Compression = sinks.CompressionTypeNone
		} else {
			s.Compression = sinks.CompressionType(o.GetTuning().Compression)
		}
		s.Batch = common.NewApiBatch(o)
		s.Buffer = common.NewApiBuffer(o)
		s.Request = common.NewApiRequest(o)
		s.TLS = tls.NewTls(o, secrets, op)
	}, groupNameID)

	return []framework.Element{
		NormalizeStreamName(componentID, inputs),
		commontemplate.NewTemplateRemap(groupNameID, []string{componentID}, o.Cloudwatch.GroupName, groupNameID, "Cloudwatch Groupname"),
		api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = newSink
		}),
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
	return remap.New(componentID, vrl, inputs...)
}
