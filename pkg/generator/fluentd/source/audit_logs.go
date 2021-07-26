package source

import (
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
)

const HostAuditLogTemplate = `
{{define "inputSourceHostAuditTemplate" -}}
# {{.Desc}}
<source>
  @type tail
  @id audit-input
  @label @{{.OutLabel}}
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>
{{end}}`

type HostAuditLog = ConfLiteral

const OpenshiftAuditLogTemplate = `
{{define "inputSourceOpenShiftAuditTemplate" -}}
# {{.Desc}}
<source>
  @type tail
  @id openshift-audit-input
  @label @{{.OutLabel}}
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
  tag openshift-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>
{{end}}
`

type OpenshiftAuditLog = ConfLiteral

const K8sAuditLogTemplate = `
{{define "inputSourceK8sAuditTemplate" -}}
# {{.Desc}}
<source>
  @type tail
  @id k8s-audit-input
  @label @{{.OutLabel}}
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
  tag k8s-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>
{{end}}
`

type K8sAuditLog = ConfLiteral

const OVNAuditLogTemplate = `
{{define "inputSourceOVNAuditTemplate" -}}  
# {{.Desc}}
<source>
  @type tail
  @id ovn-audit-input
  @label @MEASURE
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
  tag ovn-audit.log
  refresh_interval 5
  rotate_wait 5
  read_from_head true
  <parse>
    @type none
  </parse>
</source>
{{end}}
`

type OVNAuditLogs = ConfLiteral
