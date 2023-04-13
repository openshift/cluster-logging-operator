{{ define "member" }}
    {{- if not (or .Type.IsPrimitive (eq (yamlType .Type) "string") (fieldEmbedded .Member) (hiddenMember .Member) (ignoreMember .Member)) }}

[id=_{{ .Path }}]
=== {{ .Path }}

===== Description

{{ if .Type.Elem }}{{ comments .Type.Elem.CommentLines }}{{ else }}{{ comments .Type.CommentLines }}{{end }}

=====  Type

* {{ (yamlType .Type) }}

{{ template "properties" (nodeParent .Type .Path) }}

{{ template "members" (nodeParent .Type .Path) }}

    {{- end }}
{{- end }}

{{ define "members" }}
    {{- $path := .Path -}}
    {{- if .Members }}
        {{- range .Members }}
            {{- if not (or (eq .Name "ObjectMeta") (eq (fieldName .) "TypeMeta") (ignoreMember .)) }}
                {{- template "member" (node . (printf "%s.%s" $path  (fieldName .))) }}
            {{- end}}
        {{- end}}
    {{- else if (eq (yamlType .Type) "array") -}}
        {{- template "type" (nodeParent .Elem $path) }}
    {{- else if .Elem -}}
        {{- template "type" (nodeParent .Elem $path) }}
    {{- end }}
{{ end -}}

{{ define "properties" }}
    {{- $path := .Path -}}
    {{- if .Members }}

[options="header"]
|======================
|Property|Type|Description
	  {{- range ( sortMembers .Members) -}}
	       {{- if not (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta") (ignoreMember .)) -}}
		   {{- if (fieldEmbedded . ) -}}
		       {{- template "rows" (nodeParent .Type $path)  -}}
		   {{- else -}}
		       {{- template "row" (node . $path) -}}
		   {{- end -}}
	       {{- end -}}
	  {{- end -}}
|======================
    {{- end }}
{{- end }}

{{- define "rows" -}}
    {{ $path := .Path }}
    {{- if .Members -}}
        {{- range .Members -}}
            {{- template "row" (node . $path) }}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{ define "row" }}
   {{- $path := .Path }}
   {{- with .Member}}
       {{- if not (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta") (ignoreMember .)) -}}
           {{- $extra := "" -}}
           {{- if (isDeprecatedMember .) -}}
              {{- $extra = "**(DEPRECATED)**" -}}
           {{- end -}}
           {{- if (isOptionalMember .) -}}
              {{- $extra = (printf "%s %s" $extra "*(optional)*") -}}
           {{- end }}
| {{ (fieldName .) -}}
| {{ $t := yamlType .Type }}{{- if eq $t "object" }}xref:#_{{ $path }}.{{fieldName .}}[object]{{else}}{{$t}}{{end -}}
| {{ $extra }} {{ (comments .CommentLines "summary") }}
       {{- end }}
   {{- end }}
{{ end }}
