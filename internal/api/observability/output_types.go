package observability

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/utils/set"
	"os"
)

func OutputTypeUnknown(t obsv1.OutputType) error {
	return fmt.Errorf("Unknown output type %q", t)
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

// NeedServiceAccountToken returns true if any output needs to be configured to use the token associated with the service account
func (outputs Outputs) NeedServiceAccountToken() bool {
	var auths []*obsv1.BearerToken
	for _, o := range outputs {
		switch {
		case o.Type == obsv1.OutputTypeLoki && o.Loki.Authentication != nil && o.Loki.Authentication.Token != nil:
			auths = append(auths, o.Loki.Authentication.Token)
		case o.Type == obsv1.OutputTypeLokiStack && o.LokiStack.Authentication != nil && o.LokiStack.Authentication.Token != nil:
			auths = append(auths, o.LokiStack.Authentication.Token)
		case o.Type == obsv1.OutputTypeCloudwatch && o.Cloudwatch != nil && o.Cloudwatch.Authentication.Type == obsv1.CloudwatchAuthTypeIAMRole:
			auths = append(auths, &o.Cloudwatch.Authentication.IAMRole.Token)
		case o.Type == obsv1.OutputTypeElasticsearch && o.Elasticsearch != nil && o.Elasticsearch.Authentication != nil && o.Elasticsearch.Authentication.Token != nil:
			auths = append(auths, o.Elasticsearch.Authentication.Token)
		case o.Type == obsv1.OutputTypeOTLP && o.OTLP.Authentication != nil && o.OTLP.Authentication.Token != nil:
			auths = append(auths, o.OTLP.Authentication.Token)
		}
	}
	for _, token := range auths {
		if token.From == obsv1.BearerTokenFromServiceAccount {
			return true
		}
	}

	return false
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
			keys := cloudwatchSecretKeys(a)
			// cross-account feature LOG-7687
			return appendAssumeRoleKeys(a, keys)
		}
	case obsv1.OutputTypeElasticsearch:
		if o.Elasticsearch != nil && o.Elasticsearch.Authentication != nil {
			return httpAuthKeys(o.Elasticsearch.Authentication)
		}
	case obsv1.OutputTypeGoogleCloudLogging:
		if o.GoogleCloudLogging != nil && o.GoogleCloudLogging.Authentication != nil {
			a := o.GoogleCloudLogging.Authentication
			return []*obsv1.SecretReference{a.Credentials}
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
		log.V(3).Error(OutputTypeUnknown(o.Type), "Found unsupported output type while gathering secret names")
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

// cloudwatchSecretKeys returns a list of keys from secrets in the cloudwatch output
func cloudwatchSecretKeys(auth *obsv1.CloudwatchAuthentication) (keys []*obsv1.SecretReference) {
	if auth == nil {
		return keys
	}
	switch auth.Type {
	case obsv1.CloudwatchAuthTypeAccessKey:
		keys = append(keys, &auth.AWSAccessKey.KeyId, &auth.AWSAccessKey.KeySecret)
	case obsv1.CloudwatchAuthTypeIAMRole:
		keys = append(keys, &auth.IAMRole.RoleARN)
		if auth.IAMRole.Token.From == obsv1.BearerTokenFromSecret && auth.IAMRole.Token.Secret != nil {
			keys = append(keys, &obsv1.SecretReference{
				Key:        auth.IAMRole.Token.Secret.Key,
				SecretName: auth.IAMRole.Token.Secret.Name,
			})
		}
	}
	return keys
}

// appendAssumeRoleKeys adds assume role spec to the list of secret refs
func appendAssumeRoleKeys(auth *obsv1.CloudwatchAuthentication, keys []*obsv1.SecretReference) []*obsv1.SecretReference {
	if auth != nil && auth.AssumeRole != nil {
		keys = append(keys, &auth.AssumeRole.RoleARN)
	}
	return keys
}
