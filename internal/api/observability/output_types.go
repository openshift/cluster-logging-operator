package observability

import (
	"fmt"
	"os"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/utils/set"
)

func OutputTypeUnknown(t obsv1.OutputType) error {
	return fmt.Errorf("unknown output type %q", t)
}

type Outputs []obsv1.OutputSpec

// Names returns a slice of output names
func (outputs Outputs) Names() (names []string) {
	for _, o := range outputs {
		names = append(names, o.Name)
	}
	return names
}

// Map returns a map of output name to output spec
func (outputs Outputs) Map() map[string]obsv1.OutputSpec {
	m := map[string]obsv1.OutputSpec{}
	for _, o := range outputs {
		m[o.Name] = o
	}
	return m
}

// ConfigmapNames returns a unique set of unordered configmap names
func (outputs Outputs) ConfigmapNames() []string {
	names := set.New[string]()
	for _, o := range outputs {
		if o.TLS != nil {
			names.Insert(ConfigmapsForTLS(o.TLS.TLSSpec)...)
		}
	}
	return names.UnsortedList()
}

// NeedServiceAccountToken returns true if any output needs to be configured to use a projected service account token
func (outputs Outputs) NeedServiceAccountToken() bool {
	for _, o := range outputs {
		if needsServiceAccountToken(o) {
			return true
		}
	}
	return false
}

// needsServiceAccountToken returns true if the output requires service account token projection
func needsServiceAccountToken(o obsv1.OutputSpec) bool {
	// For GCP we always project the service account token
	if o.Type == obsv1.OutputTypeGoogleCloudLogging && o.GoogleCloudLogging != nil && o.GoogleCloudLogging.Authentication != nil {
		return true
	}

	// If any output has explicit service account token config
	token := getOutputBearerToken(o)
	return token != nil && token.From == obsv1.BearerTokenFromServiceAccount
}

// getOutputBearerToken extracts the bearer token configuration from an output, if present
func getOutputBearerToken(o obsv1.OutputSpec) *obsv1.BearerToken {
	switch o.Type {
	case obsv1.OutputTypeLoki:
		if o.Loki != nil && o.Loki.Authentication != nil {
			return o.Loki.Authentication.Token
		}
	case obsv1.OutputTypeLokiStack:
		if o.LokiStack != nil && o.LokiStack.Authentication != nil {
			return o.LokiStack.Authentication.Token
		}
	case obsv1.OutputTypeCloudwatch:
		if o.Cloudwatch != nil && o.Cloudwatch.Authentication.Type == obsv1.AwsAuthTypeIAMRole {
			return &o.Cloudwatch.Authentication.IamRole.Token
		}
	case obsv1.OutputTypeS3:
		if o.S3 != nil && o.S3.Authentication.Type == obsv1.AwsAuthTypeIAMRole {
			return &o.S3.Authentication.IamRole.Token
		}
	case obsv1.OutputTypeElasticsearch:
		if o.Elasticsearch != nil && o.Elasticsearch.Authentication != nil {
			return o.Elasticsearch.Authentication.Token
		}
	case obsv1.OutputTypeOTLP:
		if o.OTLP != nil && o.OTLP.Authentication != nil {
			return o.OTLP.Authentication.Token
		}
	}
	return nil
}

// SecretNames returns a unique set of unordered secret names
func (outputs Outputs) SecretNames() []string {
	secrets := set.New[string]()
	for _, o := range outputs {
		if o.TLS != nil {
			secrets.Insert(SecretsForTLS(o.TLS.TLSSpec)...)
		}
		keys := SecretReferences(o)
		for _, k := range keys {
			if k != nil {
				secrets.Insert(k.SecretName)
			}
		}
	}
	return secrets.UnsortedList()
}

func SecretReferencesAsValueReferences(o obsv1.OutputSpec) (configs []*obsv1.ValueReference) {
	for _, auth := range SecretReferences(o) {
		if auth != nil {
			configs = append(configs, &obsv1.ValueReference{
				Key:        auth.Key,
				SecretName: auth.SecretName,
			})
		}
	}
	return configs
}

