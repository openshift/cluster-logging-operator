package functional

import (
	"encoding/base64"
	"fmt"
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"path/filepath"
	"time"
)

func (f *CollectorFunctionalFramework) WriteMessagesToApplicationLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], f.Pod.Namespace, f.Pod.Name, f.Pod.UID, constants.CollectorName)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WriteMessagesInfraContainerLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], "openshift-fake-infra", f.Pod.Name, f.Pod.UID, constants.CollectorName)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
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
	return f.WritesNApplicationLogsOfSize(numOfLogs, 100)
}

func (f *CollectorFunctionalFramework) WritesNApplicationLogsOfSize(numOfLogs, size int) error {
	msg := "$(date -u +'%Y-%m-%dT%H:%M:%S.%N%:z') stdout F $msg "
	file := fmt.Sprintf("%s/%s_%s_%s/%s/0.log", fluentdLogPath[applicationLog], f.Pod.Namespace, f.Pod.Name, f.Pod.UID, constants.CollectorName)
	logPath := filepath.Dir(file)
	result, err := f.RunCommand(constants.CollectorName, "bash", "-c", fmt.Sprintf("bash -c 'mkdir -p %s;msg=$(cat /dev/urandom|tr -dc 'a-zA-Z0-9'|fold -w %d|head -n 1);for n in $(seq 1 %d);do echo %s >> %s; done'", logPath, size, numOfLogs, msg, file))
	log.V(3).Info("CollectorFunctionalFramework.WritesNApplicationLogsOfSize", "result", result, "err", err)
	return err
}

func (f *CollectorFunctionalFramework) WriteMessagesToLog(msg string, numOfLogs int, filename string) error {
	logPath := filepath.Dir(filename)
	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	cmd := fmt.Sprintf("mkdir -p %s;for n in {1..%d};do echo \"$(echo %s|base64 -d)\" >> %s;sleep 1s;done", logPath, numOfLogs, encoded, filename)
	log.V(3).Info("Writing messages to log with command", "cmd", cmd)
	result, err := f.RunCommand(constants.CollectorName, "bash", "-c", cmd)
	log.V(3).Info("CollectorFunctionalFramework.WriteMessagesToLog", "result", result, "err", err)
	return err
}
