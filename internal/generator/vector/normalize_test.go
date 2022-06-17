package vector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Vector Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return generator.MergeElements(
			NormalizeLogs(&clfspec, op),
		)
	}
	DescribeTable("NormalizeLogs(s)", helpers.TestGenerateConfWith(f),
		Entry("Application logs only", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = '''
  level = "unknown"
  if match!(.message,r'(Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>)'){
    level = "warn"
  } else if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>'){
    level = "info"
  } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>'){
    level = "error"
  } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>'){
    level = "critical"
  } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>'){
    level = "debug"
  }
  .level = level
  del(.source_type)
  del(.stream)
  del(.kubernetes.pod_ips)
  ."@timestamp" = del(.timestamp)
'''
`,
		}),
		Entry("Infrastructure logs only", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameInfrastructure,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = '''
  level = "unknown"
  if match!(.message,r'(Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>)'){
    level = "warn"
  } else if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>'){
    level = "info"
  } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>'){
    level = "error"
  } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>'){
    level = "critical"
  } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>'){
    level = "debug"
  }
  .level = level
  del(.source_type)
  del(.stream)
  del(.kubernetes.pod_ips)
  ."@timestamp" = del(.timestamp)
'''

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = '''
  .tag = ".journal.system"
  
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
  
  level = "unknown"
  if match!(.message,r'(Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>)'){
    level = "warn"
  } else if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>'){
    level = "info"
  } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>'){
    level = "error"
  } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>'){
    level = "critical"
  } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>'){
    level = "debug"
  }
  .level = level
  
  .hostname = del(.host)
  
  # systemd’s kernel-specific metadata.
  # .systemd.k = {}
  if exists(.KERNEL_DEVICE) { .systemd.k.KERNEL_DEVICE = del(.KERNEL_DEVICE) }
  if exists(.KERNEL_SUBSYSTEM) { .systemd.k.KERNEL_SUBSYSTEM = del(.KERNEL_SUBSYSTEM) }
  if exists(.UDEV_DEVLINK) { .systemd.k.UDEV_DEVLINK = del(.UDEV_DEVLINK) }
  if exists(.UDEV_DEVNODE) { .systemd.k.UDEV_DEVNODE = del(.UDEV_DEVNODE) }
  if exists(.UDEV_SYSNAME) { .systemd.k.UDEV_SYSNAME = del(.UDEV_SYSNAME) }
  
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
  
  .time = format_timestamp!(.timestamp, format: "%FT%T%:z")
  
  ."@timestamp" = del(.timestamp)
'''

`,
		}),
		Entry("Audit logs only", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.host_audit_logs]
type = "remap"
inputs = ["raw_host_audit_logs"]
source = '''
  .tag = ".linux-audit.log"
  
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
  hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  del(.host)
  . |= {"hostname" : hostname}
'''
    
[transforms.k8s_audit_logs]
type = "remap"
inputs = ["raw_k8s_audit_logs"]
source = '''
  .tag = ".k8s-audit.log"
  
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
'''

[transforms.openshift_audit_logs]
type = "remap"
inputs = ["raw_openshift_audit_logs"]
source = '''
  .tag = ".openshift-audit.log"
  
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
'''

[transforms.ovn_audit_logs]
type = "remap"
inputs = ["raw_ovn_audit_logs"]
source = '''
  .tag = ".ovn-audit.log"
'''
`,
		}),
	)
})
