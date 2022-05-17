package vector

import (
	"strings"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
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
	log.NewLogger("vector-deploy-testing").V(2).Info("Creating config configmap")
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
	return b.AddEnvVar("LOG", "debug").
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
