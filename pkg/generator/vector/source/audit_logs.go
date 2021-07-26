package source

import (
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
)

const HostAuditLogTemplate = `
{{define "inputSourceHostAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
  type = "file"
  ignore_older_secs = 600
  include = ["/var/log/audit/audit.log"]
{{end}}`

type HostAuditLog = ConfLiteral

const OpenshiftAuditLogTemplate = `
{{define "inputSourceOpenShiftAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
  type = "file"
  ignore_older_secs = 600
  include = ["/var/log/oauth-apiserver.audit.log"]
{{end}}
`

type OpenshiftAuditLog = ConfLiteral

const K8sAuditLogTemplate = `
{{define "inputSourceK8sAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
  type = "file"
  ignore_older_secs = 600
  include = ["/var/log/kube-apiserver/audit.log"]
{{end}}
`

type K8sAuditLog = ConfLiteral
