package generator

type NilElement int

func (r NilElement) Name() string {
	return "nilElement"
}

func (r NilElement) Template() string {
	return `{{define "` + r.Name() + `"}}{{end -}}`
}

var Nil NilElement
