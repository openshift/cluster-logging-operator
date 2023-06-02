{{- define "packages" -}}
:toc:
:toclevels: 2
:toc-placement!:
toc::[]

{{ range .packages -}}

    {{- range (sortedTypes (visibleTypes .Types )) -}}
        {{if isObjectRoot . }}

== {{ (typeDisplayName .) }}
{{  (comments .CommentLines) }}
{{  template "type" (nodeParent . "") -}}

        {{end -}}
    {{end -}}

{{end -}}
{{end -}}