package functional

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func (f *CollectorFunctionalFramework) WriteMessagesToNamespace(msg, namespace string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], namespace, f.Pod.Name, f.Pod.UID, constants.CollectorName)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}
func (f *CollectorFunctionalFramework) WriteMessagesToApplicationLog(msg string, numOfLogs int) error {
	return f.WriteMessagesToApplicationLogForContainer(msg, constants.CollectorName, numOfLogs)
}
func (f *CollectorFunctionalFramework) WriteMessagesToApplicationLogForContainer(msg, container string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], f.Pod.Namespace, f.Pod.Name, f.Pod.UID, container)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

// WriteMessagesInfraContainerLog mocks writing infra container logs for the functional Framework.  This may require
// enabling the mock api adapter to get metadata for infrastructure logs since the path does not match a pod
// running on the cluster (e.g Framework.VisitConfig = functional.TestAPIAdapterConfigVisitor)
func (f *CollectorFunctionalFramework) WriteMessagesToInfraContainerLog(msg string, numOfLogs int) error {
	ns := "openshift-fake-infra"
	if strings.HasPrefix(f.Namespace, "openshift-test") {
		ns = f.Namespace
	}
	filename := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], ns, f.Pod.Name, f.Pod.UID, constants.CollectorName)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

// WriteMessagesToInfraJournalLog mocks writing infra journal log.  The framework assumes the msg is formatted as
// a JSON entry already parsed like the systemd plugin.  Ex:
// {"PRIORITY":"6","_UID":"1000","_GID":"1000","_CAP_EFFECTIVE":"0","_SELINUX_CONTEXT":"unconfined_u:unconfined_r:unconfined_t:s0-s0:c0.c1023","_AUDIT_SESSION":"3","_AUDIT_LOGINUID":"1000","_SYSTEMD_OWNER_UID":"1000","_SYSTEMD_UNIT":"user@1000.service","_SYSTEMD_SLICE":"user-1000.slice","_MACHINE_ID":"e2a074cafa5044c7a2761b4a97e249ce","_HOSTNAME":"decker","_TRANSPORT":"stdout","_SYSTEMD_USER_SLICE":"app.slice","SYSLOG_IDENTIFIER":"google-chrome.desktop","_COMM":"cat","_EXE":"/usr/bin/cat","_CMDLINE":"cat","MESSAGE":"Error in cpuinfo: failed to parse processor information from /proc/cpuinfo","_BOOT_ID":"40646b056fbe4af6a8b9543864ae0216","_STREAM_ID":"063bc071ac204a37aabc926f2f7614b0","_PID":"3194","_SYSTEMD_CGROUP":"/user.slice/user-1000.slice/user@1000.service/app.slice/app-glib-google\\x2dchrome-3188.scope/3194","_SYSTEMD_USER_UNIT":"app-glib-google\\x2dchrome-3188.scope","_SYSTEMD_INVOCATION_ID":"764ffdafa8b34ac69ec6055d5f942583"}
func (f *CollectorFunctionalFramework) WriteMessagesToInfraJournalLog(msg string, numOfLogs int) error {
	filename := "/var/log/fakejournal/0.log"
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WritesInfraContainerLogs(numOfLogs int) error {
	msg := "2021-03-31T12:59:28.573159188+00:00 stdout F test infra message"
	return f.WriteMessagesToInfraContainerLog(msg, numOfLogs)
}

func (f *CollectorFunctionalFramework) WriteMessagesToAuditLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[auditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WriteAuditHostLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[auditLog])
	msg := NewAuditHostLog(time.Now())
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WriteMessagesTok8sAuditLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[k8sAuditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WriteK8sAuditLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[k8sAuditLog])
	for numOfLogs > 0 {
		entry := NewKubeAuditLog(time.Now())
		if err := f.WriteMessagesToLog(entry, 1, filename); err != nil {
			return err
		}
		numOfLogs -= 1
	}
	return nil
}

func (f *CollectorFunctionalFramework) WriteOpenshiftAuditLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[OpenshiftAuditLog])
	for numOfLogs > 0 {
		now := CRIOTime(time.Now())
		entry := fmt.Sprintf(OpenShiftAuditLogTemplate, now, now)
		if err := f.WriteMessagesToLog(entry, 1, filename); err != nil {
			return err
		}
		numOfLogs -= 1
	}
	return nil
}

