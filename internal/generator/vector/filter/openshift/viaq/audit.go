package viaq

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
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
)

var (
	FixK8sAuditLevel       = `.k8s_audit_level = .level`
	FixOpenshiftAuditLevel = `.openshift_audit_level = .level`
)

func auditHostLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceAuditd,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			InternalContext,
			RemoveFile,
			RemoveSourceType,
			ParseHostAuditLogs,
			AddDefaultLogLevel,
			FixHostname,
			FixTimestampField,
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
			InternalContext,
			RemoveFile,
			RemoveSourceType,
			ParseAndFlatten,
			FixK8sAuditLevel,
			FixHostname,
			FixTimestampField,
			VRLOpenShiftSequence,
		}), "\n"))
}

func auditOpenshiftLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceOpenShift,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			InternalContext,
			RemoveFile,
			RemoveSourceType,
			ParseAndFlatten,
			FixOpenshiftAuditLevel,
			FixHostname,
			FixTimestampField,
			VRLOpenShiftSequence,
		}), "\n"))
}

func auditOVNLogs() string {
	return fmt.Sprintf(`
if .log_type == "%s" && .log_source == "%s" {
%s
}
`, string(obs.InputTypeAudit), obs.AuditSourceOVN,
		strings.Join(helpers.TrimSpaces([]string{
			ClusterID,
			InternalContext,
			RemoveFile,
			RemoveSourceType,
			FixLogLevel,
			FixHostname,
			FixTimestampField,
			VRLOpenShiftSequence,
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
