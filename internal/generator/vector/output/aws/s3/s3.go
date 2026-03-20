package s3

import (
	_ "embed"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/codec"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	keyPrefixID := vectorhelpers.MakeID(id, "key_prefix")
	tfs = api.Transforms{}
	tfs[keyPrefixID] = template.NewTemplateRemap(inputs, o.S3.KeyPrefix, keyPrefixID)

	sink = sinks.NewAwsS3(func(s *sinks.AwsS3) {
		s.Region = o.S3.Region
		s.Bucket = o.S3.Bucket
		s.KeyPrefix = fmt.Sprintf("{{ _internal.%s }}", keyPrefixID)
		s.Endpoint = o.S3.URL
		s.Auth = auth.New(o.Name, o.S3.Authentication, op)
		s.Encoding = common.NewApiEncoding(codec.CodecTypeJSON)
		s.Batch = common.NewApiBatch(o)
		s.Compression = sinks.CompressionType(o.GetTuning().Compression)
		s.Buffer = common.NewApiBuffer(o)
		s.Request = common.NewApiRequest(o)
		s.TLS = tls.NewTls(o, secrets, op)
	}, keyPrefixID)

	return id, sink, tfs
}
