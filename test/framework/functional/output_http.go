package functional

import (
	"bytes"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"os"
	"strings"
	"text/template"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	fluentdhelpers "github.com/openshift/cluster-logging-operator/test/helpers/fluentd"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional/common"
)

const (
	VectorHttpSourceBenchmarkConfTemplate = "" +
		`[sources.my_source]
type = "http_server"
address = "127.0.0.1:8090"
decoding.codec = "json"
framing.method = "newline_delimited"

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

[transforms.app_logs]
type = "remap"
inputs = ["my_source"]
source = '''
 .epoc_out = to_float(now())
 mytime, err = parse_timestamp(."@timestamp", format: "%Y-%m-%dT%H:%M:%S%.fZ")
 if err != null {
   mytime, err = parse_timestamp(."@timestamp", format: "%Y-%m-%dT%H:%M:%S%.f%z")
   if err != null {
     log("Unable to parse @timestamp: " + err, level: "error")
   }
 }
 .epoc_in = to_float(mytime)

'''

[sinks.my_sink]
inputs = ["app_logs"]
type = "file"
path = "{{.Path}}"
		
[sinks.my_sink.encoding]
codec = "json"
`
	VectorHttpSourceConfTemplate = "" +
		`[sources.my_source]
type = "http_server"
address = "127.0.0.1:8090"
decoding.codec = "json"
framing.method = "newline_delimited"
{{ .Username }}
{{ .Password }}

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

[transforms.app_logs]
type = "remap"
inputs = ["my_source"]
source = '''
  del(.source_type)
'''

[sinks.my_sink]
inputs = ["app_logs"]
type = "file"
path = "{{.Path}}"

[sinks.my_sink.encoding]
codec = "json"
except_fields = ["path"]
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

func VectorConfFactory(profile configv1.TLSProfileType, options ...Option) string {
	confTemplate := VectorHttpSourceConfTemplate
	if found, o := OptionsInclude("template", options); found {
		confTemplate = o.Value
	}
	path := "/tmp/app-logs"
	if found, o := OptionsInclude("path", options); found {
		path = o.Value
	}
	minTLS := ""
	ciphers := ""
	if profile != "" {
		if spec, found := configv1.TLSProfiles[profile]; found {
			minTLS = string(spec.MinTLSVersion)
			ciphers = strings.Join(spec.Ciphers, ",")
		}

	}
	tmpl, err := template.New("").Parse(confTemplate)
	if err != nil {
		log.V(0).Error(err, "Unable to parse the vector http conf template")
		os.Exit(1)
	}
	b := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(b, "", struct {
		MinTLS   string
		Ciphers  string
		Path     string
		Username helpers.OptionalPair
		Password helpers.OptionalPair
	}{
		MinTLS:   minTLS,
		Ciphers:  ciphers,
		Path:     path,
		Username: helpers.NewOptionalPair("auth.username", OptionsValue(options, "username", nil)),
		Password: helpers.NewOptionalPair("auth.password", OptionsValue(options, "password", nil)),
	}); err != nil {
		log.V(0).Error(err, "Unable execute vector http conf template")
		os.Exit(1)
	}
	return b.String()
}

func (f *CollectorFunctionalFramework) AddVectorHttpOutput(b *runtime.PodBuilder, output obs.OutputSpec, options ...Option) error {
	return f.AddVectorHttpOutputWithConfig(b, output, "", nil, options...)
}

func (f *CollectorFunctionalFramework) AddVectorHttpOutputWithConfig(b *runtime.PodBuilder, output obs.OutputSpec, profile configv1.TLSProfileType, secret *corev1.Secret, options ...Option) error {
	log.V(2).Info("Adding vector http output", "name", output.Name)
	name := strings.ToLower(output.Name)

	toml := VectorConfFactory(profile, options...)
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

func (f *CollectorFunctionalFramework) AddFluentdHttpOutput(b *runtime.PodBuilder, output obs.OutputSpec) error {
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
	b.AddContainer(name, fluentdhelpers.Image).
		AddVolumeMount(config.Name, "/tmp/config", "", false).
		WithCmd([]string{"fluentd", "-c", "/tmp/config/fluent.conf"}).
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

func (f *CollectorFunctionalFramework) AddBenchmarkForwardOutput(b *runtime.PodBuilder, output obs.OutputSpec, image string) error {
	if err := f.AddVectorHttpOutputWithConfig(b, output, "", nil, Option{"path", "/tmp/{{kubernetes.container_name}}.log"}, Option{"template", VectorHttpSourceBenchmarkConfTemplate}); err != nil {
		return err
	}
	b.GetContainer(string(obs.OutputTypeHTTP)).WithImage(image).Update()
	return nil
}
