package loki

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
)

const (
	lokiLabelKubernetesHost = "kubernetes.host"
)

var (
	defaultLabelKeys = []string{
		"log_type",

		//container labels
		"kubernetes.namespace_name",
		"kubernetes.pod_name",
		"kubernetes.container_name",
	}
	requiredLabelKeys = []string{
		lokiLabelKubernetesHost,
	}
)

type Loki struct {
	StoreID        string
	Tenant         Element
	TLSMinVersion  string
	CipherSuites   string
	URLBase        string
	LokiLabel      []string
	SecurityConfig []Element
	BufferConfig   []Element
}

func (l Loki) Name() string {
	return "lokiTemplate"
}

func (l Loki) Template() string {
	return `{{define "` + l.Name() + `" -}}
@type loki
@id {{.StoreID}}
line_format json
url {{.URLBase}}
{{if (ne .TLSMinVersion "") -}}
min_version {{.TLSMinVersion}}
{{end -}}
{{if (ne .CipherSuites "") -}}
ciphers {{.CipherSuites}}
{{end -}}
{{kv .Tenant -}}
{{compose .SecurityConfig}}
<label>
{{range $index, $label := .LokiLabel -}}
{{$label | indent 2}}
{{end -}}
</label>
{{compose .BufferConfig}}
{{end}}`
}

func (l *Loki) setTLSProfileFromOptions(op Options) {
	if version, found := op[MinTLSVersion]; found {
		opVersion := version.(string)
		if minVersion := helpers.TLSMinVersion(opVersion); minVersion != "" {
			l.TLSMinVersion = minVersion
		}
	}
	if ciphers, found := op[Ciphers]; found {
		l.CipherSuites = ciphers.(string)
	}
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []Element{
				normalize.DedotLabels(),
				LokiLabelFilter(o.Loki),
				Output(bufspec, secret, o, op),
			},
		},
	}
}

func Output(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	// url is parsable, checked at input sanitization
	u, _ := urlhelper.Parse(o.URL)
	urlBase := fmt.Sprintf("%v://%v%v", u.Scheme, u.Host, u.Path)
	storeID := helpers.StoreID("", o.Name, "")
	loki := Loki{
		StoreID:        strings.ToLower(helpers.Replacer.Replace(o.Name)),
		URLBase:        urlBase,
		Tenant:         Tenant(o.Loki),
		LokiLabel:      LokiLabel(o.Loki),
		SecurityConfig: SecurityConfig(o, secret),
		BufferConfig:   output.Buffer(output.NOKEYS, bufspec, storeID, &o),
	}
	loki.setTLSProfileFromOptions(op)
	return Match{
		MatchTags:    "**",
		MatchElement: loki,
	}
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret != nil {
		if security.HasUsernamePassword(secret) {
			up := UserNamePass{
				UsernamePath: security.SecretPath(o.Secret.Name, constants.ClientUsername),
				PasswordPath: security.SecretPath(o.Secret.Name, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if security.HasTLSCertAndKey(secret) {
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if security.HasCABundle(secret) {
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
		if security.HasBearerTokenFileKey(secret) {
			bt := BearerTokenFile{
				BearerTokenFilePath: security.SecretPath(o.Secret.Name, constants.BearerTokenFileKey),
			}
			conf = append(conf, bt)
		}
	} else if secret != nil {
		// Use secret of logcollector ServiceAccount as fallback
		conf = append(conf, CAFile{
			CAFilePath: "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt",
		}, BearerTokenFile{
			BearerTokenFilePath: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		})
	}
	return conf
}

func lokiLabelKeys(l *logging.Loki) []string {
	var keys sets.String
	if l != nil && len(l.LabelKeys) != 0 {
		keys = sets.NewString(l.LabelKeys...)
	} else {
		keys = sets.NewString(defaultLabelKeys...)
	}
	// Ensure required tags for serialization
	keys.Insert(requiredLabelKeys...)
	return keys.List()
}

// LokiLabelFilter generates record_modifier filter lines to copy Loki label fields.
// The Loki output plugin will remove these fields after creating Loki labels.
func LokiLabelFilter(l *logging.Loki) Element {
	rs := []Record{}
	for _, k := range lokiLabelKeys(l) {
		tempName := strings.Replace(k, ".", "_", -1)
		tempName = strings.Replace(tempName, "/", "_", -1)
		recordKeys := strings.Replace(k, ".", `","`, -1)
		recordKeys = strings.Replace(recordKeys, "/", "_", -1)
		var r Record
		switch k {
		case lokiLabelKubernetesHost:
			r = Record{
				Key:        fmt.Sprintf("_%v", tempName),
				Expression: "\"#{ENV['NODE_NAME']}\"",
			}
		default:
			r = Record{
				Key:        fmt.Sprintf("_%v", tempName),
				Expression: fmt.Sprintf("${record.dig(\"%v\")}", recordKeys),
			}
		}
		rs = append(rs, r)
	}
	if len(rs) == 0 {
		return Nil
	}
	return Filter{
		MatchTags: "**",
		Element: RecordModifier{
			Records: rs,
		},
	}
}

// LokiLabel generates the <label> entries for Loki output config.
// This consumes the fields generated by LokiLabelFilter.
func LokiLabel(l *logging.Loki) []string {
	labels := []string{}
	for _, k := range lokiLabelKeys(l) {
		tempName := strings.Replace(k, ".", "_", -1)
		tempName = strings.Replace(tempName, "/", "_", -1)
		labels = append(labels, fmt.Sprintf("%v _%v", tempName, tempName))
	}
	return labels
}

// LokiTenantKeys returns the components of the loki tenant key.
func LokiTenantKeys(l *logging.Loki) []string {
	if l != nil && l.TenantKey != "" {
		return strings.Split(l.TenantKey, ".")
	}
	return nil
}

func Tenant(l *logging.Loki) Element {
	if l == nil || l.TenantKey == "" {
		return Nil
	}
	keys := []string{}
	for _, k := range LokiTenantKeys(l) {
		keys = append(keys, fmt.Sprintf("%q", k))
	}
	return KV("tenant", fmt.Sprintf("${record.dig(%v)}", strings.Join(keys, ",")))
}
