package vector

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
)

const (
	ParseJson = "json"
)

func Pipelines(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	el := []generator.Element{}
	userDefined := spec.InputMap()
	for i, p := range spec.Pipelines {
		p.Name = helpers.FormatComponentID(p.Name) + "_user_defined"
		spec.Pipelines[i].Name = p.Name
		inputs := []string{}
		for _, inputName := range p.InputRefs {
			input, isUserDefined := userDefined[inputName]
			switch {
			case isUserDefined && input.HasPolicy():
				if input.GetMaxRecordsPerSecond() > 0 {
					// if threshold is zero, then don't add to pipeline
					inputs = append(inputs, fmt.Sprintf(UserDefinedSourceThrottle, input.Name))
				}
			case isUserDefined:
				inputs = append(inputs, fmt.Sprintf(UserDefinedInput, inputName))
			default:
				inputs = append(inputs, inputName)
			}
		}

		if p.DetectMultilineErrors {
			d := DetectExceptions{
				ComponentID: fmt.Sprintf("detect_exceptions_%s", p.Name),
				Inputs:      helpers.MakeInputs(inputs...),
			}
			el = append(el, d)
			inputs = []string{d.ComponentID}
		}
		vrls := []string{}
		if p.Labels != nil && len(p.Labels) != 0 {
			s, _ := json.Marshal(p.Labels)
			vrls = append(vrls, fmt.Sprintf(".openshift.labels = %s", s))
		}
		if p.Parse == ParseJson {
			parse := `
if .log_type == "application" {
  parsed, err = parse_json(.message)
  if err == null {
    .structured = parsed
    del(.message)
  }
}
`
			vrls = append(vrls, parse)
		}
		if p.Schema == constants.OtelSchema && genhelper.IsOtelOutput(op){
			schema := `
					.timeUnixNano = to_unix_timestamp(to_timestamp!(del(.@timestamp)))
					.severityText = del(.level)
			  
					# Convert syslog severity to number, default to 9 (unknown)
					if .severityText == "trace"{
						.severityNumber = "8"
					}else if .severityText == "debug"{
						.severityNumber = "7"
					}else if .severityText == "info"{
						.severityNumber = "6"
					}else if .severityText == "notice"{
						.severityNumber = "5"
					}else if .severityText == "warn"{
						.severityNumber = "4"
					}else if .severityText == "err"{
						.severityNumber = "3"
					}else if .severityText == "crit"{
						.severityNumber = "2"
					}else if .severityText == "alert"{
						.severityNumber = "1"
					}else if .severityText == "emerg"{
						.severityNumber = "0"
					}else{
						.severityNumber = "9"
					}
					
					# resources
					.resources.logs.file.path = del(.file)
					.resources.host.name= del(.hostname)
					.resources.container.name = del(.kubernetes.container_name)
					.resources.container.id = del(.kubernetes.container_id)
			  
					# split image name and tag into separate fields
					container_image_slice = split!(.kubernetes.container_image, ":", limit: 2)
					.resources.container.image.name = container_image_slice[0]
					.resources.container.image.tag = container_image_slice[1]
					del(.kubernetes.container_image)
			  
					#kuberenetes
					.resources.k8s.pod.name = del(.kubernetes.pod_name)
					.resources.k8s.pod.uid = del(.kubernetes.pod_id)
					.resources.k8s.pod.ip = del(.kubernetes.pod_ip)
					.resources.k8s.pod.owner = .kubernetes.pod_owner
					.resources.k8s.pod.annotations = del(.kubernetes.annotations)
					.resources.k8s.pod.labels = del(.kubernetes.labels)
					.resources.k8s.namespace.id = del(.kubernetes.namespace_id)
			  
					.resources.k8s.namespace.name = .kubernetes.namespace_labels."kubernetes.io/metadata.name"
					.resources.k8s.namespace.labels = del(.kubernetes.namespace_labels)
					.resources.attributes.key = "log_type"
					.resources.attributes.value = del(.log_type)
			  `
			vrls = append(vrls, schema)
		}
		vrl := SrcPassThrough
		if len(vrls) != 0 {
			vrl = strings.Join(helpers.TrimSpaces(vrls), "\n\n")
		}
		r := Remap{
			ComponentID: p.Name,
			Inputs:      helpers.MakeInputs(inputs...),
			VRL:         vrl,
		}
		el = append(el, r)

	}
	return el
}
