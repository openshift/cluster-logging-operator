package output

type Buffer struct {
	ComponentID string
	WhenFull    string
}

func NewBuffer(id string) Buffer {
	return Buffer{
		ComponentID: id,
		WhenFull:    "drop_newest",
	}
}

func (b Buffer) Name() string {
	return "buffer"
}

func (b Buffer) Template() string {
	return `{{define "` + b.Name() + `" -}}
[sinks.{{.ComponentID}}.buffer]
when_full = "{{.WhenFull}}"
{{end}}`
}
