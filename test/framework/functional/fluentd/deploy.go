package fluentd

import (
	"fmt"
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/components/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"os"
	"strconv"
	"strings"
)

type FluentdCollector struct {
	*client.Test
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
mkdir -p /var/log/pods/%s_functional_123456789-0/loader-0
`
	replace = fmt.Sprintf(replace, c.NS.Name)
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
	return b.AddEnvVar("LOG_LEVEL", adaptLogLevel()).
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

func adaptLogLevel() string {
	logLevel := "debug"
	if level, found := os.LookupEnv("LOG_LEVEL"); found {
		if i, err := strconv.Atoi(level); err == nil {
			switch i {
			case 0:
				logLevel = "error"
			case 1:
				logLevel = "info"
			case 2:
				logLevel = "debug"
			case 3 - 8:
				logLevel = "trace"
			default:
			}
		} else {
			log.V(1).Error(err, "Unable to set LOG_LEVEL from environment")
		}
	}
	return logLevel
}
