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
match1 = parse_regex(._internal.message, r'type=(?P<type>[^ ]+)') ?? {}
envelop = {}
envelop |= {"type": match1.type}

match2, err = parse_regex(._internal.message, r'msg=audit\((?P<ts_record>[^ ]+)\):')
if err == null {
  sp, err = split(match2.ts_record,":")
  if err == null && length(sp) == 2 {
      ts = parse_timestamp(sp[0],"%s.%3f") ?? ""
      envelop |= {"record_id": sp[1]}
      ._internal |= {"audit.linux" : envelop}
      ._internal."@timestamp" =  format_timestamp(ts,"%+") ?? ""
  }
} else {
  log("could not parse host audit msg. err=" + err, rate_limit_secs: 0)
}
`
	SetK8sAuditLevel       = `.k8s_audit_level = ._internal.structured.level`
	SetOpenshiftAuditLevel = `.openshift_audit_level = ._internal.structured.level`
)

func auditHostLogs() string {
	return fmt.Sprintf(`
if ._internal.log_type == "%s" && ._internal.log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceAuditd,
		strings.Join(helpers.TrimSpaces([]string{
			SetMessageOnRoot,
			`."audit.linux" = ._internal."audit.linux"`,
			AddDefaultLogLevel,
		}), "\n"))
}

func auditK8sLogs() string {
	return fmt.Sprintf(`
if ._internal.log_type == "%s" && ._internal.log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceKube,
		strings.Join(helpers.TrimSpaces([]string{
			SetK8sAuditLevel,
		}), "\n"))
}

func auditOpenshiftLogs() string {
	return fmt.Sprintf(`
if ._internal.log_type == "%s" && ._internal.log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceOpenShift,
		strings.Join(helpers.TrimSpaces([]string{
			SetOpenshiftAuditLevel,
		}), "\n"))
}

func auditOVNLogs() string {
	return ""
}
