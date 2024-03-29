{{- define "member" -}}
 {{- if not (or .Type.IsPrimitive (eq (yamlType .Type) "string")) -}}
 {{ if not (or (or (fieldEmbedded .Member) (hiddenMember .Member) (ignoreMember .Member))) }}

=== {{ .Path }}

{{ if (isDeprecatedMember .Member) }}
  {{ "[IMPORTANT]\n" -}}
  {{ "====\n" -}}
  {{ "This API key has been deprecated and is planned for removal in a future release. For more information, see the release notes for logging on Red{nbsp}Hat OpenShift.\n" -}}
  {{ "====" -}}
{{ end }}

{{ if .Type.Elem -}}
  {{ (comments .Type.Elem.CommentLines) }}
{{- else -}}
  {{  (comments .Type.CommentLines) }}
{{- end }}

Type:: {{ (yamlType .Type) }}

{{- template "properties" .Type  -}}
{{- template "members" (nodeParent .Type .Path) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "members" -}}
{{- $path := .Path -}}
{{- if .Members -}}

   {{ range .Members }}
       {{- if (or (or (eq .Name "ObjectMeta") (eq (fieldName .) "TypeMeta")) (ignoreMember .)) -}}
        {{- else -}}
              {{- if (eq (yamlType .Type) "array") }}
                {{ template "member" (node . (printf "%s.%s[]" $path  (fieldName .))) }}
              {{ else }}
                {{ template "member" (node . (printf "%s.%s" $path  (fieldName .))) }}
              {{- end -}}
       {{- end -}}
   {{- end -}}

{{- else if (eq (yamlType .Type) "array") -}}
    {{ template "type" (nodeParent .Elem $path) }}
{{- else if .Elem -}}
    {{ template "type" (nodeParent .Elem $path) }}
{{- end -}}
{{- end -}}

{{- define "properties" -}}
{{- if .Members }}

[options="header"]
|======================
|Property|Type|Description
{{ "\n" }}
{{ range (sortMembers .Members) -}}
{{- if or (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta")) (ignoreMember .) -}}
{{- else -}}
{{- if ne (fieldName .) "" }}{{ "\n" }}{{ end -}}
{{- template "row" . }}
{{- end -}}
{{- end -}}
|======================
{{- end -}}
{{- end -}}

{{- define "row" }}
{{- if or (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta")) (ignoreMember .) -}}
{{- else -}}
{{- $extra := "" -}}
{{- if isDeprecatedMember . -}}
{{- $extra = "\n[IMPORTANT]\n====\nThis API key has been deprecated and is planned for removal in a future release. For more information, see the release notes for logging on Red{nbsp}Hat OpenShift.\n====\n" -}}
{{- end -}}
{{- if isOptionalMember . -}}
{{- $extra = "\n[NOTE]\n====\nThis API key is optional.\n====\n" -}}
{{- end -}}
{{- printf "| %s\n| %s\n| %s%s\n" (fieldName .) (yamlType .Type) (comments .CommentLines "") $extra -}}
{{- end -}}
{{- end -}}
