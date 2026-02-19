package s3

import (
	_ "embed"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func New(id string, o *observability.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	keyPrefixID := vectorhelpers.MakeID(id, "key_prefix")
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}

	newSink := sinks.NewAwsS3(func(s *sinks.AwsS3) {
		s.Region = o.S3.Region
		s.Bucket = o.S3.Bucket
		s.KeyPrefix = fmt.Sprintf("{{ _internal.%s }}", keyPrefixID)
		s.Endpoint = o.S3.URL
		s.Auth = auth.New(o.Name, o.S3.Authentication, op)
		s.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
		s.Batch = common.NewApiBatch(o)
		s.Compression = sinks.CompressionType(o.GetTuning().Compression)
		s.Buffer = common.NewApiBuffer(o)
		s.Request = common.NewApiRequest(o)
		s.TLS = tls.NewTls(o, secrets, op)
	}, keyPrefixID)

	return []framework.Element{
		template.TemplateRemap(keyPrefixID, inputs, o.S3.KeyPrefix, keyPrefixID, "S3 Key Prefix"),
		api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = newSink
		}),
	}
}
