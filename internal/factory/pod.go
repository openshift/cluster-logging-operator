package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/utils"
	core "k8s.io/api/core/v1"
)

// NewContainer stubs an instance of a Container for cluster logging.  Note the
// imageName is an alias for a well-known logging image(e.g. fluentd)
func NewContainer(containerName string, imageName string, pullPolicy core.PullPolicy, resources core.ResourceRequirements) core.Container {
	return core.Container{
		Name:            containerName,
		Image:           utils.GetComponentImage(imageName),
		ImagePullPolicy: pullPolicy,
		Resources:       resources,
	}
}
