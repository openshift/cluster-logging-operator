package elements

const ToStdOut = `
{{define "toStdout" -}}
# {{.Desc}}
<match {{.Pattern}}>
 @type stdout
</match>
{{end}}`
