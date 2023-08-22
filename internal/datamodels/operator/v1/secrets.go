package v1

// The following struct exists solely for the purpose of being able to dynamically generate and publish an API of the
// opinionated keys supported by ClusterLogForwarder.  It is in no way intended to be used for development.

// +kubebuilder:object:root=true
// ClusterLogForwarderSecret provides the set of supported keys that are recognized by the operator when reconciling
// the secret associated with a ClusterLogForwarder output.  These keys are not supported for every output type.  See
// the individual output documentation for supported authentication mechanisms.
type ClusterLogForwarderSecret struct {

	//A client public key
	ClientCertKey string `json:"tls.crt"`
	//A client private key
	ClientPrivateKey string `json:"tls.key"`
	//A Certificate Authority bundle
	TrustedCABundleKey string `json:"ca-bundle.crt"`
	//The TLS passphrase
	Passphrase string `json:"passphrase"`
	//A bearer token
	BearerTokenFileKey string `json:"token"`
	//A user name
	ClientUsername string `json:"username"`
	//A password
	ClientPassword string `json:"password"`
	//Enable SASL
	SASLEnable string `json:"sasl.enable"`
	//SASL mechanisms
	SASLMechanisms string `json:"sasl.mechanisms"`
	//SASL over SSL
	DeprecatedSaslOverSSL string `json:"sasl_over_ssl"`
	//A shared key
	SharedKey string `json:"shared_key"`
	//An AWS secret access key
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
	//An AWS access key ID
	AWSAccessKeyID string `json:"aws_access_key_id"`
	//An AWS credentials key
	AWSCredentialsKey string `json:"credentials"`
	//An AWS role ARN
	AWSWebIdentityRoleKey string `json:"role_arn"`
	//The HEC token for authorizing against a Splunk endpoint
	SplunkHecToken string `json:"hecToken"`
	//The Google application credentials JSON
	GoogleAppCredentials string `json:"google-application-credentials.json"`
}
