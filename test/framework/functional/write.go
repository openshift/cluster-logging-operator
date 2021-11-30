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
	filename := fmt.Sprintf("%s/%s_%s_%s-%s.log", fluentdLogPath[applicationLog], f.Pod.Name, f.Pod.Namespace, constants.CollectorName, f.fluentContainerId)
	return f.WriteMessagesToLog(msg, numOfLogs, filename)
}

func (f *CollectorFunctionalFramework) WriteMessagesInfraContainerLog(msg string, numOfLogs int) error {
	filename := fmt.Sprintf("%s/%s_%s_%s-%s.log", fluentdLogPath[applicationLog], f.Pod.Name, "openshift-fake-infra", constants.FluentdName, f.fluentContainerId)
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
	//podname_ns_containername-containerid.log
	//functional_testhack-16511862744968_fluentd-90a0f0a7578d254eec640f08dd155cc2184178e793d0289dff4e7772757bb4f8.log
	filepath := fmt.Sprintf("/var/log/containers/%s_%s_%s-%s.log", f.Pod.Name, f.Pod.Namespace, constants.CollectorName, f.fluentContainerId)
	result, err := f.RunCommand(constants.CollectorName, "bash", "-c", fmt.Sprintf("bash -c 'mkdir -p /var/log/containers;echo > %s;msg=$(cat /dev/urandom|tr -dc 'a-zA-Z0-9'|fold -w %d|head -n 1);for n in $(seq 1 %d);do echo %s >> %s; done'", filepath, size, numOfLogs, msg, filepath))
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
