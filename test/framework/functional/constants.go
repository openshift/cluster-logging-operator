package functional

import (
	"time"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

const (
	applicationLog        = "application"
	auditLog              = "audit"
	ovnAuditLog           = "ovn"
	k8sAuditLog           = "k8s"
	oauthAuditLog         = "oauth-audit-logs"
	OpenshiftAuditLog     = "openshift-audit-logs"
	ApplicationLogFile    = "/tmp/app-logs"
	InfrastructureLogFile = "/tmp/infra.log"
	FunctionalNodeName    = "functional-test-node"

	ApplicationLogDir    = "/var/log/pods"
	AuditLogDir          = "/var/log/audit"
	OvnAuditLogDir       = "/var/log/ovn"
	OauthAuditLogDir     = "/var/log/oauth-apiserver"
	OpenshiftAuditLogDir = "/var/log/openshift-apiserver"
	K8sAuditLogDir       = "/var/log/kube-apiserver"
)

var (
	maxDuration          time.Duration
	defaultRetryInterval time.Duration

	fluentdLogPath = map[string]string{
		applicationLog:    ApplicationLogDir,
		auditLog:          AuditLogDir,
		ovnAuditLog:       OvnAuditLogDir,
		oauthAuditLog:     OauthAuditLogDir,
		OpenshiftAuditLog: OpenshiftAuditLogDir,
		k8sAuditLog:       K8sAuditLogDir,
	}
	outputLogFile = map[string]map[string]string{
		logging.OutputTypeHttp: {
			logging.InputNameApplication:    ApplicationLogFile,
			logging.InputNameAudit:          ApplicationLogFile,
			logging.InputNameInfrastructure: ApplicationLogFile,
		},
		logging.OutputTypeFluentdForward: {
			applicationLog:                  ApplicationLogFile,
			auditLog:                        "/tmp/audit-logs",
			ovnAuditLog:                     "/tmp/audit-logs",
			k8sAuditLog:                     "/tmp/audit-logs",
			logging.InputNameInfrastructure: "/tmp/infra-logs",
		},
		logging.OutputTypeSyslog: {
			applicationLog:                  "/tmp/infra.log",
			auditLog:                        "/tmp/infra.log",
			k8sAuditLog:                     "/tmp/infra.log",
			ovnAuditLog:                     "/tmp/infra.log",
			logging.InputNameInfrastructure: "/tmp/infra.log",
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
