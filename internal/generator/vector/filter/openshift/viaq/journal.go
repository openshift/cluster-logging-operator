package viaq

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	AddHostName      = `.hostname = del(.internal.host)`
	AddJournalLogTag = `.tag = ".journal.system"`
	AddTime          = `.time = format_timestamp!(.internal.timestamp, format: "%FT%T%:z")`

	FixJournalLogLevel = `
if ._internal.PRIORITY == "8" || ._internal.PRIORITY == 8 {
	._internal.level = "trace"
} else {
	priority = to_int!(._internal.PRIORITY)
	._internal.level, err = to_syslog_level(priority)
	if err != null {
		log("Unable to determine level from PRIORITY: " + err, level: "error")
		log(., level: "error")
		._internal.level = "unknown"
	} else {
		del(._internal.PRIORITY)
	}
}
`
	DeleteJournalLogFields = `
del(.source_type)
del(._CPU_USAGE_NSEC)
del(.__REALTIME_TIMESTAMP)
del(.__MONOTONIC_TIMESTAMP)
del(._SOURCE_REALTIME_TIMESTAMP)
del(.JOB_RESULT)
del(.JOB_TYPE)
del(.TIMESTAMP_BOOTTIME)
del(.TIMESTAMP_MONOTONIC)
`
	SystemK = `
# systemd’s kernel-specific metadata.
# .systemd.k = {}
if exists(._internal.KERNEL_DEVICE) { ._internal.systemd.k.KERNEL_DEVICE = del(._internal.KERNEL_DEVICE) }
if exists(._internal.KERNEL_SUBSYSTEM) { ._internal.systemd.k.KERNEL_SUBSYSTEM = del(._internal.KERNEL_SUBSYSTEM) }
if exists(._internal.UDEV_DEVLINK) { ._internal.systemd.k.UDEV_DEVLINK = del(._internal.UDEV_DEVLINK) }
if exists(._internal.UDEV_DEVNODE) { ._internal.systemd.k.UDEV_DEVNODE = del(._internal.UDEV_DEVNODE) }
if exists(._internal.UDEV_SYSNAME) { ._internal.systemd.k.UDEV_SYSNAME = del(._internal.UDEV_SYSNAME) }
`
	SystemT = `
# trusted journal fields, fields that are implicitly added by the journal and cannot be altered by client code.
._internal.systemd.t = {}
if exists(._internal._AUDIT_LOGINUID) { ._internal.systemd.t.AUDIT_LOGINUID = del(._internal._AUDIT_LOGINUID) }
if exists(._internal._BOOT_ID) { ._internal.systemd.t.BOOT_ID = del(._internal._BOOT_ID) }
if exists(._internal._AUDIT_SESSION) { ._internal.systemd.t.AUDIT_SESSION = del(._internal._AUDIT_SESSION) }
if exists(._internal._CAP_EFFECTIVE) { ._internal.systemd.t.CAP_EFFECTIVE = del(._internal._CAP_EFFECTIVE) }
if exists(._internal._CMDLINE) { ._internal.systemd.t.CMDLINE = del(._internal._CMDLINE) }
if exists(._internal._COMM) { ._internal.systemd.t.COMM = del(._internal._COMM) }
if exists(._internal._EXE) { ._internal.systemd.t.EXE = del(._internal._EXE) }
if exists(._internal._GID) { ._internal.systemd.t.GID = del(._internal._GID) }
if exists(._internal._HOSTNAME) { ._internal.systemd.t.HOSTNAME = ._internal.hostname }
if exists(._internal._LINE_BREAK) { ._internal.systemd.t.LINE_BREAK = del(._internal._LINE_BREAK) }
if exists(._internal._MACHINE_ID) { ._internal.systemd.t.MACHINE_ID = del(._internal._MACHINE_ID) }
if exists(._internal._PID) { ._internal.systemd.t.PID = del(._internal._PID) }
if exists(._internal._SELINUX_CONTEXT) { ._internal.systemd.t.SELINUX_CONTEXT = del(._internal._SELINUX_CONTEXT) }
if exists(._internal._SOURCE_REALTIME_TIMESTAMP) { ._internal.systemd.t.SOURCE_REALTIME_TIMESTAMP = del(._internal._SOURCE_REALTIME_TIMESTAMP) }
if exists(._internal._STREAM_ID) { ._internal.systemd.t.STREAM_ID = ._internal._STREAM_ID }
if exists(._internal._SYSTEMD_CGROUP) { ._internal.systemd.t.SYSTEMD_CGROUP = del(._internal._SYSTEMD_CGROUP) }
if exists(._internal._SYSTEMD_INVOCATION_ID) {._internal.systemd.t.SYSTEMD_INVOCATION_ID = ._internal._SYSTEMD_INVOCATION_ID}
if exists(._internal._SYSTEMD_OWNER_UID) { ._internal.systemd.t.SYSTEMD_OWNER_UID = del(._internal._SYSTEMD_OWNER_UID) }
if exists(._internal._SYSTEMD_SESSION) { ._internal.systemd.t.SYSTEMD_SESSION = del(._internal._SYSTEMD_SESSION) }
if exists(._internal._SYSTEMD_SLICE) { ._internal.systemd.t.SYSTEMD_SLICE = del(._internal._SYSTEMD_SLICE) }
if exists(._internal._SYSTEMD_UNIT) { ._internal.systemd.t.SYSTEMD_UNIT = del(._internal._SYSTEMD_UNIT) }
if exists(._internal._SYSTEMD_USER_UNIT) { ._internal.systemd.t.SYSTEMD_USER_UNIT = del(._internal._SYSTEMD_USER_UNIT) }
if exists(._internal._TRANSPORT) { ._internal.systemd.t.TRANSPORT = del(._internal._TRANSPORT) }
if exists(._internal._UID) { ._internal.systemd.t.UID = del(._internal._UID) }
`
	SystemU = `
# fields that are directly passed from clients and stored in the journal.
._internal.systemd.u = {}
if exists(._internal.CODE_FILE) { ._internal.systemd.u.CODE_FILE = del(._internal.CODE_FILE) }
if exists(._internal.CODE_FUNC) { ._internal.systemd.u.CODE_FUNCTION = del(._internal.CODE_FUNC) }
if exists(._internal.CODE_LINE) { ._internal.systemd.u.CODE_LINE = del(._internal.CODE_LINE) }
if exists(._internal.ERRNO) { ._internal.systemd.u.ERRNO = del(._internal.ERRNO) }
if exists(._internal.MESSAGE_ID) { ._internal.systemd.u.MESSAGE_ID = del(._internal.MESSAGE_ID) }
if exists(._internal.SYSLOG_FACILITY) { ._internal.systemd.u.SYSLOG_FACILITY = del(._internal.SYSLOG_FACILITY) }
if exists(._internal.SYSLOG_IDENTIFIER) { ._internal.systemd.u.SYSLOG_IDENTIFIER = del(._internal.SYSLOG_IDENTIFIER) }
if exists(._internal.SYSLOG_PID) { ._internal.systemd.u.SYSLOG_PID = del(._internal.SYSLOG_PID) }
if exists(._internal.RESULT) { ._internal.systemd.u.RESULT = del(._internal.RESULT) }
if exists(._internal.UNIT) { ._internal.systemd.u.UNIT = del(._internal.UNIT) }
`
)

func NewJournal(id string, inputs ...string) framework.Element {
	return DropJournalDebugLogs(id, inputs...)
}

func journalLogs() string {
	return fmt.Sprintf(`
if .log_source == "%s" {
  %s
}
`, obs.InfrastructureSourceNode, journalLogsVRL())
}

func journalLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		AddJournalLogTag,
		FixJournalLogLevel,
		AddHostName,
		AddTime,
		`.systemd = ._internal.systemd`,
	}), "\n\n")
}

func DropJournalDebugLogs(id string, inputs ...string) framework.Element {
	return Filter{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		Condition:   `(.internal.log_source == "node" && .internal.PRIORITY != "7" && .internal.PRIORITY != 7)  || .internal.log_source == "container" || .internal.log_type == "audit"`,
	}
}
