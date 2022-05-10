package source

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

const HostAuditLogTemplate = `
{{define "inputSourceHostAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
ignore_older_secs = 600
include = ["/var/log/audit/audit.log"]
{{end}}`

type HostAuditLog = generator.ConfLiteral

const OpenshiftAuditLogTemplate = `
{{define "inputSourceOpenShiftAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
ignore_older_secs = 600
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]
{{end}}
`

type OpenshiftAuditLog = generator.ConfLiteral

const K8sAuditLogTemplate = `
{{define "inputSourceK8sAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
ignore_older_secs = 600
include = ["/var/log/kube-apiserver/audit.log"]
{{end}}
`

type OVNAuditLog = generator.ConfLiteral

const OVNAuditLogTemplate = `
{{define "inputSourceOVNAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
ignore_older_secs = 600
include = ["/var/log/ovn/acl-audit-log.log"]
{{end}}
`

type K8sAuditLog = generator.ConfLiteral
