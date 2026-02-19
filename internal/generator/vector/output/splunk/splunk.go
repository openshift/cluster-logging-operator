package splunk

import (
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/remap"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// VRL template for detecting default Splunk 'source' field by .log_type and .log_source
var sourceTmpl = `
# Splunk 'source' field detection
if ._internal.log_type == "infrastructure" && ._internal.log_source == "node" {
    ._internal.splunk.source = to_string!(._internal.systemd.u.SYSLOG_IDENTIFIER || "")
}
if ._internal.log_source == "container" {
   	._internal.splunk.source = join!([._internal.kubernetes.namespace_name, ._internal.kubernetes.pod_name, ._internal.kubernetes.container_name], "_")
}
if ._internal.log_type == "audit" {
   ._internal.splunk.source = ._internal.log_source
}
`

// VRL template to set payload for log event and detecting 'sourcetype'.
// If payload is object will be set 'sourcetype':"_json",
// 'sourcetype': "generic_single_line" otherwise
var payloadKeyTmpl = `
payloadKey = %s
if !is_null(payloadKey) {
	value = get!(., payloadKey) 
	if !is_null(value) {
        internal = ._internal
        . = {}
        . = set!(., payloadKey, value)
        ._internal = internal
		if is_object(value) {
			._internal.splunk.sourcetype = "_json"
		} else {
			._internal.splunk.sourcetype = "generic_single_line"
		}
	} else {
		._internal.splunk.sourcetype = "_json"
	}
}
`

// VRL template to proceed indexed fields:
// - the nested field convert to root-level, original path remove from object
// - "." and "/" replaced with "_"
var indexedFieldsRemap = `
# Splunk indexed fields
indexed_fields = %s
for_each(indexed_fields) -> |_, field| {
	value = get!(., field) 
	if !is_null(value) {
		new_key = replace(join!(field,"_"), r'[\./]', "_")  
		if !is_string(value) {
		  if is_object(value) {
			value = encode_json(value)
		  } else {
			value = to_string!(value)
		  }
		}
        . = remove!(., field, true)
		. = set!(., [new_key], value)	
	} else {
        log("Path " + join!(field, ".") + " not found in log event", level: "warn")
    }
}
`

<<<<<<< HEAD
<<<<<<< HEAD
type Splunk struct {
	ComponentID   string
	Inputs        string
	Endpoint      string
	DefaultToken  string
	Index         framework.Element
	IndexedFields framework.Element
	Source        framework.Element
	SourceType    framework.Element
	HostKey       framework.Element
	common.RootMixin
}

func (s Splunk) Name() string {
	return "SplunkVectorTemplate"
}

func (s Splunk) Template() string {
	return `{{define "` + s.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "splunk_hec_logs"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
{{.Compression}}
default_token = "{{.DefaultToken}}"
{{kv .Index -}}
timestamp_key = "._internal.timestamp"
{{kv .IndexedFields}}
{{kv .Source -}}
{{kv .SourceType -}}
{{kv .HostKey -}}
{{end}}
`
}

func (s *Splunk) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o *observability.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	inputID := vectorhelpers.MakeID(id, "timestamp")

	var indexTemplate framework.Element
	splunkIndexID := ""
	if hasIndexKey(o.Splunk) {
		splunkIndexID = vectorhelpers.MakeID(id, "splunk_index")
		indexTemplate = commontemplate.TemplateRemap(splunkIndexID, []string{inputID}, o.Splunk.Index, splunkIndexID, "Splunk Index")
		inputID = splunkIndexID
	}

	var builder strings.Builder
	if o.Splunk.Source != "" {
		builder.WriteString(fmt.Sprintf("._internal.splunk.source = %s", commontemplate.TransformUserTemplateToVRL(o.Splunk.Source)))
	} else {
		builder.WriteString(sourceTmpl)
	}

	builder.WriteString("\n._internal.splunk.sourcetype = \"_json\"\n")

	if o.Splunk.PayloadKey != "" {
		path := vectorhelpers.SplitPath(string(o.Splunk.PayloadKey))
		quotedSegments := vectorhelpers.QuotePathSegments(path)
		quotedPathArray := fmt.Sprintf("[%s]", strings.Join(quotedSegments, ","))
		builder.WriteString(fmt.Sprintf(payloadKeyTmpl, quotedPathArray))
	}

	var indexedFields []string
	if o.Splunk.IndexedFields != nil {
		pathSegmentArrayStr, remapped := vectorhelpers.GenerateQuotedPathSegmentArrayStr(o.Splunk.IndexedFields)
		builder.WriteString(fmt.Sprintf(indexedFieldsRemap, pathSegmentArrayStr))
		indexedFields = remapped
	}
	splunkMetadataID := vectorhelpers.MakeID(id, "metadata")
	metadata := remap.New(splunkMetadataID, builder.String(), inputID)

	inputID = splunkMetadataID

	return []framework.Element{
		fixTimestampFormat(vectorhelpers.MakeID(id, "timestamp"), inputs),
		indexTemplate,
		metadata,
		api.NewConfig(func(config *api.Config) {
			config.Sinks[id] = sinks.NewSplunkHecLogs(o.Splunk.URL, func(s *sinks.SplunkHecLogs) {
				s.Index = tenant(o.Splunk, splunkIndexID)
				s.DefaultToken = defaultToken(o.Splunk)
				s.Compression = sinks.CompressionType(o.GetTuning().Compression)
				s.Source = "{{ ._internal.splunk.source }}"
				s.SourceType = "{{ ._internal.splunk.sourcetype }}"
				s.HostKey = "._internal.hostname"
				s.TimestampKey = "._internal.timestamp"
				s.IndexedFields = indexedFields
				s.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
				s.Batch = common.NewApiBatch(o)
				s.Buffer = common.NewApiBuffer(o)
				s.Request = common.NewApiRequest(o)
				s.TLS = tls.NewTls(o, secrets, op)
			}, inputID)
		}),
	}
}

func defaultToken(o *obs.Splunk) string {
	authentication := o.Authentication
	if authentication != nil && authentication.Token != nil {
		return vectorhelpers.SecretFrom(authentication.Token)
	}
	return ""
}

func fixTimestampFormat(componentID string, inputs []string) framework.Element {
	var vrl = `
ts, err = parse_timestamp(._internal.timestamp,"%+")
if err != null {
	log("could not parse timestamp. err=" + err, rate_limit_secs: 0)
} else {
	._internal.timestamp = ts
}
`
	return remap.New(componentID, vrl, inputs...)
}

func hasIndexKey(s *obs.Splunk) bool {
	return s != nil && s.Index != ""
}

func tenant(s *obs.Splunk, index string) string {
	if !hasIndexKey(s) {
		return ""
	}
	return fmt.Sprintf("{{ ._internal.%s }}", index)
}
