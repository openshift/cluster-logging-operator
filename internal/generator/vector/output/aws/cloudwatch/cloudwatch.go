package cloudwatch

import (
	_ "embed"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/codec"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	groupNameField                  = "cw_group_name"
	templatedInternalGroupNameField = `{{ _internal.` + groupNameField + ` }}`
)

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	componentID := vectorhelpers.MakeID(id, "normalize_streams")
	tfs = api.Transforms{}
	tfs[componentID] = NormalizeStreamName(inputs)
	groupNameID := vectorhelpers.MakeID(id, "group_name")
	tfs[groupNameID] = commontemplate.NewTemplateRemap([]string{componentID}, o.Cloudwatch.GroupName, groupNameField)
	sink = sinks.NewAwsCloudwatchLogs(func(s *sinks.AwsCloudwatchLogs) {
		s.Region = o.Cloudwatch.Region
		s.Endpoint = o.Cloudwatch.URL
		s.Auth = auth.New(o.Name, o.Cloudwatch.Authentication, op)
		s.GroupName = templatedInternalGroupNameField
		s.StreamName = "{{ stream_name }}"
		s.Encoding = common.NewApiEncoding(codec.CodecTypeJSON)
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

	return id, sink, tfs
}

func NormalizeStreamName(inputs []string) types.Transform {
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
	return transforms.NewRemap(vrl, inputs...)
}
