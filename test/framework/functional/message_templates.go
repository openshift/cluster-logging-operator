package functional

import (
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

var (
	TemplateForAnyPipelineMetadata = types.PipelineMetadata{
		Collector: types.Collector{
			Ipaddr4:    "**optional**",
			Inputname:  "**optional**",
			Name:       "**optional**",
			Version:    "**optional**",
			ReceivedAt: time.Time{},
		},
	}
	templateForAnyKubernetes = types.Kubernetes{
		ContainerName:    "*",
		ContainerID:      "**optional**",
		PodName:          "*",
		NamespaceName:    "*",
		NamespaceID:      "**optional**",
		ContainerImage:   "*",
		ContainerImageID: "**optional**",
		PodID:            "*",
		PodIP:            "**optional**",
		Host:             "**optional**",
		MasterURL:        "**optional**",
		FlatLabels:       []string{"*"},
		NamespaceLabels:  map[string]string{"*": "*"},
		Annotations:      map[string]string{"*": "*"},
	}
	templateForInfraKubernetes = types.Kubernetes{
		ContainerID:       "**optional**",
		ContainerName:     "*",
		PodName:           "*",
		NamespaceName:     "*",
		NamespaceID:       "**optional**",
		OrphanedNamespace: "**optional**",
		ContainerImage:    "**optional**",
		ContainerImageID:  "**optional**",
		PodID:             "**optional**",
		PodIP:             "**optional**",
		Host:              "**optional**",
		MasterURL:         "**optional**",
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
		Hostname:  "*",
		ViaqMsgID: "**optional**",
		Openshift: types.OpenshiftMeta{
			Labels:    map[string]string{"*": "*"},
			Sequence:  types.NewOptionalInt(""),
			ClusterID: "*",
		},
		PipelineMetadata: TemplateForAnyPipelineMetadata,
		Docker: types.Docker{
			ContainerID: "**optional**",
		},
		Kubernetes: templateForAnyKubernetes,
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
		ViaqMsgID: "**optional**",
		Openshift: types.OpenshiftMeta{
			Labels:    map[string]string{"*": "*"},
			Sequence:  types.NewOptionalInt(""),
			ClusterID: "*",
		},
		PipelineMetadata: TemplateForAnyPipelineMetadata,
		Docker: types.Docker{
			ContainerID: "**optional**",
		},
		Kubernetes: templateForInfraKubernetes,
	}
}
