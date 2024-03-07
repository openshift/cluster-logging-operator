{{- define "packages" -}}

// Module included in the following assemblies:
//
// * /logging/api_reference/logging-5-x-reference.adoc
// UPDATE THE NAME OF THE ASSEMBLY FOR THE API VERSION

// :_mod-docs-content-type: <REFERENCE>
// [id="filename_{context}"]

{{ range .packages -}}

    {{- range (sortedTypes (visibleTypes .Types )) -}}
        {{if isObjectRoot . }}

= {{ (typeDisplayName .) }}

{{  (comments .CommentLines) }}
{{  template "type" (nodeParent . "") -}}

        {{end -}}
    {{end -}}

{{end -}}
{{end -}}
