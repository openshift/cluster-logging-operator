package viaq

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

const (
	AddDefaultLogLevel = `.level = "default"`

	ParseHostAuditLogs = `
match1 = parse_regex(.message, r'type=(?P<type>[^ ]+)') ?? {}
envelop = {}
envelop |= {"type": match1.type}

match2, err = parse_regex(.message, r'msg=audit\((?P<ts_record>[^ ]+)\):')
if err == null {
  sp, err = split(match2.ts_record,":")
  if err == null && length(sp) == 2 {
      ts = parse_timestamp(sp[0],"%s.%3f") ?? ""
      envelop |= {"record_id": sp[1]}
      . |= {"audit.linux" : envelop}
      . |= {"@timestamp" : format_timestamp(ts,"%+") ?? ""}
  }
} else {
  log("could not parse host audit msg. err=" + err, rate_limit_secs: 0)
}
`
	HostAuditLogTag = ".linux-audit.log"
	K8sAuditLogTag  = ".k8s-audit.log"
	OpenAuditLogTag = ".openshift-audit.log"
	OvnAuditLogTag  = ".ovn-audit.log"
)

var (
	AddHostAuditTag        = fmt.Sprintf(".tag = %q", HostAuditLogTag)
	AddK8sAuditTag         = fmt.Sprintf(".tag = %q", K8sAuditLogTag)
	AddOpenAuditTag        = fmt.Sprintf(".tag = %q", OpenAuditLogTag)
	AddOvnAuditTag         = fmt.Sprintf(".tag = %q", OvnAuditLogTag)
	FixK8sAuditLevel       = `.k8s_audit_level = .level`
	FixOpenshiftAuditLevel = `.openshift_audit_level = .level`
)

func auditHostLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, logging.InputNameAudit, logging.AuditSourceAuditd,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			AddHostAuditTag,
			ParseHostAuditLogs,
			AddDefaultLogLevel,
			FixHostname,
			FixTimestampField,
		}), "\n"))
}

func auditK8sLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, logging.InputNameAudit, logging.AuditSourceKube,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			AddK8sAuditTag,
			ParseAndFlatten,
			FixK8sAuditLevel,
			FixHostname,
			FixTimestampField,
		}), "\n"))
}

func auditOpenshiftLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, logging.InputNameAudit, logging.AuditSourceOpenShift,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			AddOpenAuditTag,
			ParseAndFlatten,
			FixOpenshiftAuditLevel,
			FixHostname,
			FixTimestampField,
		}), "\n"))
}

func auditOVNLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, logging.InputNameAudit, logging.AuditSourceOVN,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			AddOvnAuditTag,
			FixLogLevel,
			FixHostname,
			FixTimestampField,
		}), "\n"))
}

func NormalizeK8sAuditLogs(inputs, id string) []framework.Element {
	return []framework.Element{
		elements.Remap{
			ComponentID: id,
			Inputs:      helpers.MakeInputs(inputs),
			VRL:         auditK8sLogs(),
		},
	}
}
