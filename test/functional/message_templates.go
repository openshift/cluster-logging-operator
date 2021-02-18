package functional

import (
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"time"
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
		ViaqIndexName:    "app-write",
		Level:            "unknown",
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
