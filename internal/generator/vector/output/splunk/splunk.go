package splunk

import (
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/codec"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
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

// VRL template to set payload for log event and sourcetype.
// If payload is key is invalid, sourcetype will fall back to '_json'
var payloadKeysourceTypeTmpl = `
payloadKey = %s
sourceType = %s
if !is_null(payloadKey) {
	value = get!(., payloadKey) 
	if !is_null(value) {
        internal = ._internal
        . = {}
        . = set!(., payloadKey, value)
        ._internal = internal
	    ._internal.splunk.sourcetype = sourceType
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

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	inputID := vectorhelpers.MakeID(id, "timestamp")
	tfs = api.Transforms{}
	splunkIndexID := ""
	if hasIndexKey(o.Splunk) {
		splunkIndexID = vectorhelpers.MakeID(id, "splunk_index")
		tfs[splunkIndexID] = commontemplate.NewTemplateRemap([]string{inputID}, o.Splunk.Index, splunkIndexID)
		inputID = splunkIndexID
	}

	var builder strings.Builder
	if o.Splunk.Source != "" {
		builder.WriteString(fmt.Sprintf("._internal.splunk.source = %s", commontemplate.TransformUserTemplateToVRL(o.Splunk.Source)))
	} else {
		builder.WriteString(sourceTmpl)
	}

	if o.Splunk.PayloadKey != "" {
		path := vectorhelpers.SplitPath(string(o.Splunk.PayloadKey))
		quotedSegments := vectorhelpers.QuotePathSegments(path)
		quotedPathArray := fmt.Sprintf("[%s]", strings.Join(quotedSegments, ","))
		if o.Splunk.SourceType != "" {
			builder.WriteString(fmt.Sprintf(payloadKeysourceTypeTmpl, quotedPathArray, commontemplate.TransformUserTemplateToVRL(o.Splunk.SourceType)))
		} else {
			builder.WriteString(fmt.Sprintf(payloadKeyTmpl, quotedPathArray))
		}
	} else {
		builder.WriteString("\n._internal.splunk.sourcetype = \"_json\"\n")
	}

	var indexedFields []string
	if o.Splunk.IndexedFields != nil {
		pathSegmentArrayStr, remapped := vectorhelpers.GenerateQuotedPathSegmentArrayStr(o.Splunk.IndexedFields)
		builder.WriteString(fmt.Sprintf(indexedFieldsRemap, pathSegmentArrayStr))
		indexedFields = remapped
	}
	splunkMetadataID := vectorhelpers.MakeID(id, "metadata")
	tfs[splunkMetadataID] = transforms.NewRemap(builder.String(), inputID)
	inputID = splunkMetadataID

	tfs[vectorhelpers.MakeID(id, "timestamp")] = fixTimestampFormat(inputs)
	sink = sinks.NewSplunkHecLogs(o.Splunk.URL, func(s *sinks.SplunkHecLogs) {
		s.Index = tenant(o.Splunk, splunkIndexID)
		s.DefaultToken = defaultToken(o.Splunk)
		s.Compression = sinks.CompressionType(o.GetTuning().Compression)
		s.Source = "{{ ._internal.splunk.source }}"
		s.SourceType = "{{ ._internal.splunk.sourcetype }}"
		s.HostKey = "._internal.hostname"
		s.TimestampKey = "._internal.timestamp"
		s.IndexedFields = indexedFields
		s.Encoding = common.NewApiEncoding(codec.CodecTypeJSON)
		s.Batch = common.NewApiBatch(o)
		s.Buffer = common.NewApiBuffer(o)
		s.Request = common.NewApiRequest(o)
		s.TLS = tls.NewTls(o, secrets, op)
	}, inputID)
	return id, sink, tfs
}

func defaultToken(o *obs.Splunk) string {
	authentication := o.Authentication
	if authentication != nil && authentication.Token != nil {
		return vectorhelpers.SecretFrom(authentication.Token)
	}
	return ""
}

func fixTimestampFormat(inputs []string) types.Transform {
	var vrl = `
ts, err = parse_timestamp(._internal.timestamp,"%+")
if err != null {
	log("could not parse timestamp. err=" + err, rate_limit_secs: 0)
} else {
	._internal.timestamp = ts
}
`
	return transforms.NewRemap(vrl, inputs...)
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