func (f *CollectorFunctionalFramework) WriteMessagesToOpenshiftAuditLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[OpenshiftAuditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}
func (f *CollectorFunctionalFramework) WriteMessagesToOAuthAuditLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/audit.log", fluentdLogPath[oauthAuditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WriteMessagesToOVNAuditLog(msg string, numOfLogs int) error {

	filename := fmt.Sprintf("%s/acl-audit-log.log", fluentdLogPath[ovnAuditLog])
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WriteOVNAuditLog(numOfLogs int) error {
	filename := fmt.Sprintf("%s/acl-audit-log.log", fluentdLogPath[ovnAuditLog])
	for numOfLogs > 0 {
		entry := NewOVNAuditLog(time.Now())
		if err := f.WriteMessagesToLog(entry, 1, filename); err != nil {
			return err
		}
		numOfLogs -= 1
	}
	return nil
}

func (f *CollectorFunctionalFramework) WritesApplicationLogs(numOfLogs int) error {
	return f.WritesNApplicationLogsOfSize(numOfLogs, 100, 1)
}

func (f *CollectorFunctionalFramework) WritesApplicationLogsWithDelay(numOfLogs int, delay float32) error {
	return f.WritesNApplicationLogsOfSize(numOfLogs, 100, delay)
}

func (f *CollectorFunctionalFramework) WritesNApplicationLogsOfSize(numOfLogs, size int, delay float32) error {
	msg := "$(date -u +'%Y-%m-%dT%H:%M:%S.%N%:z') stdout F $msg "
	file := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], f.Pod.Namespace, f.Pod.Name, f.Pod.UID, constants.CollectorName)
	logPath := filepath.Dir(file)
	log.V(3).Info("Writing message to app log with path", "path", logPath)
	result, err := f.RunCommand(constants.CollectorName, "bash", "-c", fmt.Sprintf("bash -c 'mkdir -p %s;msg=$(cat /dev/urandom|tr -dc 'a-zA-Z0-9'|fold -w %d|head -n 1);for n in $(seq 1 %d);do echo %s >> %s; sleep %fs; done'", logPath, size, numOfLogs, msg, file, delay))
	log.V(3).Info("WritesNApplicationLogsOfSize", "namespace", f.Pod.Namespace, "result", result, "err", err)
	return err
}

func (f *CollectorFunctionalFramework) WriteMessagesToLog(msg string, numOfLogs int, filename string) error {
	logPath := filepath.Dir(filename)
	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	cmd := fmt.Sprintf("mkdir -p %s;for n in {1..%d};do echo \"$(echo %s|base64 -d)\" >> %s;sleep 1s;done", logPath, numOfLogs, encoded, filename)
	log.V(3).Info("Writing messages to log with command", "cmd", cmd)
	result, err := f.RunCommand(constants.CollectorName, "bash", "-c", cmd)
	log.V(3).Info("WriteMessagesToLog", "namespace", f.Pod.Namespace, "result", result, "err", err)
	return err
}

// WriteMessagesWithNotUTF8SymbolsToLog write 12 symbols in ISO-8859-1 encoding
// need to use small hack with 'sed' replacement because if try to use something like:
// 'echo -e \xC0\xC1' Go always convert every undecodeable byte into '\ufffd'.
// More details here: https://github.com/golang/go/issues/38006
func (f *CollectorFunctionalFramework) WriteMessagesWithNotUTF8SymbolsToLog() error {
	filename := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], f.Pod.Namespace, f.Pod.Name,
		f.Pod.UID, constants.CollectorName)
	logPath := filepath.Dir(filename)
	cmd := fmt.Sprintf("mkdir -p %s; echo -e \"$(echo '%s stdout F yC0yC1yF5yF6yF7yF8yF9yFAyFByFCyFDyFE' | sed -r 's/y/\\\\x/g')\"  >> %s;",
		logPath, CRIOTime(time.Now()), filename)
	log.V(3).Info("Writing messages to log with command", "cmd", cmd)
	result, err := f.RunCommand(constants.CollectorName, "bash", "-c", cmd)
	log.V(3).Info("WriteMessagesWithNotUTF8SymbolsToLog", "namespace", f.Pod.Namespace, "result", result, "err", err)
	return err
}
