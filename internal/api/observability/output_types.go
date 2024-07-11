package observability

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/set"
	"os"
)

func OutputTypeUnknown(t obsv1.OutputType) error {
	return fmt.Errorf("Unknown output type %q", t)
}

type Outputs []obsv1.OutputSpec

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
		case o.Type == obsv1.OutputTypeCloudwatch && o.Cloudwatch != nil && o.Cloudwatch.Authentication.Type == obsv1.CloudwatchAuthTypeIAMRole && o.Cloudwatch.Authentication.IAMRole.Token != nil:
			auths = append(auths, o.Cloudwatch.Authentication.IAMRole.Token)
		case o.Type == obsv1.OutputTypeElasticsearch && o.Elasticsearch != nil && o.Elasticsearch.Authentication != nil && o.Elasticsearch.Authentication.Token != nil:
			auths = append(auths, o.Elasticsearch.Authentication.Token)
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
		keys := SecretKeys(o)
		for _, k := range keys {
			if k != nil {
				secrets.Insert(k.Secret.Name)
			}
		}
	}
	return secrets.UnsortedList()
}

// SecretKeysAsConfigMapOrSecretKeys
func SecretKeysAsConfigMapOrSecretKeys(o obsv1.OutputSpec) (configs []*obsv1.ConfigMapOrSecretKey) {
	for _, auth := range SecretKeys(o) {
		if auth != nil {
			configs = append(configs, &obsv1.ConfigMapOrSecretKey{
				Key:    auth.Key,
				Secret: auth.Secret,
			})
		}
	}
	return configs
}

// SecretKeys returns a list of the keys associated with an output.  It is possible for a list entry
// to be nil if it was not specified for the output
func SecretKeys(o obsv1.OutputSpec) []*obsv1.SecretKey {
	switch o.Type {
	case obsv1.OutputTypeAzureMonitor:
		if o.AzureMonitor != nil && o.AzureMonitor.Authentication != nil {
			return []*obsv1.SecretKey{o.AzureMonitor.Authentication.SharedKey}
		}
	case obsv1.OutputTypeCloudwatch:
		if o.Cloudwatch != nil && o.Cloudwatch.Authentication != nil {
			a := o.Cloudwatch.Authentication
			return cloudwatchAuthKeys(a)
		}
	case obsv1.OutputTypeElasticsearch:
		if o.Elasticsearch != nil && o.Elasticsearch.Authentication != nil {
			return httpAuthKeys(o.Elasticsearch.Authentication)
		}
	case obsv1.OutputTypeGoogleCloudLogging:
		if o.GoogleCloudLogging != nil && o.GoogleCloudLogging.Authentication != nil {
			a := o.GoogleCloudLogging.Authentication
			return []*obsv1.SecretKey{a.Credentials}
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
			return []*obsv1.SecretKey{a.SASL.Password, a.SASL.Username}
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
			return []*obsv1.SecretKey{o.Splunk.Authentication.Token}
		}
	case obsv1.OutputTypeSyslog:
	default:
		log.V(0).Error(OutputTypeUnknown(o.Type), "Found unsupported output type while gathering secret names")
		os.Exit(1)
	}
	return []*obsv1.SecretKey{}
}

func httpAuthKeys(auth *obsv1.HTTPAuthentication) []*obsv1.SecretKey {
	if auth != nil {
		keys := []*obsv1.SecretKey{
			auth.Username,
			auth.Password,
		}
		if auth.Token != nil && auth.Token.From == obsv1.BearerTokenFromSecret && auth.Token.Secret != nil {
			keys = append(keys, &obsv1.SecretKey{
				Key: auth.Token.Secret.Key,
				Secret: &v1.LocalObjectReference{
					Name: auth.Token.Secret.Name,
				},
			})
		}
		return keys
	}
	return []*obsv1.SecretKey{}
}

func lokiStackKeys(auth *obsv1.LokiStackAuthentication) (keys []*obsv1.SecretKey) {
	if auth != nil {
		if auth.Token != nil && auth.Token.From == obsv1.BearerTokenFromSecret && auth.Token.Secret != nil {
			keys = append(keys, &obsv1.SecretKey{
				Key: auth.Token.Secret.Key,
				Secret: &v1.LocalObjectReference{
					Name: auth.Token.Secret.Name,
				},
			})
		}
	}
	return keys
}

func cloudwatchAuthKeys(auth *obsv1.CloudwatchAuthentication) (keys []*obsv1.SecretKey) {
	if auth != nil {
		if auth.AWSAccessKey != nil {
			keys = append(keys, auth.AWSAccessKey.KeyID, auth.AWSAccessKey.KeySecret)
		}
		if auth.IAMRole != nil {
			keys = append(keys, auth.IAMRole.RoleARN)
			if auth.IAMRole.Token != nil && auth.IAMRole.Token.From == obsv1.BearerTokenFromSecret && auth.IAMRole.Token.Secret != nil {
				keys = append(keys, &obsv1.SecretKey{
					Key: auth.IAMRole.Token.Secret.Key,
					Secret: &v1.LocalObjectReference{
						Name: auth.IAMRole.Token.Secret.Name,
					},
				})
			}
		}
	}
	return keys
}
