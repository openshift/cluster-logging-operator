package vector

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	FixLogLevel = `
if !exists(.level) {
  .level = "default"
  if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
    .level = "info"
  } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
    .level = "warn"
  } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
    .level = "error"
  } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
    .level = "critical"
  } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
    .level = "debug"
  } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
    .level = "notice"
  } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
    .level = "alert"
  } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
    .level = "emergency"
  }
}
`
	RemoveSourceType  = `del(.source_type)`
	RemoveStream      = `del(.stream)`
	RemovePodIPs      = `del(.kubernetes.pod_ips)`
	FixTimestampField = `."@timestamp" = del(.timestamp)`
	AddJournalLogTag  = `.tag = ".journal.system"`
	AddHostName       = `.hostname = del(.host)`
	AddTime           = `.time = format_timestamp!(.timestamp, format: "%FT%T%:z")`

	DeleteJournalLogFields = `
del(.source_type)
del(._CPU_USAGE_NSEC)
del(.__REALTIME_TIMESTAMP)
del(.__MONOTONIC_TIMESTAMP)
del(._SOURCE_REALTIME_TIMESTAMP)
del(.PRIORITY)
del(.JOB_RESULT)
del(.JOB_TYPE)
del(.TIMESTAMP_BOOTTIME)
del(.TIMESTAMP_MONOTONIC)
`
	SystemK = `
# systemd’s kernel-specific metadata.
# .systemd.k = {}
if exists(.KERNEL_DEVICE) { .systemd.k.KERNEL_DEVICE = del(.KERNEL_DEVICE) }
if exists(.KERNEL_SUBSYSTEM) { .systemd.k.KERNEL_SUBSYSTEM = del(.KERNEL_SUBSYSTEM) }
if exists(.UDEV_DEVLINK) { .systemd.k.UDEV_DEVLINK = del(.UDEV_DEVLINK) }
if exists(.UDEV_DEVNODE) { .systemd.k.UDEV_DEVNODE = del(.UDEV_DEVNODE) }
if exists(.UDEV_SYSNAME) { .systemd.k.UDEV_SYSNAME = del(.UDEV_SYSNAME) }
`
	SystemT = `
# trusted journal fields, fields that are implicitly added by the journal and cannot be altered by client code.
.systemd.t = {}
if exists(._AUDIT_LOGINUID) { .systemd.t.AUDIT_LOGINUID = del(._AUDIT_LOGINUID) }
if exists(._BOOT_ID) { .systemd.t.BOOT_ID = del(._BOOT_ID) }
if exists(._AUDIT_SESSION) { .systemd.t.AUDIT_SESSION = del(._AUDIT_SESSION) }
if exists(._CAP_EFFECTIVE) { .systemd.t.CAP_EFFECTIVE = del(._CAP_EFFECTIVE) }
if exists(._CMDLINE) { .systemd.t.CMDLINE = del(._CMDLINE) }
if exists(._COMM) { .systemd.t.COMM = del(._COMM) }
if exists(._EXE) { .systemd.t.EXE = del(._EXE) }
if exists(._GID) { .systemd.t.GID = del(._GID) }
if exists(._HOSTNAME) { .systemd.t.HOSTNAME = .hostname }
if exists(._LINE_BREAK) { .systemd.t.LINE_BREAK = del(._LINE_BREAK) }
if exists(._MACHINE_ID) { .systemd.t.MACHINE_ID = del(._MACHINE_ID) }
if exists(._PID) { .systemd.t.PID = del(._PID) }
if exists(._SELINUX_CONTEXT) { .systemd.t.SELINUX_CONTEXT = del(._SELINUX_CONTEXT) }
if exists(._SOURCE_REALTIME_TIMESTAMP) { .systemd.t.SOURCE_REALTIME_TIMESTAMP = del(._SOURCE_REALTIME_TIMESTAMP) }
if exists(._STREAM_ID) { .systemd.t.STREAM_ID = ._STREAM_ID }
if exists(._SYSTEMD_CGROUP) { .systemd.t.SYSTEMD_CGROUP = del(._SYSTEMD_CGROUP) }
if exists(._SYSTEMD_INVOCATION_ID) {.systemd.t.SYSTEMD_INVOCATION_ID = ._SYSTEMD_INVOCATION_ID}
if exists(._SYSTEMD_OWNER_UID) { .systemd.t.SYSTEMD_OWNER_UID = del(._SYSTEMD_OWNER_UID) }
if exists(._SYSTEMD_SESSION) { .systemd.t.SYSTEMD_SESSION = del(._SYSTEMD_SESSION) }
if exists(._SYSTEMD_SLICE) { .systemd.t.SYSTEMD_SLICE = del(._SYSTEMD_SLICE) }
if exists(._SYSTEMD_UNIT) { .systemd.t.SYSTEMD_UNIT = del(._SYSTEMD_UNIT) }
if exists(._SYSTEMD_USER_UNIT) { .systemd.t.SYSTEMD_USER_UNIT = del(._SYSTEMD_USER_UNIT) }
if exists(._TRANSPORT) { .systemd.t.TRANSPORT = del(._TRANSPORT) }
if exists(._UID) { .systemd.t.UID = del(._UID) }
`
	SystemU = `
# fields that are directly passed from clients and stored in the journal.
.systemd.u = {}
if exists(.CODE_FILE) { .systemd.u.CODE_FILE = del(.CODE_FILE) }
if exists(.CODE_FUNC) { .systemd.u.CODE_FUNCTION = del(.CODE_FUNC) }
if exists(.CODE_LINE) { .systemd.u.CODE_LINE = del(.CODE_LINE) }
if exists(.ERRNO) { .systemd.u.ERRNO = del(.ERRNO) }
if exists(.MESSAGE_ID) { .systemd.u.MESSAGE_ID = del(.MESSAGE_ID) }
if exists(.SYSLOG_FACILITY) { .systemd.u.SYSLOG_FACILITY = del(.SYSLOG_FACILITY) }
if exists(.SYSLOG_IDENTIFIER) { .systemd.u.SYSLOG_IDENTIFIER = del(.SYSLOG_IDENTIFIER) }
if exists(.SYSLOG_PID) { .systemd.u.SYSLOG_PID = del(.SYSLOG_PID) }
if exists(.RESULT) { .systemd.u.RESULT = del(.RESULT) }
if exists(.UNIT) { .systemd.u.UNIT = del(.UNIT) }
`
	ParseHostAuditLogs = `
match1 = parse_regex(.message, r'type=(?P<type>[^ ]+)') ?? {}
envelop = {}
envelop |= {"type": match1.type}

match2, err = parse_regex(.message, r'msg=audit\((?P<ts_record>[^ ]+)\):')
if err == null {
  sp = split(match2.ts_record,":")
  if length(sp) == 2 {
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
	ParseAndFlatten = `. = merge(., parse_json!(string!(.message))) ?? .
