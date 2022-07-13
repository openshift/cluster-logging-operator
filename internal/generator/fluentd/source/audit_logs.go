package source

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"strconv"
)

type AuditLogLiteral = generator.ConfLiteral
type AuditLog struct {
	AuditLogLiteral
	Tunings *logging.FluentdInFileSpec
}

func (cl AuditLog) ReadLinesLimit() string {
	if cl.Tunings == nil || cl.Tunings.ReadLinesLimit <= 0 {
		return ""
	}
	return "\n  read_lines_limit " + strconv.Itoa(cl.Tunings.ReadLinesLimit)
}

const HostAuditLogTemplate = `
{{define "inputSourceHostAuditTemplate" -}}
# {{.Desc}}
<source>
  @type tail
  @id audit-input
  @label @{{.OutLabel}}
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  {{- .ReadLinesLimit }}
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>
{{end}}`

const OpenshiftAuditLogTemplate = `
{{define "inputSourceOpenShiftAuditTemplate" -}}
# {{.Desc}}
<source>
  @type tail
  @id openshift-audit-input
  @label @{{.OutLabel}}
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
  {{- .ReadLinesLimit }}
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

const K8sAuditLogTemplate = `
{{define "inputSourceK8sAuditTemplate" -}}
# {{.Desc}}
<source>
  @type tail
  @id k8s-audit-input
  @label @{{.OutLabel}}
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
  {{- .ReadLinesLimit }}
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

const OVNAuditLogTemplate = `
{{define "inputSourceOVNAuditTemplate" -}}  
# {{.Desc}}
<source>
  @type tail
  @id ovn-audit-input
  @label @{{.OutLabel}}
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
  {{- .ReadLinesLimit }}
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
