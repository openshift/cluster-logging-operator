package tls

import (
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
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
	if tls != nil && (tls.CAFile != "" || (tls.CRTFile != "" && tls.KeyFile != "")) {
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
		conf.CAFile = helpers.ValuePath(spec.CA, "%s")
		conf.CRTFile = helpers.ValuePath(spec.Certificate, "%s")
		conf.KeyFile = helpers.SecretPath(spec.Key, "%s")
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

// SetTLSProfile updates the tls, cipher, and curve specs from the options given
func SetTLSProfile(t *transport.TLS, op utils.Options) *transport.TLS {
	if version, found := op[framework.MinTLSVersion]; found {
		t.MinTlsVersion = version.(string)
	}
	if ciphers, found := op[framework.Ciphers]; found {
		t.CipherSuites = ciphers.(string)
	}
	if groups, found := op[framework.Groups]; found {
		t.Curves = groups.(string)
	}
	return t
}
