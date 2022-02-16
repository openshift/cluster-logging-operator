package fluentd

const (
	MultilineDetectExceptionTemplate = `
{{define "matchMultilineDetectException" -}}
<match kubernetes.**>
  @type detect_exceptions
  remove_tag_prefix 'kubernetes'
  message message
  force_line_breaks true
  multiline_flush_interval .2
</match>{{end}}`
)