del(.message)
`
	FixHostname = `.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`
)

var (
	AddHostAuditTag = fmt.Sprintf(".tag = %q", HostAuditLogTag)
	AddK8sAuditTag  = fmt.Sprintf(".tag = %q", K8sAuditLogTag)
	AddOpenAuditTag = fmt.Sprintf(".tag = %q", OpenAuditLogTag)
	AddOvnAuditTag  = fmt.Sprintf(".tag = %q", OvnAuditLogTag)
)

func NormalizeLogs(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	types := generator.GatherSources(spec, op)
	var el []generator.Element = make([]generator.Element, 0)
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el, NormalizeContainerLogs("raw_container_logs", "container_logs")...)
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el, NormalizeJournalLogs("raw_journal_logs", "journal_logs")...)
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el, NormalizeHostAuditLogs(RawHostAuditLogs, HostAuditLogs)...)
		el = append(el, NormalizeK8sAuditLogs(RawK8sAuditLogs, K8sAuditLogs)...)
		el = append(el, NormalizeOpenshiftAuditLogs(RawOpenshiftAuditLogs, OpenshiftAuditLogs)...)
		el = append(el, NormalizeOVNAuditLogs(RawOvnAuditLogs, OvnAuditLogs)...)
	}
	return el
}

func NormalizeContainerLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				FixLogLevel,
				RemoveSourceType,
				RemoveStream,
				RemovePodIPs,
				FixTimestampField,
			}), "\n"),
		},
	}
}

func NormalizeJournalLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				AddJournalLogTag,
				DeleteJournalLogFields,
				FixLogLevel,
				AddHostName,
				SystemK,
				SystemT,
				SystemU,
				AddTime,
				FixTimestampField,
			}), "\n\n"),
		},
	}
}

func NormalizeHostAuditLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				AddHostAuditTag,
				ParseHostAuditLogs,
			}), "\n\n"),
		},
	}
}

func NormalizeK8sAuditLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				AddK8sAuditTag,
				ParseAndFlatten,
			}), "\n"),
		},
	}
}

func NormalizeOpenshiftAuditLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				AddOpenAuditTag,
				ParseAndFlatten,
			}), "\n"),
		},
	}
}

func NormalizeOVNAuditLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				AddOvnAuditTag,
				FixLogLevel,
			}), "\n"),
		},
	}
}
