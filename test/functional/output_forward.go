package functional

import (
	"fmt"
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
	log_level trace
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
	`
)

func (f *FluentdFunctionalFramework) addForwardOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding forward output", "name", output.Name)
	name := strings.ToLower(output.Name)
	configName := fmt.Sprintf("%s-config", name)
	log.V(2).Info("Creating configmap", "name", configName)
	config := runtime.NewConfigMap(b.Pod.Namespace, configName, map[string]string{
		"fluent.conf": unsecureFluentConf,
	})
	if err := f.test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding container", "name", name)
	b.AddContainer(name, utils.GetComponentImage(constants.FluentdName)).
		AddVolumeMount(config.Name, "/tmp/config", "", true).
		WithCmd("fluentd -c /tmp/config/fluent.conf").
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}
