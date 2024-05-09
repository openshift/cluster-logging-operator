package common

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
	"net/url"
)

type UserNamePass struct {
	Username string
	Password string
}

type SharedKey struct {
	Key string
}

type Passphrase struct {
	KeyPass string
}

type BearerToken struct {
	Token string
}

func TLS(id string, o logging.OutputSpec, secret *corev1.Secret, op framework.Options) []framework.Element {
	if o.Secret != nil || (o.TLS != nil && o.TLS.InsecureSkipVerify) {
		if tlsConf := GenerateTLSConfWithID(id, o, secret, op, false); tlsConf != nil {
			tlsConf.NeedsEnabled = false
			return []framework.Element{tlsConf}
		}
	}
	return []framework.Element{}
}

type TLSConf struct {
	ComponentID        string
	NeedsEnabled       bool
	InsecureSkipVerify bool
	TlsMinVersion      string
	CipherSuites       string
	CAFilePath         string
	CertPath           string
	KeyPath            string
	PassPhrase         string
}

func NewTLSConf(id string, o logging.OutputSpec, op framework.Options) TLSConf {
	conf := TLSConf{
		ComponentID:        id,
		NeedsEnabled:       true,
		InsecureSkipVerify: o.TLS != nil && o.TLS.InsecureSkipVerify,
	}
	conf.SetTLSProfileFromOptions(op)
	return conf
}

func (conf *TLSConf) SetTLSProfileFromOptions(op framework.Options) {
	if version, found := op[framework.MinTLSVersion]; found {
		conf.TlsMinVersion = version.(string)
	}
	if ciphers, found := op[framework.Ciphers]; found {
		conf.CipherSuites = ciphers.(string)
	}
	fmt.Println(conf)
}

func addTLSSettings(o logging.OutputSpec, secret *corev1.Secret, conf *TLSConf) bool {
	addTLS := false
	if o.Secret != nil && (o.Name == logging.OutputNameDefault || HasTLSCertAndKey(secret)) {
		addTLS = true
		conf.CertPath = SecretPath(o.Secret.Name, constants.ClientCertKey)
		conf.KeyPath = SecretPath(o.Secret.Name, constants.ClientPrivateKey)
	}

	if o.Secret != nil && (o.Name == logging.OutputNameDefault || HasCABundle(secret)) {
		addTLS = true
		conf.CAFilePath = SecretPath(o.Secret.Name, constants.TrustedCABundleKey)
	}

	if HasPassphrase(secret) {
		addTLS = true
		conf.PassPhrase = security.GetFromSecret(secret, constants.Passphrase)
	}
	if conf.TlsMinVersion != "" || conf.CipherSuites != "" {
		addTLS = true
	}
	if conf.InsecureSkipVerify {
		addTLS = true
	}

	return addTLS
}

func (t TLSConf) Name() string {
	return "vectorTLS"
}

func (t TLSConf) Template() string {
	if !t.NeedsEnabled {
		return `{{define "vectorTLS" -}}{{end}}`
	}
	return `
{{define "vectorTLS" -}}
[sinks.{{.ComponentID}}.tls]
{{- if ne .TlsMinVersion "" }}
min_tls_version = "{{ .TlsMinVersion }}"
{{- end }} 
{{- if ne .CipherSuites "" }}
ciphersuites = "{{ .CipherSuites }}"
{{- end }}
{{- if .InsecureSkipVerify }}
verify_certificate = false
verify_hostname = false
{{- end }}
{{- if and .KeyPath .CertPath }}
key_file = {{ .KeyPath }}
crt_file = {{ .CertPath }}
{{- end }}
{{- if .CAFilePath }}
ca_file = {{ .CAFilePath }}
{{- end }}
{{- if .PassPhrase }}
key_pass = "{{ .PassPhrase }}"
{{- end }}
{{ end }}`
}

var NoSecrets = map[string]*corev1.Secret{}

func HasUsernamePassword(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.ClientUsername, constants.ClientPassword)
}

func HasTLSCertAndKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.ClientCertKey, constants.ClientPrivateKey)
}

func HasCABundle(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.TrustedCABundleKey)
}

func HasSharedKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.SharedKey)
}

func HasPassphrase(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.Passphrase)
}

func HasBearerTokenFileKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.BearerTokenFileKey)
}

func HasAwsRoleArnKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSWebIdentityRoleKey)
}

func HasAwsCredentialsKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSCredentialsKey)
}

// GetKey if found return value and ok=true, else ok=false
func GetKey(secret *corev1.Secret, key string) (data []byte, ok bool) {
	if secret == nil {
		return nil, false
	}
	data, ok = secret.Data[key]
	return data, ok
}

// HasKeys true if all keys are present.
func HasKeys(secret *corev1.Secret, keys ...string) bool {
	for _, k := range keys {
		_, ok := GetKey(secret, k)
		if !ok {
			return false
		}
	}
	return true
}

func SecretPath(name string, file string) string {
	return helpers.SecretPath(name, file)
}

// TryKeys try keys in turn return data for fist one present with ok=true.
// If none present return ok=false.
func TryKeys(secret *corev1.Secret, keys ...string) (data []byte, ok bool) {
	for _, k := range keys {
		data, ok := GetKey(secret, k)
		if ok {
			return data, true
		}
	}
	return nil, false
}

func GetFromSecret(secret *corev1.Secret, name string) string {
	if secret != nil {
		return string(secret.Data[name])
	}
	return ""
}

func GenerateTLSConfWithID(id string, o logging.OutputSpec, secret *corev1.Secret, op framework.Options, genTLSConf bool) *TLSConf {
	if !genTLSConf {
		if o.URL == "" {
			genTLSConf = true
		} else if u, _ := url.Parse(o.URL); u != nil {
			genTLSConf = urlhelper.IsTLSScheme(u.Scheme)
		}
	}
	if genTLSConf {
		tlsConf := NewTLSConf(id, o, op)
		if addTLSSettings(o, secret, &tlsConf) {
			return &tlsConf
		}
	}
	return nil
}
