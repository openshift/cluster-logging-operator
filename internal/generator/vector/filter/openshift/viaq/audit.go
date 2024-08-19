package viaq

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
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
      if ts != "" { .timestamp = ts }
      ."@timestamp" = format_timestamp(.timestamp, "%+") ?? .timestamp
      envelop |= {"record_id": sp[1]}
      . |= {"audit.linux" : envelop}
  }
} else {
  log("could not parse host audit msg. err=" + err, rate_limit_secs: 0)
}
`
)

var (
	FixK8sAuditLevel       = `.k8s_audit_level = .level`
	FixOpenshiftAuditLevel = `.openshift_audit_level = ._internal.level`
)

func auditHostLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceAuditd,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			Message,
			RemoveFile,
			RemoveSourceType,
			ParseHostAuditLogs,
			AddDefaultLogLevel,
			FixHostname,
			SetTimestampField,
			VRLOpenShiftSequence,
		}), "\n"))
}

func auditK8sLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceKube,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			Message,
			RemoveFile,
			RemoveSourceType,
			ParseAndFlatten,
			FixK8sAuditLevel,
			FixHostname,
			SetTimestampField,
			VRLOpenShiftSequence,
		}), "\n"))
}

func auditOpenshiftLogs() string {
	return strings.Join(helpers.TrimSpaces([]string{
		MoveStructuredToRoot,
		FixHostname,
		FixOpenshiftAuditLevel,
		SetTimestampField,
		SetOpenShift,
		ClusterID,
		VRLOpenShiftSequence,
	}), "\n")
}

func auditOVNLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceOVN,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			RemoveFile,
			RemoveSourceType,
			FixLogLevel,
			FixHostname,
			SetTimestampField,
			VRLOpenShiftSequence,
		}), "\n"))
}
