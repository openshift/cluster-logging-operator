package elements

type RecordModifier struct {
	Records    []Record
	RemoveKeys []string
}

func (rm RecordModifier) Name() string {
	return "recordModifierTemplate"
}

func (rm RecordModifier) Template() string {
	return `{{define "` + rm.Name() + `"  -}}
@type record_modifier
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
