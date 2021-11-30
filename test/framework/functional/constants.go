package functional

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"time"
)

const (
	applicationLog     = "application"
	auditLog           = "audit"
	ovnAuditLog        = "ovn"
	k8sAuditLog        = "k8s"
	infraLog           = "infra"
	OpenshiftAuditLog  = "openshift-audit-logs"
	ApplicationLogFile = "/tmp/app-logs"
	FunctionalNodeName = "functional-test-node"
)

var (
	maxDuration          time.Duration
	defaultRetryInterval time.Duration

	fluentdLogPath = map[string]string{
		applicationLog:    "/var/log/containers",
		auditLog:          "/var/log/audit",
		ovnAuditLog:       "/var/log/ovn",
		OpenshiftAuditLog: "/var/log/oauth-apiserver",
		k8sAuditLog:       "/var/log/kube-apiserver",
	}
	outputLogFile = map[string]map[string]string{
		logging.OutputTypeFluentdForward: {
			applicationLog: ApplicationLogFile,
			auditLog:       "/tmp/audit-logs",
			ovnAuditLog:    "/tmp/audit-logs",
			k8sAuditLog:    "/tmp/audit-logs",
			infraLog:       "/tmp/infra-logs",
		},
		logging.OutputTypeSyslog: {
			applicationLog: "/var/log/infra.log",
			auditLog:       "/var/log/infra.log",
			k8sAuditLog:    "/var/log/infra.log",
			ovnAuditLog:    "/var/log/infra.log",
			infraLog:       "/var/log/infra.log",
		},
		logging.OutputTypeKafka: {
			applicationLog: "/var/log/app.log",
			auditLog:       "/var/log/infra.log",
			k8sAuditLog:    "/var/log/audit.log",
			ovnAuditLog:    "/var/log/ovnaudit.log",
		},
	}
)

func init() {
	maxDuration, _ = time.ParseDuration("5m")
	defaultRetryInterval, _ = time.ParseDuration("10s")
}
