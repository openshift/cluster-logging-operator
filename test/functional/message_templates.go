package functional

import (
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

var (
	TemplateForAnyPipelineMetadata = types.PipelineMetadata{
		Collector: types.Collector{
			Ipaddr4:    "*",
			Inputname:  "*",
			Name:       "*",
			Version:    "*",
			ReceivedAt: time.Time{},
		},
	}
	templateForAnyKubernetes = types.Kubernetes{
		ContainerName:    "*",
		PodName:          "*",
		NamespaceName:    "*",
		NamespaceID:      "*",
		ContainerImage:   "*",
		ContainerImageID: "*",
		PodID:            "*",
		PodIP:            "**optional**",
		Host:             "*",
		MasterURL:        "*",
		FlatLabels:       []string{"*"},
		NamespaceLabels:  map[string]string{"*": "*"},
	}
)

func NewApplicationLogTemplate() types.ApplicationLog {
	return types.ApplicationLog{
		Timestamp:        time.Time{},
		Message:          "*",
		LogType:          "application",
		ViaqIndexName:    "app-write",
		Level:            "*",
		Hostname:         "*",
		ViaqMsgID:        "*",
		OpenshiftLabels:  types.OpenshiftMeta{Labels: map[string]string{"*": "*"}},
		PipelineMetadata: TemplateForAnyPipelineMetadata,
		Docker: types.Docker{
			ContainerID: "*",
		},
		Kubernetes: templateForAnyKubernetes,
	}
}
