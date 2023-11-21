package otel

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"strings"

	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func Transform(id string, inputs []string) Element {
	return Remap{
		Desc:        "Normalize log records to OTEL schema",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL: strings.TrimSpace(`
# Tech preview, OTEL for application logs only
if .log_type == "application" {
	# Convert @timestamp to nano and delete @timestamp
	.timeUnixNano = to_unix_timestamp!(del(.@timestamp), unit:"nanoseconds")

	.severityText = del(.level)

	# Convert syslog severity keyword to number, default to 9 (unknown)
	.severityNumber = to_syslog_severity(.severityText) ?? 9

	# resources
	.resources.logs.file.path = del(.file)
	.resources.host.name= del(.hostname)
	.resources.container.name = del(.kubernetes.container_name)
	.resources.container.id = del(.kubernetes.container_id)
  
	# split image name and tag into separate fields
	container_image_slice = split!(.kubernetes.container_image, ":", limit: 2)
	if null != container_image_slice[0] { .resources.container.image.name = container_image_slice[0] }
	if null != container_image_slice[1] { .resources.container.image.tag = container_image_slice[1] }
	del(.kubernetes.container_image)
	
	# kuberenetes
	.resources.k8s.pod.name = del(.kubernetes.pod_name)
	.resources.k8s.pod.uid = del(.kubernetes.pod_id)
	.resources.k8s.pod.ip = del(.kubernetes.pod_ip)
	.resources.k8s.pod.owner = .kubernetes.pod_owner
	.resources.k8s.pod.annotations = del(.kubernetes.annotations)
	.resources.k8s.pod.labels = del(.kubernetes.labels)
	.resources.k8s.namespace.id = del(.kubernetes.namespace_id)
	.resources.k8s.namespace.name = .kubernetes.namespace_labels."kubernetes.io/metadata.name"
	.resources.k8s.namespace.labels = del(.kubernetes.namespace_labels)
	.resources.attributes.log_type = del(.log_type)
}
`),
	}
}
