{{- define "packages" -}}
= Logging Data Model Reference

:toc:
:toclevels: 2
:doctype: book

{{ range .packages -}}

= Package {{ packageDisplayName . }}

    {{- range (sortedTypes (visibleTypes .Types )) -}}
        {{- if isObjectRoot . -}}

        == {{ (typeDisplayName .) }}
        {{  (comments .CommentLines) }}
        {{- template "type" (nodeParent . "") -}}

        {{- end -}}
    {{- end -}}

{{- end -}}
{{- end -}}