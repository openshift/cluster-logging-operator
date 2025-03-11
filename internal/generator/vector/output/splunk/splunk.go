package splunk

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"regexp"
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

// Match quoted strings like "foo" or "foo/bar-baz"
var quoteRegex = regexp.MustCompile(`^".+"$`)

var indexedFieldsRemap = `
indexed_fields = %s

for_each(indexed_fields) -> |_, field| {
    value = get!(., field) 
    if value != null {
        new_key = replace(join!(field,"_"), r'[\./]', "_")  
        if !is_string(value) {
          if is_object(value) {
            value = encode_json(value)
          } else {
            value, err = to_string(value)
          }
        }
        . =  set!(., [new_key], value)
        . = remove!(., field, true)
    }
}
`

type Splunk struct {
	ComponentID   string
	Inputs        string
	Endpoint      string
	DefaultToken  string
	Index         Element
	IndexedFields string
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
timestamp_key = "@timestamp"
{{if .IndexedFields -}}
indexedFields = {{.IndexedFields}}
{{- end}}
{{end}}
`
}

func (s *Splunk) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}

	inputID := vectorhelpers.MakeID(id, "timestamp")
	splunkIndexID := ""
	var indexTemplate Element
	var indexedFieldsTemplate Element
	if hasIndexKey(o.Splunk) {
		splunkIndexID = vectorhelpers.MakeID(id, "splunk_index")
		indexTemplate = commontemplate.TemplateRemap(splunkIndexID, []string{inputID}, o.Splunk.Index, splunkIndexID, "Splunk Index")
		inputID = splunkIndexID
	}
	if o.Splunk.Tuning != nil && o.Splunk.Tuning.IndexedFields != nil {
		splunkIndexID = vectorhelpers.MakeID(id, "indexed_fields")
		pathSegmentArrayStr, _ := generateQuotedPathSegmentArrayStr(o.Splunk.Tuning.IndexedFields)
		indexedFieldsTemplate = Remap{ComponentID: splunkIndexID, Inputs: vectorhelpers.MakeInputs(inputID), VRL: fmt.Sprintf(indexedFieldsRemap, pathSegmentArrayStr)}
		inputID = splunkIndexID
	}

	splunkSink := sink(id, o, []string{inputID}, splunkIndexID, secrets, op)

	if strategy != nil {
		strategy.VisitSink(splunkSink)
	}
	return []Element{
		FixTimestampFormat(vectorhelpers.MakeID(id, "timestamp"), inputs),
		indexTemplate,
		indexedFieldsTemplate,
		splunkSink,
		common.NewEncoding(id, common.CodecJSON),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, index string, secrets observability.Secrets, op Options) *Splunk {
	s := &Splunk{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Endpoint:    o.Splunk.URL,
		Index:       Tenant(o.Splunk, index),
		RootMixin:   common.NewRootMixin("none"),
	}
	authentication := o.Splunk.Authentication
	if authentication != nil && authentication.Token != nil {
		s.DefaultToken = vectorhelpers.SecretFrom(authentication.Token)
	}
	if o.Splunk.Tuning != nil && o.Splunk.Tuning.IndexedFields != nil {
		_, s.IndexedFields = generateQuotedPathSegmentArrayStr(o.Splunk.Tuning.IndexedFields)
	}
	return s
}

func FixTimestampFormat(componentID string, inputs []string) Element {
	var vrl = `
ts, err = parse_timestamp(.@timestamp,"%+")
if err != null {
	log("could not parse timestamp. err=" + err, rate_limit_secs: 0)
} else {
	.@timestamp = ts
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

// generateQuotedPathSegmentArrayStr generates the final string of the array of array of path segments
// to feed into VRL
func generateQuotedPathSegmentArrayStr(fieldPathArray []obs.FieldPath) (string, string) {
	quotedPathArray := []string{}
	flatted := []string{}
	for _, fieldPath := range fieldPathArray {
		if strings.ContainsAny(string(fieldPath), "/.") {
			flat := strings.ReplaceAll(string(fieldPath), ".", "_")
			flat = strings.ReplaceAll(flat, "\"", "")
			flat = strings.ReplaceAll(flat, "/", "_")
			flatted = append(flatted, fmt.Sprintf("%q", flat))
		}
		f := func(path obs.FieldPath) string {
			splitPathSegments := splitPath(string(path))
			pathArray := quotePathSegments(splitPathSegments)
			return fmt.Sprintf("[%s]", strings.Join(pathArray, ","))
		}
		quotedPathArray = append(quotedPathArray, f(fieldPath))
		//for _, d := range dedottedFields {
		//	label, found := strings.CutPrefix(string(fieldPath), d)
		//	if found && strings.ContainsAny(label, "/.") {
		//		label = strings.ReplaceAll(label, ".", "_")
		//		label = strings.ReplaceAll(label, "/", "_")
		//		quotedPathArray = append(quotedPathArray, f(obs.FieldPath(d+label)))
		//	}
		//}
	}
	return fmt.Sprintf("[%s]", strings.Join(quotedPathArray, ",")),
		fmt.Sprintf("[%s]", strings.Join(flatted, ","))
}

// splitPath splits a fieldPath by `.` and reassembles the quoted path segments that also contain `.`
// Example: `.foo."@some"."d.f.g.o111-22/333".foo_bar`
// Resultant Array: ["foo","@some",`"d.f.g.o111-22/333"`,"foo_bar"]
func splitPath(path string) []string {
	result := []string{}

	splitPath := strings.Split(path, ".")

	var currSegment string
	for _, part := range splitPath {
		if part == "" {
			continue
		} else if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			result = append(result, part)
		} else if strings.HasPrefix(part, `"`) {
			currSegment = part
		} else if strings.HasSuffix(part, `"`) {
			currSegment += "." + part
			result = append(result, currSegment)
			currSegment = ""
		} else if currSegment != "" {
			currSegment += "." + part
		} else {
			result = append(result, part)
		}
	}
	return result
}

// quotePathSegments quotes all path segments as needed for VRL
func quotePathSegments(pathArray []string) []string {
	for i, field := range pathArray {
		// Don't surround in quotes if already quoted
		if quoteRegex.MatchString(field) {
			continue
		}
		// Put quotes around path segments
		pathArray[i] = fmt.Sprintf("%q", field)
	}
	return pathArray
}
