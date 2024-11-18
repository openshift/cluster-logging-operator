package vector

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"regexp"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/collector/vector"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional/common"
)

type VectorCollector struct {
	*client.Test
}

func (c *VectorCollector) String() string {
	return constants.VectorName
}

func (c *VectorCollector) DeployConfigMapForConfig(name, config, clfName, clfYaml string) error {
	log.V(2).Info("Creating config configmap")
	configmap := runtime.NewConfigMap(c.NS.Name, name, map[string]string{})
	runtime.NewConfigMapBuilder(configmap).
		Add(vector.ConfigFile, config).
		Add("clfyaml", clfYaml).
		Add("run.sh", fmt.Sprintf(vector.RunVectorScript, vector.GetDataPath(c.NS.Name, clfName)))
	if err := c.Create(configmap); err != nil {
		return err
	}
	return nil
}

func (c *VectorCollector) BuildCollectorContainer(b *runtime.ContainerBuilder, nodeName string) *runtime.ContainerBuilder {
	return b.AddEnvVar("VECTOR_LOG", common.AdaptLogLevel()).
		AddEnvVarFromFieldRef("POD_IP", "status.podIP").
		AddEnvVar("NODE_NAME", nodeName).
		AddEnvVar("VECTOR_INTERNAL_LOG_RATE_LIMIT", "0").
		AddEnvVarFromFieldRef("VECTOR_SELF_NODE_NAME", "spec.nodeName").
		AddVolumeMount("config", "/etc/vector", "", true).
		AddVolumeMount("entrypoint", "/opt/app-root/src/run.sh", "run.sh", true).
		AddVolumeMount("certs", "/etc/collector/metrics", "", true).
		WithCmd([]string{"/opt/app-root/src/run.sh"})
}

func (c *VectorCollector) IsStarted(logs string) bool {
	return strings.Contains(logs, "Vector has started.")
}

func (c *VectorCollector) Image() string {
	return utils.GetComponentImage(constants.VectorName)
}

const fakeJournal = `
[sources.fake_input_infrastructure_journal]
type = "file"
include = ["/var/log/fakejournal/0.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

[transforms.parse_journal_message]
type= "remap"
inputs = ["fake_input_infrastructure_journal"]
source = '''
  content = parse_json!(.message)
  .parsed = content
'''

[transforms.input_infrastructure_journal]
type = "remap"
inputs = ["parse_journal_message"]
source = '''
if exists(.parsed) {
  for_each(object!(.parsed)) -> |key,value| {
    . = set!(.,[key],value)
  }	
  del(.parsed)
}
.host = del(.hostname)
if exists(.MESSAGE) {
  .message = del(.MESSAGE)
}
'''

`

func (c *VectorCollector) ModifyConfig(conf string) string {
	//remove journal for now since we can not mimic it
	re := regexp.MustCompile(`(?msU)\[sources\.input_infrastructure_journal\].*^\n`)
	return string(re.ReplaceAll([]byte(conf), []byte(fakeJournal)))
}
