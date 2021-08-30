package generator

type ConfLiteral struct {
	ComponentID
	TemplateName string
	Desc         string
	InLabel
	OutLabel
	Pattern     string
	TemplateStr string
}

func (b ConfLiteral) Name() string {
	return b.TemplateName
}

func (b ConfLiteral) Template() string {
	return b.TemplateStr
}

func Comment(c string) Element {
	return ConfLiteral{
		Desc:         c,
		TemplateName: "comment",
		TemplateStr: `{{define "comment" -}}
# {{.Desc}}
{{- end}}`,
	}
}
