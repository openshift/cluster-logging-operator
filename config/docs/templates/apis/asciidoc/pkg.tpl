{{- define "packages" -}}
= Logging API reference
:toc:
:toclevels: 1
:toc-placement!:
:doctype:book

toc::[]

    {{ range .packages -}}
= {{.ApiGroup}}/{{.ApiVersion}}

        {{ range (sortedTypes (visibleTypes .Types )) -}}
            {{ if isObjectRoot . -}}
== {{ (typeDisplayName .) }}
{{ comments .CommentLines }}
{{  template "type" (nodeParent . "") -}}
            {{end -}}
        {{end -}}
    {{end -}}
{{end -}}
