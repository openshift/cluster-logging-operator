package normalize

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

const (
	AddHostName      = `.hostname = del(.host)`
	AddJournalLogTag = `.tag = ".journal.system"`
	AddTime          = `.time = format_timestamp!(.timestamp, format: "%FT%T%:z")`

	FixJournalLogLevel = `
if .PRIORITY == "8" || .PRIORITY == 8 {
	.level = "trace"
} else {
	priority = to_int!(.PRIORITY)
	.level, err = to_syslog_level(priority)
	if err != null {
		log("Unable to determine level from PRIORITY: " + err, level: "error")
		log(., level: "error")
		.level = "unknown"
	} else {
		del(.PRIORITY)
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
)

func JournalLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				ClusterID,
				AddJournalLogTag,
				DeleteJournalLogFields,
				FixJournalLogLevel,
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

func DropJournalDebugLogs(inLabel, outLabel string) []generator.Element {
	return []generator.Element{
		Filter{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			Condition:   `.PRIORITY != \"7\" && .PRIORITY != 7`,
		},
	}
}
