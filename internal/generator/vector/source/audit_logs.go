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
ignore_older_secs = 3600
max_read_bytes = 3145728
rotate_wait_ms = 5000
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
ignore_older_secs = 3600
max_read_bytes = 3145728
rotate_wait_ms = 5000
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
ignore_older_secs = 3600
max_read_bytes = 3145728
rotate_wait_ms = 5000
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
ignore_older_secs = 3600
max_read_bytes = 3145728
rotate_wait_ms = 5000
{{end}}
`

type K8sAuditLog = framework.ConfLiteral

func NewHostAuditLog(id string) HostAuditLog {
	return HostAuditLog{
		ComponentID:  id,
		Desc:         "Logs from host audit",
		TemplateName: "inputSourceHostAuditTemplate",
		TemplateStr:  HostAuditLogTemplate,
	}
}

func NewK8sAuditLog(id string) K8sAuditLog {
	return K8sAuditLog{
		ComponentID:  id,
		Desc:         "Logs from kubernetes audit",
		TemplateName: "inputSourceK8sAuditTemplate",
		TemplateStr:  K8sAuditLogTemplate,
	}
}

func NewOpenshiftAuditLog(id string) OpenshiftAuditLog {
	return OpenshiftAuditLog{
		ComponentID:  id,
		Desc:         "Logs from openshift audit",
		TemplateName: "inputSourceOpenShiftAuditTemplate",
		TemplateStr:  OpenshiftAuditLogTemplate,
	}
}

func NewOVNAuditLog(id string) OVNAuditLog {
	return OVNAuditLog{
		ComponentID:  id,
		Desc:         "Logs from ovn audit",
		TemplateName: "inputSourceOVNAuditTemplate",
		TemplateStr:  OVNAuditLogTemplate,
	}
}
