package types

import "time"

// JournalLog is linux journal logs
type JournalLog struct {
	ViaQCommon          `json:",inline,omitempty"`
	Tag                 string    `json:"tag,omitempty"`
	Time                time.Time `json:"time,omitempty"`
	STREAMID            string    `json:"_STREAM_ID,omitempty"`
	SYSTEMDINVOCATIONID string    `json:"_SYSTEMD_INVOCATION_ID,omitempty"`
	Systemd             Systemd   `json:"systemd,omitempty"`
}

type K struct {
	KERNEL_DEVICE    string `json:"KERNEL_DEVICE,omitempty"`
	KERNEL_SUBSYSTEM string `json:"KERNEL_SUBSYSTEM,omitempty"`
	UDEV_DEVLINK     string `json:"UDEV_DEVLINK,omitempty"`
	UDEV_DEVNODE     string `json:"UDEV_DEVNODE,omitempty"`
	UDEV_SYSNAME     string `json:"UDEV_SYSNAME,omitempty"`
}

type T struct {
	AUDIT_LOGINUID      string `json:"AUDIT_LOGINUID,omitempty"`
	AUDIT_SESSION       string `json:"AUDIT_SESSION,omitempty"`
	BOOTID              string `json:"BOOT_ID,omitempty"`
	CAPEFFECTIVE        string `json:"CAP_EFFECTIVE,omitempty"`
	CMDLINE             string `json:"CMDLINE,omitempty"`
	COMM                string `json:"COMM,omitempty"`
	EXE                 string `json:"EXE,omitempty"`
	GID                 string `json:"GID,omitempty"`
	HOSTNAME            string `json:"HOSTNAME,omitempty"`
	LINE_BREAK          string `json:"LINE_BREAK,omitempty"`
	MACHINEID           string `json:"MACHINE_ID,omitempty"`
	PID                 string `json:"PID,omitempty"`
	SELINUXCONTEXT      string `json:"SELINUX_CONTEXT,omitempty"`
	STREAMID            string `json:"STREAM_ID,omitempty"`
	SYSTEMDCGROUP       string `json:"SYSTEMD_CGROUP,omitempty"`
	SYSTEMDINVOCATIONID string `json:"SYSTEMD_INVOCATION_ID,omitempty"`
	SYSTEMD_OWNER_UID   string `json:"SYSTEMD_OWNER_UID,omitempty"`
	SYSTEMD_SESSION     string `json:"SYSTEMD_SESSION,omitempty"`
	SYSTEMDSLICE        string `json:"SYSTEMD_SLICE,omitempty"`
	SYSTEMDUNIT         string `json:"SYSTEMD_UNIT,omitempty"`
	SYSTEMD_USER_UNIT   string `json:"SYSTEMD_USER_UNIT,omitempty"`
	TRANSPORT           string `json:"TRANSPORT,omitempty"`
	UID                 string `json:"UID,omitempty"`
}

type U struct {
	CODE_FILE        string `json:"CODE_FILE,omitempty"`
	CODE_FUNCTION    string `json:"CODE_FUNCTION,omitempty"`
	CODE_LINE        string `json:"CODE_LINE,omitempty"`
	ERRNO            string `json:"ERRNO,omitempty"`
	MESSAGE_ID       string `json:"MESSAGE_ID,omitempty"`
	SYSLOG_FACILITY  string `json:"SYSLOG_FACILITY,omitempty"`
	SYSLOGIDENTIFIER string `json:"SYSLOG_IDENTIFIER,omitempty"`
	SYSLOG_PID       string `json:"SYSLOG_PID,omitempty"`
	RESULT           string `json:"RESULT,omitempty"`
	UNIT             string `json:"UNIT,omitempty"`
}

type Systemd struct {
	K K `json:"k,omitempty"`
	T T `json:"t,omitempty"`
	U U `json:"u,omitempty"`
}
