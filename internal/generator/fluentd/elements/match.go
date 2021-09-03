package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type Match struct {
	Desc         string
	MatchTags    string
	MatchElement generator.Element
}

func (m Match) Name() string {
	return "matchTemplate"
}

func (m Match) Template() string {
	return `{{define "` + m.Name() + `"  -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
<match {{.MatchTags}}>
{{compose_one .MatchElement | indent 2}}
</match>
{{end}}`
}
