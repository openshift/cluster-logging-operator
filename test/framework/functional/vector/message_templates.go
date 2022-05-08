package vector

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
		ContainerName:   "*",
		PodName:         "*",
		PodNodeName:     "*",
		NamespaceName:   "*",
		ContainerID:     "*",
		ContainerImage:  "*",
		PodID:           "*",
		PodIP:           "**optional**",
		FlatLabels:      []string{"*"},
		NamespaceLabels: map[string]string{"*": "*"},
		Annotations:     map[string]string{"*": "*"},
	}
	templateForInfraKubernetes = types.Kubernetes{
		ContainerName:     "*",
		PodName:           "*",
		PodNodeName:       "*",
		NamespaceName:     "*",
		ContainerID:       "*",
		OrphanedNamespace: "**optional**",
		ContainerImage:    "*",
		PodID:             "**optional**",
		PodIP:             "**optional**",
		FlatLabels:        []string{"*"},
		NamespaceLabels:   map[string]string{"*": "*"},
		Annotations:       map[string]string{"*": "*"},
	}
)

func NewApplicationLogTemplate() types.ApplicationLog {
	return types.ApplicationLog{
		Timestamp: time.Time{},
		Message:   "*",
		LogType:   "application",
		Level:     "*",
		Hostname:  "",
		ViaqMsgID: "",
		Openshift: types.OpenshiftMeta{
			Labels:   map[string]string{"*": "*"},
			Sequence: types.NewOptionalInt(""),
		},
		PipelineMetadata: types.PipelineMetadata{},
		Docker:           types.Docker{},
		Kubernetes:       templateForAnyKubernetes,
		WriteIndex:       "app-write",
	}
}

// NewContainerInfrastructureLogTemplate creates a generally expected template for infrastructure container logs
func NewContainerInfrastructureLogTemplate() types.ApplicationLog {
	return types.ApplicationLog{
		Timestamp: time.Time{},
		Message:   "*",
		LogType:   "infrastructure",
		Level:     "*",
		Hostname:  "*",
		ViaqMsgID: "*",
		Openshift: types.OpenshiftMeta{
			Labels:   map[string]string{"*": "*"},
			Sequence: types.NewOptionalInt(""),
		},
		PipelineMetadata: types.PipelineMetadata{},
		Docker:           types.Docker{},
		Kubernetes:       templateForInfraKubernetes,
		WriteIndex:       "infra-write",
	}
}
