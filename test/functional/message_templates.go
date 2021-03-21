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

func NewApplicationLogTemplate() types.ApplicationLog{
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

func NewLogTemplate() types.AllLog{
	return types.AllLog{
		Timestamp:    time.Time{},
		Message:       "*",
		ViaqIndexName: "",
		Level:         "unknown",
		ViaqMsgID:     "*",
		PipelineMetadata: templateForAnyPipelineMetadata,
		Docker: types.Docker{
			ContainerID: "*",
		},
		Kubernetes: templateForAnyKubernetes,
	}
}
