package functional

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"time"
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

	fileLogPaths = map[string]string{
		applicationLog:    ApplicationLogDir,
		auditLog:          AuditLogDir,
		ovnAuditLog:       OvnAuditLogDir,
		oauthAuditLog:     OauthAuditLogDir,
		OpenshiftAuditLog: OpenshiftAuditLogDir,
		k8sAuditLog:       K8sAuditLogDir,
	}
	outputLogFile = map[string]map[string]string{
		string(obs.OutputTypeHTTP): {
			string(obs.InputTypeApplication):    ApplicationLogFile,
			string(obs.InputTypeAudit):          ApplicationLogFile,
			string(obs.InputTypeInfrastructure): ApplicationLogFile,
		},
		string(obs.OutputTypeOTLP): {
			string(obs.InputTypeApplication):    ApplicationLogFile,
			string(obs.InputTypeAudit):          ApplicationLogFile,
			string(obs.InputTypeInfrastructure): ApplicationLogFile,
		},
		string(obs.OutputTypeSyslog): {
			applicationLog:                      "/tmp/infra.log",
			auditLog:                            "/tmp/infra.log",
			k8sAuditLog:                         "/tmp/infra.log",
			ovnAuditLog:                         "/tmp/infra.log",
			string(obs.InputTypeInfrastructure): "/tmp/infra.log",
		},
		string(obs.OutputTypeKafka): {
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
