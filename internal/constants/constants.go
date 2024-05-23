package constants

import (
	v1 "k8s.io/api/core/v1"
)

const (

	// Keys used in ClusterLogForwarder.Output Secrets keys.
	// Documented with OutputSpec.Secret in ../../apis/logging/v1/cluster_log_forwarder_types.go
	//
	// WARNING: changing or removing values here is a breaking API change.

	// TLS keys, used by any output that supports TLS.

	ClientCertKey      = "tls.crt"
	ClientPrivateKey   = "tls.key"
	TrustedCABundleKey = "ca-bundle.crt"
	Passphrase         = "passphrase"
	BearerTokenFileKey = "token"

	// Username/Password keys, used by any output with username/password authentication.

	ClientUsername = "username"
	ClientPassword = "password"

	// SASL keys, used by any output that supports SASL.

	SASLEnable        = "sasl.enable"
	SASLMechanisms    = "sasl.mechanisms"
	SASLAllowInsecure = "sasl.allow-insecure"

	// Output-specific keys

	SharedKey                = "shared_key"            // fluent forward
	DeprecatedSaslOverSSL    = "sasl_over_ssl"         // Kafka
	AWSSecretAccessKey       = "aws_secret_access_key" //nolint:gosec
	AWSAccessKeyID           = "aws_access_key_id"
	AWSRoleSessionName       = "cluster-logging" // identifier for role logging session
	AWSCredentialsKey        = "credentials"     // credrequest key to check for sts-formatted secret
	AWSWebIdentityRoleKey    = "role_arn"        // manual key to check for sts-formatted secret
	AWSWebIdentityTokenName  = "collector-sts-token"
	AWSWebIdentityTokenMount = "/var/run/secrets/openshift/serviceaccount" //nolint:gosec // default location for volume mount

	AWSRegionEnvVarKey           = "AWS_REGION"
	AWSRoleArnEnvVarKey          = "AWS_ROLE_ARN"
	AWSRoleSessionEnvVarKey      = "AWS_ROLE_SESSION_NAME"
	AWSWebIdentityTokenEnvVarKey = "AWS_WEB_IDENTITY_TOKEN_FILE" //nolint:gosec

	SplunkHECTokenKey = `hecToken`

	TokenKey          = "token"
	LogCollectorToken = "logcollector-token"

	UnHealthyStatus = "0"
	HealthyStatus   = "1"
	UnManagedStatus = "0"
	ManagedStatus   = "1"
	IsPresent       = "1"
	IsNotPresent    = "0"

	SingletonName = "instance"
	OpenshiftNS   = "openshift-logging"

	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	ElasticsearchFQDN          = "elasticsearch"
	ElasticsearchName          = "elasticsearch"
	ElasticsearchPort          = "9200"
	FluentdName                = "fluentd"
	VectorName                 = "vector"
	KibanaName                 = "kibana"
	LogfilesmetricexporterName = "logfilesmetricexporter"
	LogStoreURL                = "https://" + ElasticsearchFQDN + ":" + ElasticsearchPort
	CollectorSecretName        = "collector"
	PodSecurityLabelEnforce    = "pod-security.kubernetes.io/enforce"
	PodSecurityLabelAudit      = "pod-security.kubernetes.io/audit"
	PodSecurityLabelWarn       = "pod-security.kubernetes.io/warn"
	PodSecurityLabelValue      = "privileged"
	PodSecuritySyncLabel       = "security.openshift.io/scc.podSecurityLabelSync"

	CollectorName               = "collector"
	CollectorConfigSecretName   = "collector-config"
	CollectorMetricSecretName   = "collector-metrics"
	CollectorServiceAccountName = "logcollector"
	CollectorTrustedCAName      = "collector-trusted-ca-bundle"
	CollectorDeploymentKind     = "collector-deployment-kind"
	DeploymentType              = "deployment"

	VectorImageEnvVar             = "RELATED_IMAGE_VECTOR"
	LogfilesmetricImageEnvVar     = "RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER"
	ClusterInfrastructureInstance = "cluster"

	ContainerLogDir = "/var/log/containers"
	PodLogDir       = "/var/log/pods"

	// Annotation Names
	AnnotationServingCertSecretName = "service.beta.openshift.io/serving-cert-secret-name"

	// K8s recommended label names: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
	LabelK8sName      = "app.kubernetes.io/name"       // The name of the application (string)
	LabelK8sInstance  = "app.kubernetes.io/instance"   // A unique name identifying the instance of an application (string)
	LabelK8sVersion   = "app.kubernetes.io/version"    // The current version of the application (e.g., a semantic version, revision hash, etc.) (string)
	LabelK8sComponent = "app.kubernetes.io/component"  // The component within the architecture (string)
	LabelK8sPartOf    = "app.kubernetes.io/part-of"    // The name of a higher level application this one is part of (string)
	LabelK8sManagedBy = "app.kubernetes.io/managed-by" // The tool being used to manage the operation of an application (string)
	LabelK8sCreatedBy = "app.kubernetes.io/created-by" // The controller/user who created this resource (string)

	ClusterLogging         = "cluster-logging"
	ClusterLoggingOperator = "cluster-logging-operator"

	// Commonly-used label names.
	LabelComponent = "component"

	EventReasonReconcilingLoggingCR = "ReconcilingLoggingCR"
	EventReasonCreateObject         = "CreateObject"
	EventReasonGetObject            = "GetObject"
	EventReasonRemoveObject         = "RemoveObject"
	EventReasonUpdateObject         = "UpdateObject"

	MigrateInputPrefix         = "migrateInput"
	MigrateInputApplication    = "migrateInputapplication"
	MigrateInputInfrastructure = "migrateInputinfrastructure"
	MigrateInputAudit          = "migrateInputaudit"
	MigrateDefaultOutput       = "default-replaced"
	ClusterLoggingAvailable    = "isClusterLoggingAvailable"

	OptimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"
	OTELSchema             = "opentelemetry"

	HTTPReceiverPort   = 8443
	HTTPFormat         = "kubeAPIAudit"
	SyslogReceiverPort = 10514

	LabelHTTPInputService   = "http-input-service"
	LabelSyslogInputService = "syslog-input-service"
)

var ReconcileForGlobalProxyList = []string{CollectorTrustedCAName}
var ExtraNoProxyList = []string{ElasticsearchFQDN}

func DefaultTolerations() []v1.Toleration {
	return []v1.Toleration{
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
		{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}
}

const VolumeNameTrustedCA = "trusted-ca"
