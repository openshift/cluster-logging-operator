package otel

import (
	"regexp"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/collector/otel"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional/common"
)

type OTELCollector struct {
	*client.Test
}

func (c *OTELCollector) String() string {
	return constants.OTELCollectorName
}

func (c *OTELCollector) DeployConfigMapForConfig(name, config, clfName, clfYaml string) error {
	log.V(2).Info("Creating config configmap for OTEL collector")
	configmap := runtime.NewConfigMap(c.NS.Name, name, map[string]string{})
	runtime.NewConfigMapBuilder(configmap).
		Add(otel.ConfigFile, config).
		Add("clfyaml", clfYaml)
	if err := c.Create(configmap); err != nil {
		return err
	}
	return nil
}

func (c *OTELCollector) BuildCollectorContainer(b *runtime.ContainerBuilder, nodeName string) *runtime.ContainerBuilder {
	return b.AddEnvVar("OTEL_LOG_LEVEL", common.AdaptLogLevel()).
		AddEnvVarFromFieldRef("POD_IP", "status.podIP").
		AddEnvVar("NODE_NAME", nodeName).
		AddEnvVarFromFieldRef("OTEL_RESOURCE_ATTRIBUTES_NODE_NAME", "spec.nodeName").
		AddVolumeMount("config", "/etc/otelcol", "", true).
		AddVolumeMount("certs", "/etc/collector/metrics", "", true).
		WithCmd([]string{"/otelcol-contrib", "--config=/etc/otelcol/" + otel.ConfigFile})
}

func (c *OTELCollector) IsStarted(logs string) bool {
	return strings.Contains(logs, "Everything is ready.")
}

func (c *OTELCollector) Image() string {
	return utils.GetComponentImage(constants.OTELCollectorName)
}

const fakeJournal = `
  filelog/fake_journal:
    include:
      - /var/log/fakejournal/0.log
    start_at: beginning
    include_file_path: true
    include_file_name: false
    operators:
      - type: json_parser
        id: json-parser
`

func (c *OTELCollector) ModifyConfig(conf string) string {
	re := regexp.MustCompile(`(?msU)  journalctl.*?^\n`)
	return string(re.ReplaceAll([]byte(conf), []byte(fakeJournal)))
}
