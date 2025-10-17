package s3

import (
	_ "embed"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/aws"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type Endpoint struct {
	URL string
}

func (e Endpoint) Name() string {
	return "awsS3EndpointTemplate"
}

func (e Endpoint) Template() (ret string) {
	ret = `{{define "` + e.Name() + `" -}}`
	if e.URL != "" {
		ret += `endpoint = "{{ .URL }}"`
	}
	ret += `{{end}}`
	return
}

type Compression struct {
	Algorithm string
}

func (c Compression) Name() string {
	return "awsS3CompressionTemplate"
}

func (c Compression) Template() (ret string) {
	ret = `{{define "` + c.Name() + `" -}}`
	if c.Algorithm != "" && c.Algorithm != "none" {
		ret += `compression = "{{ .Algorithm }}"`
	}
	ret += `{{end}}`
	return
}

type S3 struct {
	Desc           string
	ComponentID    string
	Inputs         string
	Region         string
	Bucket         string
	KeyPrefix      string
	Compression    Element
	EndpointConfig Element
	SecurityConfig Element
	common.RootMixin
}

func (e S3) Name() string {
	return "s3Template"
}

func (e S3) Template() string {
	return `{{define "` + e.Name() + `" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[sinks.{{.ComponentID}}]
type = "aws_s3"
inputs = {{.Inputs}}
region = "{{.Region}}"
{{compose_one .Compression}}
bucket = "{{.Bucket}}"
key_prefix = "{{"{{"}} _internal.{{.KeyPrefix}} {{"}}"}}"
{{compose_one .SecurityConfig}}
healthcheck.enabled = false
{{compose_one .EndpointConfig}}
{{- end}}
`
}

func (s *S3) SetCompression(algo string) {
	s.Compression = Compression{
		Algorithm: algo,
	}
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	keyPrefixID := vectorhelpers.MakeID(id, "key_prefix")
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	sink := sink(id, o, []string{keyPrefixID}, secrets, op, o.S3.Region, o.S3.Bucket, keyPrefixID)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	return []Element{
		template.TemplateRemap(keyPrefixID, inputs, o.S3.KeyPrefix, keyPrefixID, "S3 Key Prefix"),
		sink,
		common.NewEncoding(id, common.CodecJSON),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, op Options, region, bucket, keyPrefix string) *S3 {
	return &S3{
		Desc:           "Amazon S3",
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		Region:         region,
		Bucket:         bucket,
		KeyPrefix:      keyPrefix,
		Compression:    compressionConfig(o.S3),
		SecurityConfig: aws.AuthConfig(o.Name, o.S3.Authentication, op, secrets),
		EndpointConfig: endpointConfig(o.S3),
		RootMixin:      common.NewRootMixin("none"),
	}
}

func endpointConfig(s3 *obs.S3) Element {
	if s3 == nil {
		return Endpoint{}
	}
	return Endpoint{
		URL: s3.URL,
	}
}

func compressionConfig(s3 *obs.S3) Element {
	if s3 == nil || s3.Tuning == nil || s3.Tuning.Compression == "" {
		return Compression{}
	}
	return Compression{
		Algorithm: s3.Tuning.Compression,
	}
}
