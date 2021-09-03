package elements

type Record struct {
	Key        string
	Expression string
}

type RecordTransformer struct {
	Records    []Record
	RemoveKeys []string
}

func (rm RecordTransformer) Name() string {
	return "recordTransformerTemplate"
}

func (rm RecordTransformer) Template() string {
	return `{{define "` + rm.Name() + `"  -}}
@type record_transformer
{{if .Records -}}
enable_ruby true
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
