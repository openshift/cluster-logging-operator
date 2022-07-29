package elements

const ToStdOut = `
{{define "toStdout" -}}
# {{.Desc}}
<match {{.Pattern}}>
 @type stdout
</match>
{{end}}`

type StdOutFilter struct {
	Pattern string
}

func (c StdOutFilter) Name() string {
	return "filterToStdout"
}

func (c StdOutFilter) Template() string {
	return `{{define "` + c.Name() + `"  -}}
<filter {{.Pattern}}>
 @type stdout
</filter>
{{end}}`
}
