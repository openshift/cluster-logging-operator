package source

type JournalLogs struct {
	SourceID     string
	Desc         string
	SourceType   string
	ExcludePaths string
}

func (jl JournalLogs) ComponentID() string {
	return jl.SourceID
}

func (jl JournalLogs) Type() string {
	return jl.SourceType
}

func (jl JournalLogs) Name() string {
	return "journald"
}

func (jl JournalLogs) Template() string {
	return `{{define "` + jl.Name() + `" -}}
# {{.Desc}}
[sources.{{.SourceType}}]
  type = "{{.SourceType}}"
{{end}}`
}
