{{- define "packages" -}}

////
:_mod-docs-content-type: ASSEMBLY
include::_attributes/common-attributes.adoc[]
include::_attributes/attributes-openshift-dedicated.adoc[]
[id="logging-5-x-reference"]
= Logging 5.x API reference
:context: filename

toc::[]

** These release notes are generated from the content in the openshift/cluster-logging-operator repository.
** Do not modify the content here manually except for the metadata and section IDs - changes to the content should be made in the source code.
////

{{ range .packages -}}

    {{- range (sortedTypes (visibleTypes .Types )) -}}
        {{if isObjectRoot . }}

["id=logging-5-x-reference-{{ (typeDisplayName .) }}"]
== {{ (typeDisplayName .) }}

{{  (comments .CommentLines) }}
{{  template "type" (nodeParent . "") -}}

        {{end -}}
    {{end -}}

{{end -}}
{{end -}}
