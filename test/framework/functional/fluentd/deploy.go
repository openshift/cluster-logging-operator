package fluentd

import (
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/components/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"strings"
)

type FluentdCollector struct {
	*client.Test
}


func(c *FluentdCollector) DeployConfigMapForConfig(name, config, clfYaml string) error {
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
	return b.AddEnvVar("LOG_LEVEL", "debug").
		AddEnvVarFromFieldRef("POD_IP", "status.podIP").
		AddEnvVar("NODE_NAME", nodeName).
		AddVolumeMount("config", "/etc/fluent/configs.d/user", "", true).
		AddVolumeMount("entrypoint", "/opt/app-root/src/run.sh", "run.sh", true).
		AddVolumeMount("certs", "/etc/fluent/metrics", "", true)
}

func (c *FluentdCollector) IsStarted(logs string) bool {
	// if fluentd started successfully return success
	return strings.Contains(logs, "flush_thread actually running")
}
