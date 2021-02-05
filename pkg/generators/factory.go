package generators

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

//Generator is a template engine
type Generator struct {
	*template.Template
}

//New creates an instance of a template engine for a set of templates
func New(name string, addFunctions *template.FuncMap, templates ...string) (*Generator, error) {
	tmpl := template.New(name).Funcs(*addFunctions).Funcs(funcMap)
	var err error
	for i, s := range templates {
		tmpl, err = tmpl.Parse(s)
		if err != nil {
			// Include the first line containing the template name in the error message.
			templateName := "unknown"
			lines := strings.SplitN(s, "\n", 2)
			if len(lines) > 0 {
				templateName = lines[0]
			}
			return nil, fmt.Errorf("Error parsing %v template %v %q: %s", name, i, templateName, err)
		}
	}
	return &Generator{tmpl}, nil
}

//Execute the named template using the given data
func (gen *Generator) Execute(namedTemplate string, data interface{}) (string, error) {
	var out bytes.Buffer
	err := gen.Template.ExecuteTemplate(&out, namedTemplate, data)
	return out.String(), err
}
