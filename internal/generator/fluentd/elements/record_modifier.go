package elements

// Fluentd including some plugins treats logs as a BINARY by default to forward.
// But sometimes storage can't proceed some chars, eg ElasticSearch.
// Plugin record_modifier can help resolve such problem.
// In char_encoding from:to case, it replaces invalid character with safe character.
// <filter pattern>
//   @type record_modifier
//   # change char encoding from 'ASCII-8BIT' to 'UTF-8'
//   char_encoding ascii-8bit:utf-8
//  </filter>
// */

const DefaultCharEncoding = "ascii-8bit:utf-8"
const CharEncoding = "charEncoding"

type RecordModifier struct {
	Records      []Record
	RemoveKeys   []string
	CharEncoding string
}

func (rm RecordModifier) Name() string {
	return "recordModifierTemplate"
}

func (rm RecordModifier) Template() string {
	return `{{define "` + rm.Name() + `"  -}}
@type record_modifier
{{if .CharEncoding -}}
char_encoding {{.CharEncoding}}
{{end -}}
{{if .Records -}}
<record>
{{- range $Index, $Record := .Records}}
  {{$Record.Key}} {{$Record.Expression}}
{{- end}}
</record>
{{end -}}
{{if .RemoveKeys -}}
remove_keys {{comma_separated .RemoveKeys}}
{{end -}}
{{end}}
`
}
