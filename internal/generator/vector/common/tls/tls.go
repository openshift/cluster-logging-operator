package tls

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	ExcludeInsecureSkipVerify = "ExcludeInsecureSkipVerify"
)

var (
	emptyTLSConf = transport.TLS{}
)

func NewTlsEnabled(comp observability.TransportLayerSecurity, secrets observability.Secrets, op utils.Options, options ...framework.Option) *transport.TlsEnabled {
	tls := NewTls(comp, secrets, op, options...)
	if tls != nil && tls.CRTFile != "" {
		return &transport.TlsEnabled{
			TLS:     *tls,
			Enabled: true,
		}
	}
	return nil
}
func NewTls(comp observability.TransportLayerSecurity, secrets observability.Secrets, op utils.Options, options ...framework.Option) (conf *transport.TLS) {
	if outURL, found := framework.HasOption(framework.URL, options); found {
		if !url.IsSecure(outURL.(string)) {
			return nil
		}
	}
	conf = &transport.TLS{}
	if comp != nil && comp.GetTlsSpec() != nil {
		spec := comp.GetTlsSpec()
		conf.CAFile = ValuePath(spec.CA, "%s")
		conf.CRTFile = ValuePath(spec.Certificate, "%s")
		conf.KeyFile = SecretPath(spec.Key, "%s")
		conf.KeyPass = secrets.AsString(spec.KeyPassphrase)
		if _, found := framework.HasOption(ExcludeInsecureSkipVerify, options); !found && comp.IsInsecureSkipVerify() {
			conf.VerifyCertificate = utils.GetPtr(false)
			conf.VerifyHostname = utils.GetPtr(false)
		}
	}
	SetTLSProfile(conf, op)
	if *conf == emptyTLSConf {
		return nil
	}
	return conf
}

// SetTLSProfile updates the tls and cipher specs from the options given
// TODO: Remove internal/generator/vector/output/common/tls
func SetTLSProfile(t *transport.TLS, op utils.Options) *transport.TLS {
	if version, found := op[framework.MinTLSVersion]; found {
		t.MinTlsVersion = version.(string)
	}
	if ciphers, found := op[framework.Ciphers]; found {
		t.CipherSuites = ciphers.(string)
	}
	return t
}

func ValuePath(resource *obs.ValueReference, formatter ...string) string {
	if resource == nil {
		return ""
	}
	if resource.SecretName != "" {
		return helpers.SecretPath(resource.SecretName, resource.Key, formatter...)
	} else if resource.ConfigMapName != "" {
		return helpers.ConfigPath(resource.ConfigMapName, resource.Key, formatter...)
	}
	return ""
}

func SecretPath(resource *obs.SecretReference, formatter ...string) string {
	if resource == nil || resource.SecretName == "" {
		return ""
	}
	return helpers.SecretPath(resource.SecretName, resource.Key, formatter...)
}
