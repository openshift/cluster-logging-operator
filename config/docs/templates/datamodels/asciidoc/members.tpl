{{ define "member" }}


{{ if not (or (or (hiddenMember .Member) (ignoreMember .Member))) }}

    {{ if not (fieldEmbedded .Member) }}
    === {{ .Path }}
    ===== Description
       {{ $extra := "" }}
       {{ if (isDeprecatedMember .Member) }}
          {{ $extra = "**(DEPRECATED)**" }}
       {{ end }}
       {{ if (isOptionalMember .Member) }}
          {{ $extra = (printf "%s %s" $extra "*(optional)*") }}
       {{ end }}
      {{$extra}} {{ (comments .CommentLines) }}

    =====  Type
    * {{ (yamlType .Type) }}
    {{ end }}
    {{ if not (or .Type.IsPrimitive (eq (yamlType .Type) "string")) }}
        {{- template "properties" .Type  -}}
        {{ template "members" (nodeParent .Type .Path) }}
    {{ end }}
{{ end }}
{{ end }}

{{ define "members" }}
{{ $path := .Path }}
{{ if .Members }}

   {{ range .Members }}
       {{ if (or (eq .Name "ObjectMeta") (eq (fieldName .) "TypeMeta")) }}
        {{ else }}
         {{ if (fieldEmbedded .) }}
            {{ template "member"  (node . $path ) }}
         {{ else }}
              {{if (eq (yamlType .Type) "array") }}
                {{ template "member" (node . (printf "%s.%s[]" $path  (fieldName .))) }}
              {{ else }}
                {{ template "member" (node . (printf "%s.%s" $path  (fieldName .))) }}
              {{ end }}

         {{ end }}
       {{ end }}
   {{ end }}

{{ else if (eq (yamlType .Type) "array") }}
    {{ template "type" (nodeParent . $path) }}
{{ else }}
{{end }}

{{end }}


{{- define "properties" -}}
{{- if .Members }}

[options="header"]
|======================
|Property|Type|Description
    {{ range ( sortMembers .Members) -}}
       {{- if (or (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta")) (ignoreMember .)) -}}
       {{- else -}}
         {{- if (fieldEmbedded . ) -}}
           {{- template "rows" .Type  -}}
         {{- else -}}
           {{- template "row" .  -}}
         {{- end -}}
       {{- end -}}
   {{- end -}}
|======================
{{- end -}}
{{- end -}}

{{ define "rows" }}
{{ if .Members }}
   {{ range .Members }}
        {{ template "row" . }}
   {{ end }}
{{ end }}
{{ end }}

{{ define "row" }}
       {{ if (or (or (eq (fieldName .) "metadata") (eq (fieldName .) "TypeMeta")) (ignoreMember .)) }}
       {{ else }}
           {{ $extra := "" }}
           {{ if (isDeprecatedMember .) }}
              {{ $extra = "**(DEPRECATED)**" }}
           {{ end }}
           {{ if (isOptionalMember .) }}
              {{ $extra = (printf "%s %s" $extra "*(optional)*") }}
           {{ end }}
           |{{ (fieldName .) }}
           |{{ (yamlType .Type)}}
           a| {{ $extra }} {{ (comments .CommentLines "summary")}}
       {{ end }}
{{ end }}