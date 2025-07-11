package splunk

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
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

type Splunk struct {
	ComponentID   string
	Inputs        string
	Endpoint      string
	DefaultToken  string
	Index         Element
	IndexedFields Element
	Source        Element
	SourceType    Element
	HostKey       Element
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

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op utils.Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	inputID := vectorhelpers.MakeID(id, "timestamp")

	var indexTemplate Element
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

	if o.Splunk.IndexedFields != nil {
		pathSegmentArrayStr, remapped := vectorhelpers.GenerateQuotedPathSegmentArrayStr(o.Splunk.IndexedFields)
		builder.WriteString(fmt.Sprintf(indexedFieldsRemap, pathSegmentArrayStr))
		op["indexed_fields"] = remapped
	}
	splunkMetadataID := vectorhelpers.MakeID(id, "metadata")
	metadata := Remap{ComponentID: splunkMetadataID,
		Inputs: vectorhelpers.MakeInputs(inputID),
		VRL:    builder.String()}

	inputID = splunkMetadataID
	splunkSink := sink(id, o, []string{inputID}, splunkIndexID, secrets, op)

	if strategy != nil {
		strategy.VisitSink(splunkSink)
	}
	return []Element{
		FixTimestampFormat(vectorhelpers.MakeID(id, "timestamp"), inputs),
		indexTemplate,
		metadata,
		splunkSink,
		common.NewEncoding(id, common.CodecJSON),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, index string, secrets observability.Secrets, op utils.Options) *Splunk {
	s := &Splunk{
		ComponentID:   id,
		Inputs:        vectorhelpers.MakeInputs(inputs...),
		Endpoint:      o.Splunk.URL,
		Index:         Tenant(o.Splunk, index),
		RootMixin:     common.NewRootMixin("none"),
		Source:        KV("source", `"{{ ._internal.splunk.source }}"`),
		SourceType:    KV("sourcetype", `"{{ ._internal.splunk.sourcetype }}"`),
		HostKey:       KV("host_key", `"._internal.hostname"`),
		IndexedFields: IndexedFields(o.Splunk, op),
	}
	authentication := o.Splunk.Authentication
	if authentication != nil && authentication.Token != nil {
		s.DefaultToken = vectorhelpers.SecretFrom(authentication.Token)
	}
	return s
}

func FixTimestampFormat(componentID string, inputs []string) Element {
	var vrl = `
ts, err = parse_timestamp(._internal.timestamp,"%+")
if err != null {
	log("could not parse timestamp. err=" + err, rate_limit_secs: 0)
} else {
	._internal.timestamp = ts
}
`
	return Remap{
		Desc:        "Ensure timestamp field well formatted for Splunk",
		ComponentID: componentID,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
}

func hasIndexKey(s *obs.Splunk) bool {
	return s != nil && s.Index != ""
}

func Tenant(s *obs.Splunk, index string) Element {
	if !hasIndexKey(s) {
		return Nil
	}
	return KV("index", fmt.Sprintf(`"{{ ._internal.%s }}"`, index))
}

func IndexedFields(s *obs.Splunk, op utils.Options) Element {
	if s.IndexedFields != nil && op.Has("indexed_fields") {
		in, _ := utils.GetOption[string](op, "indexed_fields", "")
		return KV("indexed_fields", in)
	}
	return Nil
}
