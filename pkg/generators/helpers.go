package generators

import (
	"bytes"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	funcMap = template.FuncMap{
		"indent":  indent,
		"include": include,
	}
)

//IncludeTemplate allows root templates to include partials by reusing
//the root template object.  Callers can provide hints for use in temporarily changing
//behavior of the included template
type IncludeTemplate interface {
	Template() *template.Template
	SetHints(hints []string)
	Hints() sets.String
}

//include a named template using the given binding and hints
func include(name string, context IncludeTemplate, hints ...string) (string, error) {
	origHints := context.Hints()
	context.SetHints(hints)
	var out bytes.Buffer
	if err := context.Template().ExecuteTemplate(&out, name, context); err != nil {
		context.SetHints(origHints.List())
		return "", err
	}
	context.SetHints(origHints.List())
	return out.String(), nil
}

//indent helper function to prefix each line of the output by N spaces
func indent(length int, in string) string {
	pad := strings.Repeat(" ", length)
	return pad + strings.Replace(in, "\n", "\n"+pad, -1)
}
