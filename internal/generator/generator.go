package generator

import (
	"bytes"
	"strings"
	"text/template"
)

// Element is a basic unit of configuration. It wraps a golang template along with the data type to hold the data template needs. A type implementing
type Element interface {
	Name() string
	Template() string
}

// Section is a collection of Elements at a high level division of configuration. e.g. Inputs, Outputs, Ingress etc. It is used to show a breakdown of generated configuration + adding a comment along with the declared section to document the meaning of the section.
type Section struct {
	Elements []Element
	Comment  string
}

// InLabel defines the name of the <label> section in fluentd configuration
type InLabel = string

// OutLabel defined the name of the next @label in the fluentd configuration
type OutLabel = string

// ComponentID is used to define a component in vector configuration, a source, sink, etc
type ComponentID = string

// Generator converts an array of Elements to configuration. It is intentionally kept stateless.
type Generator int

// MakeGenerator creates Generator
func MakeGenerator() Generator {
	return Generator(0)
}

// GenerateConf converts array of Element into a configuration.
func (g Generator) GenerateConf(es ...Element) (string, error) {
	conf, err := g.generate(es)
	if err != nil {
		return conf, err
	}
	return strings.TrimSpace(conf), nil
}

func (g Generator) generate(es []Element) (string, error) {
	if len(es) == 0 {
		return "", nil
	}
	t := template.New("generate")
	f := template.FuncMap{
		"compose": g.generate,
		"compose_one": func(e Element) (string, error) {
			return g.generate([]Element{e})
		},
		"kv": func(e Element) (string, error) {
			s, err := g.generate([]Element{e})
			if strings.TrimSpace(s) == "" {
				return s, err
			}
			return s + "\n", err
		},
		"indent": indent,
		"comma_separated": func(arr []string) string {
			return strings.Join(arr, ", ")
		},
	}
	f["optional"] = f["kv"]
	t.Funcs(f)
	b := &bytes.Buffer{}
	for i, e := range es {
		if e == nil || e == Nil {
			continue
		}
		t = template.Must(t.Parse(e.Template()))
		err := t.ExecuteTemplate(b, e.Name(), e)
		if err != nil {
			return "error in conf generation", err
		}
		if i < len(es)-1 {
			b.Write([]byte("\n"))
		}
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

// MergeElements merges multiple arrays of Elements into a single array of Element
func MergeElements(els ...[]Element) []Element {
	merged := make([]Element, 0)
	for _, el := range els {
		merged = append(merged, el...)
	}
	return merged
}

// MergeSections merges multiple Sections into a single array of Element
func MergeSections(sections []Section) []Element {
	merged := make([]Element, 0)
	for _, s := range sections {
		merged = append(merged, s.Elements...)
	}
	return merged
}

// indent helper function to prefix each line of the output by N spaces
func indent(length int, in string) string {
	if len(in) == 0 {
		return ""
	}
	pad := strings.Repeat(" ", length)
	inlines := strings.Split(in, "\n")
	outlines := make([]string, len(inlines))
	for i, inline := range inlines {
		// empty lines not indented
		// if strings.TrimSpace(inline) == "" {
		// 	outlines[i] = ""
		// } else {
		// 	outlines[i] = pad + inline
		// }
		// empty lines indented
		outlines[i] = pad + inline
	}
	return strings.Join(outlines, "\n")
}
