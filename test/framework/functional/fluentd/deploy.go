package fluentd

import (
	. "github.com/openshift/cluster-logging-operator/test/framework/functional/common"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
)

type FluentdCollector struct {
	*client.Test
}

func (c *FluentdCollector) String() string {
	return constants.FluentdName
}

func (c *FluentdCollector) DeployConfigMapForConfig(name, config, clfYaml string) error {
	log.V(2).Info("Creating config configmap")
	configmap := runtime.NewConfigMap(c.NS.Name, name, map[string]string{})

	//create dirs that dont exist in testing
	replace := `#!/bin/bash
mkdir -p /var/log/{kube-apiserver,oauth-apiserver,audit,ovn}
for d in kube-apiserver oauth-apiserver audit; do
  touch /var/log/$d/{audit.log,acl-audit-log.log}
done
`
	runScript := strings.Replace(fluentd.RunScript, "#!/bin/bash\n", replace, 1)
	runtime.NewConfigMapBuilder(configmap).
		Add("fluent.conf", config).
		Add("run.sh", runScript).
		Add("clfyaml", clfYaml)
	if err := c.Create(configmap); err != nil {
		return err
	}
	return nil
}

func (c *FluentdCollector) BuildCollectorContainer(b *runtime.ContainerBuilder, nodeName string) *runtime.ContainerBuilder {
	return b.AddEnvVar("LOG_LEVEL", AdaptLogLevel()).
		AddEnvVarFromFieldRef("POD_IP", "status.podIP").
		AddEnvVarFromFieldRef("K8S_NODE_NAME", "spec.nodeName").
		AddEnvVar("NODE_NAME", nodeName).
		AddVolumeMount("config", "/etc/fluent/configs.d/user", "", true).
		AddVolumeMount("entrypoint", "/opt/app-root/src/run.sh", "run.sh", true).
		AddVolumeMount("certs", "/etc/collector/metrics", "", true)
}

func (c *FluentdCollector) IsStarted(logs string) bool {
	phase, err := oc.Get().Pod().WithNamespace(c.NS.Name).Name("functional").OutputJsonpath("{.status.phase}").Run()
	if err != nil {
		log.V(1).Error(err, "Unable to determine if the collector started")
		return false
	}
	return "Running" == strings.TrimSpace(phase)
}

func (c *FluentdCollector) Image() string {
	return utils.GetComponentImage(constants.FluentdName)
}
