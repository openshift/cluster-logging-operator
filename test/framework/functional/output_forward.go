package functional

import (
	fluentdhelpers "github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

const (
	defaultFluentdforwardPort = "24224"
	unsecureFluentConf        = `
<system>
  log_level debug
</system>
<source>
  @type forward
  port 24224
</source>
<filter **>
	@type stdout
	include_time_key true 
</filter>

<match kubernetes.** var.log.**>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
  <format>
    @type json
  </format>
  <buffer>
    @type file
    append true
  </buffer>
</match>

<match journal.system>
  @type file
  append true
  path /tmp/infra.logs
  symlink_path /tmp/infra-logs
  <format>
    @type json
  </format>
  <buffer>
    @type file
    append true
  </buffer>
</match>

<filter linux-audit.log**>
  @type parser
  key_name @timestamp
  reserve_data true
  <parse>
	@type regexp
	expression (?<time>[^\]]*)
    time_type string
	time_key time
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</filter>

<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
	@type file
	path /tmp/audit.logs
	append true
	symlink_path /tmp/audit-logs
	<format>
		@type json
	</format>
</match>
	
<match **>
  @type stdout
</match>`
)

func (f *CollectorFunctionalFramework) addForwardOutputWithConf(b *runtime.PodBuilder, output logging.OutputSpec, conf string) error {
	log.V(2).Info("Adding forward output", "name", output.Name)
	name := strings.ToLower(output.Name)
	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"fluent.conf": conf,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "fluent.conf", unsecureFluentConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding container", "name", name)
	b.AddContainer(name, fluentdhelpers.Image).
		AddVolumeMount(config.Name, "/tmp/config", "", false).
		WithCmd([]string{"fluentd", "-c", "/tmp/config/fluent.conf"}).
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

func (f *CollectorFunctionalFramework) AddForwardOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	outURL, err := url.Parse(output.URL)
	if err != nil {
		return err
	}
	config := strings.Replace(unsecureFluentConf, defaultFluentdforwardPort, outURL.Port(), 1)
	return f.addForwardOutputWithConf(b, output, config)
}