// SecretReferences returns a list of the keys associated with an output.  It is possible for a list entry
// to be nil if it was not specified for the output
func SecretReferences(o obsv1.OutputSpec) []*obsv1.SecretReference {
	switch o.Type {
	case obsv1.OutputTypeAzureMonitor:
		if o.AzureMonitor != nil && o.AzureMonitor.Authentication != nil {
			return []*obsv1.SecretReference{o.AzureMonitor.Authentication.SharedKey}
		}
	case obsv1.OutputTypeCloudwatch:
		if o.Cloudwatch != nil && o.Cloudwatch.Authentication != nil {
			a := o.Cloudwatch.Authentication
			return awsSecretKeys(a)
		}
	case obsv1.OutputTypeS3:
		if o.S3 != nil && o.S3.Authentication != nil {
			a := o.S3.Authentication
			return awsSecretKeys(a)
		}
	case obsv1.OutputTypeElasticsearch:
		if o.Elasticsearch != nil && o.Elasticsearch.Authentication != nil {
			return httpAuthKeys(o.Elasticsearch.Authentication)
		}
	case obsv1.OutputTypeGoogleCloudLogging:
		if o.GoogleCloudLogging != nil && o.GoogleCloudLogging.Authentication != nil {
			auth := o.GoogleCloudLogging.Authentication
			if auth.Credentials != nil {
				return []*obsv1.SecretReference{auth.Credentials}
			}
		}
	case obsv1.OutputTypeHTTP:
		if o.HTTP != nil && o.HTTP.Authentication != nil {
			return httpAuthKeys(o.HTTP.Authentication)
		}
	case obsv1.OutputTypeOTLP:
		if o.OTLP != nil && o.OTLP.Authentication != nil {
			return httpAuthKeys(o.OTLP.Authentication)
		}
	case obsv1.OutputTypeKafka:
		if o.Kafka != nil && o.Kafka.Authentication != nil {
			a := o.Kafka.Authentication
			return []*obsv1.SecretReference{a.SASL.Password, a.SASL.Username}
		}
	case obsv1.OutputTypeLoki:
		if o.Loki != nil {
			return httpAuthKeys(o.Loki.Authentication)
		}
	case obsv1.OutputTypeLokiStack:
		if o.LokiStack != nil {
			return lokiStackKeys(o.LokiStack.Authentication)
		}
	case obsv1.OutputTypeSplunk:
		if o.Splunk != nil && o.Splunk.Authentication != nil {
			return []*obsv1.SecretReference{o.Splunk.Authentication.Token}
		}
	case obsv1.OutputTypeSyslog:
	default:
		log.V(0).Error(OutputTypeUnknown(o.Type), "Found unsupported output type while gathering secret names")
		os.Exit(1)
	}
	return []*obsv1.SecretReference{}
}

func httpAuthKeys(auth *obsv1.HTTPAuthentication) []*obsv1.SecretReference {
	if auth != nil {
		keys := []*obsv1.SecretReference{
			auth.Username,
			auth.Password,
		}
		if auth.Token != nil && auth.Token.From == obsv1.BearerTokenFromSecret && auth.Token.Secret != nil {
			keys = append(keys, &obsv1.SecretReference{
				Key:        auth.Token.Secret.Key,
				SecretName: auth.Token.Secret.Name,
			})
		}
		return keys
	}
	return []*obsv1.SecretReference{}
}

func lokiStackKeys(auth *obsv1.LokiStackAuthentication) (keys []*obsv1.SecretReference) {
	if auth != nil {
		if auth.Token != nil && auth.Token.From == obsv1.BearerTokenFromSecret && auth.Token.Secret != nil {
			keys = append(keys, &obsv1.SecretReference{
				Key:        auth.Token.Secret.Key,
				SecretName: auth.Token.Secret.Name,
			})
		}
	}
	return keys
}

// awsSecretKeys returns a list of keys from secrets in the s3 and cloudwatch outputs
func awsSecretKeys(auth *obsv1.AwsAuthentication) (keys []*obsv1.SecretReference) {
	if auth == nil {
		return keys
	}
	switch auth.Type {
	case obsv1.AwsAuthTypeAccessKey:
		keys = append(keys, &auth.AwsAccessKey.KeyId, &auth.AwsAccessKey.KeySecret)
	case obsv1.AwsAuthTypeIAMRole:
		keys = append(keys, &auth.IamRole.RoleARN)
		if auth.IamRole.Token.From == obsv1.BearerTokenFromSecret && auth.IamRole.Token.Secret != nil {
			keys = append(keys, &obsv1.SecretReference{
				Key:        auth.IamRole.Token.Secret.Key,
				SecretName: auth.IamRole.Token.Secret.Name,
			})
		}
	}
	keys = appendAssumeRoleKeys(auth, keys)
	return keys
}

// appendAssumeRoleKeys adds assume role secret keys to the refs
func appendAssumeRoleKeys(auth *obsv1.AwsAuthentication, keys []*obsv1.SecretReference) []*obsv1.SecretReference {
	if auth != nil && auth.AssumeRole != nil {
		keys = append(keys, &auth.AssumeRole.RoleARN)
	}
	return keys
}
