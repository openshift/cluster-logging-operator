package vector

import (
	"regexp"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
	. "github.com/openshift/cluster-logging-operator/test/framework/functional/common"
)

const entrypointScript = `#!/bin/bash
mkdir -p /var/lib/vector

/usr/bin/vector
`

type VectorCollector struct {
	*client.Test
}

func (c *VectorCollector) String() string {
	return constants.VectorName
}

func (c *VectorCollector) DeployConfigMapForConfig(name, config, clfYaml string) error {
	log.V(2).Info("Creating config configmap")
	configmap := runtime.NewConfigMap(c.NS.Name, name, map[string]string{})
	runtime.NewConfigMapBuilder(configmap).
		Add("vector.toml", config).
		Add("clfyaml", clfYaml).
		Add("run.sh", entrypointScript)
	if err := c.Create(configmap); err != nil {
		return err
	}
	return nil
}

func (c *VectorCollector) BuildCollectorContainer(b *runtime.ContainerBuilder, nodeName string) *runtime.ContainerBuilder {
	return b.AddEnvVar("VECTOR_LOG", AdaptLogLevel()).
		AddEnvVarFromFieldRef("POD_IP", "status.podIP").
		AddEnvVar("NODE_NAME", nodeName).
		AddEnvVarFromFieldRef("VECTOR_SELF_NODE_NAME", "spec.nodeName").
		AddVolumeMount("config", "/etc/vector", "", true).
		AddVolumeMount("entrypoint", "/opt/app-root/src/run.sh", "run.sh", true).
		WithCmd([]string{"/opt/app-root/src/run.sh"})
}

func (c *VectorCollector) IsStarted(logs string) bool {
	return strings.Contains(logs, "Vector has started.")
}

func (c *VectorCollector) Image() string {
	return utils.GetComponentImage(constants.VectorName)
}

const fakeJournal = `
[sources.fake_raw_journal_logs]
type = "file"
include = ["/var/log/fakejournal/0.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

[transforms.parse_journal_message]
type= "remap"
inputs = ["fake_raw_journal_logs"]
source = '''
  content = parse_json!(.message)
  .parsed = content
'''

[transforms.raw_journal_logs]
type = "lua"
inputs = ["parse_journal_message"]
version = "2"
hooks.process = "process"
source = '''
function process(event, emit)
  if event.log.parsed ~= nil then
    for k,v in pairs(event.log.parsed) do
      event.log[k] = v
    end
    event.log.parsed = nil
  end
  event.log.host = event.log.hostname
  event.log.hostname = nil
  if event.log.MESSAGE ~= nil then
    event.log.message = event.log.MESSAGE
    event.log.MESSAGE = nil
  end
  emit(event)
end
'''

`

func (c *VectorCollector) ModifyConfig(conf string) string {
	//remove journal for now since we can not mimic it
	re := regexp.MustCompile(`(?msU)\[sources\.raw_journal_logs\].*^\n`)
	return string(re.ReplaceAll([]byte(conf), []byte(fakeJournal)))
}
