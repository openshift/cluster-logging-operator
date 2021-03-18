package functional

import (
	"strings"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

const (
	unsecureFluentConf = `
<system>
  log_level debug
</system>
<source>
  @type forward
</source>
<filter **>
	@type stdout
	include_time_key true 
</filter>

<match kubernetes.**>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
  <format>
    @type json
  </format>
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

<match linux-audit.log** k8s-audit.log** openshift-audit.log**>
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

	unsecureFluentConfBenchmark = `
<system>
  log_level debug
</system>
<source>
  @type forward
</source>
<filter **>
	@type stdout
	include_time_key true 
</filter>

<filter kubernetes.**>
  @type record_transformer
  enable_ruby
  <record>
    epoc_out ${Time.now.to_f}
    epoc_in ${Time.parse(record['@timestamp']).to_f}
    level info
  </record>
</filter>

<match kubernetes.**>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
  <format>
    @type json
  </format>
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

<match linux-audit.log** k8s-audit.log** openshift-audit.log**>
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

func (f *FluentdFunctionalFramework) addForwardOutputWithConf(b *runtime.PodBuilder, output logging.OutputSpec, conf string) error {
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
	b.AddContainer(name, utils.GetComponentImage(constants.FluentdName)).
		AddVolumeMount(config.Name, "/tmp/config", "", false).
		WithCmd("fluentd -c /tmp/config/fluent.conf").
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

func (f *FluentdFunctionalFramework) AddForwardOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	return f.addForwardOutputWithConf(b, output, unsecureFluentConf)
}

func (f *FluentdFunctionalFramework) AddBenchmarkForwardOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	return f.addForwardOutputWithConf(b, output, unsecureFluentConfBenchmark)
}
