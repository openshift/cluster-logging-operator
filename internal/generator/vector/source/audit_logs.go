package source

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

const HostAuditLogTemplate = `
{{define "inputSourceHostAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
include = ["/var/log/audit/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000
{{end}}`

type HostAuditLog = framework.ConfLiteral

const OpenshiftAuditLogTemplate = `
{{define "inputSourceOpenShiftAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log","/var/log/oauth-server/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000
{{end}}
`

type OpenshiftAuditLog = framework.ConfLiteral

const K8sAuditLogTemplate = `
{{define "inputSourceK8sAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
include = ["/var/log/kube-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000
{{end}}
`

type OVNAuditLog = framework.ConfLiteral

const OVNAuditLogTemplate = `
{{define "inputSourceOVNAuditTemplate" -}}
# {{.Desc}}
[sources.{{.ComponentID}}]
type = "file"
include = ["/var/log/ovn/acl-audit-log.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000
{{end}}
`

type K8sAuditLog = framework.ConfLiteral
