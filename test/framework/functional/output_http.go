package functional

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional/common"
)

const (
	VectorHttpSourceConfTemplate = `
[sources.my_source]
type = "http"
address = "127.0.0.1:8090"
encoding = "ndjson"

{{ if ne .MinTLS "" }}
[sources.my_source.tls]
enabled = true
{{ if ne .MinTLS "" }}
min_tls_version = "{{.MinTLS}}"
{{ end }}
{{ if ne .Ciphers "" }}
ciphersuites = "{{.Ciphers}}"
{{ end }}
key_file = "/tmp/secrets/http/tls.key"
crt_file = "/tmp/secrets/http/tls.crt"
{{ end }}

[sinks.my_sink]
inputs = ["my_source"]
type = "file"
path = "/tmp/app-logs"

[sinks.my_sink.encoding]
codec = "json"
`
	FluentdHttpSourceConf = `
<system>
  log_level debug
</system>
<source>
  @type http
  port 8090
  bind 0.0.0.0
  body_size_limit 32m
  keepalive_timeout 10s
</source>
# send fluentd logs to stdout
<match fluent.**>
  @type stdout
</match>
<match **>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
  <format>
    @type json
  </format>
</match>
`
)

func VectorConfFactory(profile configv1.TLSProfileType) string {
	minTLS := ""
	ciphers := ""
	if profile != "" {
		if spec, found := configv1.TLSProfiles[profile]; found {
			minTLS = string(spec.MinTLSVersion)
			ciphers = strings.Join(spec.Ciphers, ",")
		}

	}
	tmpl, err := template.New("").Parse(VectorHttpSourceConfTemplate)
	if err != nil {
		log.V(0).Error(err, "Unable to parse the vector http conf template")
		os.Exit(1)
	}
	b := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(b, "", struct {
		MinTLS  string
		Ciphers string
	}{
		MinTLS:  minTLS,
		Ciphers: ciphers,
	}); err != nil {
		log.V(0).Error(err, "Unable execute vector http conf template")
		os.Exit(1)
	}
	return b.String()
}

func (f *CollectorFunctionalFramework) AddVectorHttpOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	return f.AddVectorHttpOutputWithConfig(b, output, "", nil)
}

func (f *CollectorFunctionalFramework) AddVectorHttpOutputWithConfig(b *runtime.PodBuilder, output logging.OutputSpec, profile configv1.TLSProfileType, secret *corev1.Secret) error {
	log.V(2).Info("Adding vector http output", "name", output.Name)
	name := strings.ToLower(output.Name)

	toml := VectorConfFactory(profile)
	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"vector.toml": toml,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "vector.toml", toml)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding vector container", "name", name)
	containerBuilder := b.AddContainer(name, utils.GetComponentImage(constants.VectorName)).
		AddVolumeMount(config.Name, "/tmp/config", "", false).
		AddEnvVar("VECTOR_LOG", common.AdaptLogLevel()).
		AddEnvVar("VECTOR_INTERNAL_LOG_RATE_LIMIT", "0").
		WithCmd([]string{"vector", "--config-toml", "/tmp/config/vector.toml"})
	if secret != nil {
		containerBuilder.AddVolumeMount(secret.Name, "/tmp/secrets/http", "", true)
		b.AddSecretVolume(secret.Name, secret.Name)
	}
	containerBuilder.End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

func (f *CollectorFunctionalFramework) AddFluentdHttpOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding fluentd http output", "name", output.Name)
	name := strings.ToLower(output.Name)

	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"fluent.conf": FluentdHttpSourceConf,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "fluent.conf", FluentdHttpSourceConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding fluentd container", "name", name)
	b.AddContainer(name, utils.GetComponentImage(constants.FluentdName)).
		AddVolumeMount(config.Name, "/tmp/config", "", false).
		WithCmd([]string{"fluentd", "-c", "/tmp/config/fluent.conf"}).
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}
