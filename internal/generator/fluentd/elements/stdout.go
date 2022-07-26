package elements

const ToStdOut = `
{{define "toStdout" -}}
# {{.Desc}}
<match {{.Pattern}}>
 @type stdout
</match>
{{end}}`

const FilterToStdOut = `
{{define "filterToStdout" -}}
# {{.Desc}}
<filter {{.Pattern}}>
 @type stdout
</filter>
{{end}}`
