package functional

import (
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"time"
)

var (
	templateForAnyPipelineMetadata = types.PipelineMetadata{
		Collector: types.Collector{
			Ipaddr4:    "*",
			Inputname:  "*",
			Name:       "*",
			Version:    "*",
			ReceivedAt: time.Time{},
		},
	}
	templateForAnyKubernetes = types.Kubernetes{
		ContainerName:     "*",
		PodName:           "*",
		NamespaceName:     "*",
		NamespaceID:       "*",
		OrphanedNamespace: "*",
	}
)

func NewApplicationLogTempate() types.ApplicationLog {
	return types.ApplicationLog{
		Timestamp:        time.Time{},
		Message:          "*",
		ViaqIndexName:    "app-write",
		Level:            "unknown",
		ViaqMsgID:        "*",
		PipelineMetadata: templateForAnyPipelineMetadata,
		Docker: types.Docker{
			ContainerID: "*",
		},
		Kubernetes: templateForAnyKubernetes,
	}
}
