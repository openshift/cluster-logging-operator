package constants

import (
	v1 "k8s.io/api/core/v1"
)

const (

	// TLS keys, used by any output that supports TLS.

	ClientCertKey      = "tls.crt"
	ClientPrivateKey   = "tls.key"
	TrustedCABundleKey = "ca-bundle.crt"
	Passphrase         = "passphrase"

	// Username/Password keys, used by any output with username/password authentication.

	ClientUsername = "username"
	ClientPassword = "password"

	// Output-specific keys

	SharedKey                    = "shared_key"            // fluent forward
	AWSSecretAccessKey           = "aws_secret_access_key" //nolint:gosec
	AWSAccessKeyID               = "aws_access_key_id"
	AWSRoleSessionName           = "cluster-logging" // identifier for role logging session
	AWSCredentialsKey            = "credentials"     // credrequest key to check for sts-formatted secret
	AWSWebIdentityRoleKey        = "role_arn"        // manual key to check for sts-formatted secret
	AWSRegionEnvVarKey           = "AWS_REGION"
	AWSRoleArnEnvVarKey          = "AWS_ROLE_ARN"
	AWSRoleSessionEnvVarKey      = "AWS_ROLE_SESSION_NAME"
	AWSWebIdentityTokenEnvVarKey = "AWS_WEB_IDENTITY_TOKEN_FILE" //nolint:gosec

	SplunkHECTokenKey = `hecToken`

	TokenKey          = "token"
	LogCollectorToken = "logcollector-token"

	SingletonName = "instance"
	OpenshiftNS   = "openshift-logging"

	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"

	//ServiceAccountSecretPath is the path to find the projected serviceAccount token and other SA secrets
	ServiceAccountSecretPath   = "/var/run/ocp-collector/serviceaccount"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	ElasticsearchFQDN          = "elasticsearch"
	ElasticsearchName          = "elasticsearch"
	VectorName                 = "vector"
	KibanaName                 = "kibana"
	LogfilesmetricexporterName = "logfilesmetricexporter"
	PodSecurityLabelEnforce    = "pod-security.kubernetes.io/enforce"
	PodSecurityLabelValue      = "privileged"
	// Disable gosec linter, complains "possible hard-coded secret"
	CollectorSecretsDir         = "/var/run/ocp-collector/secrets" //nolint:gosec
	ConfigMapBaseDir            = "/var/run/ocp-collector/config"
	CollectorName               = "collector"
	CollectorConfigSecretName   = "collector-config"
	CollectorMetricSecretName   = "collector-metrics"
	CollectorServiceAccountName = "logcollector"
	CollectorTrustedCAName      = "collector-trusted-ca-bundle"

	VectorImageEnvVar         = "RELATED_IMAGE_VECTOR"
	LogfilesmetricImageEnvVar = "RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER"

	ContainerLogDir = "/var/log/containers"
	PodLogDir       = "/var/log/pods"

	// Annotation Names
	AnnotationServingCertSecretName = "service.beta.openshift.io/serving-cert-secret-name"

	ClusterLogging         = "cluster-logging"
	ClusterLoggingOperator = "cluster-logging-operator"

	OptimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

	HTTPReceiverPort   = 8443
	HTTPFormat         = "kubeAPIAudit"
	SyslogReceiverPort = 10514

	VolumeNameTrustedCA = "trusted-ca"

	STDOUT = "stdout"
	STDERR = "stderr"
)

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
