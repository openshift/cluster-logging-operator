package vector

import (
	"encoding/json"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	ParseJson = "json"
)

func Pipelines(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	el := []generator.Element{}
	userDefinedInputs := spec.InputMap()
	for i, p := range spec.Pipelines {
		p.Name = helpers.FormatComponentID(p.Name) + "_user_defined"
		spec.Pipelines[i].Name = p.Name
		inputs := []string{}
		for _, inputName := range p.InputRefs {
			input, isUserDefined := userDefinedInputs[inputName]
			if isUserDefined {
				if input.HasPolicy() && input.GetMaxRecordsPerSecond() >= 0 {
					// if threshold is zero, then don't add to pipeline
					inputs = append(inputs, fmt.Sprintf(UserDefinedSourceThrottle, input.Name))
				} else {
					if input.Application != nil {
						inputs = append(inputs, fmt.Sprintf(UserDefinedInput, inputName))
					}
					if input.Receiver != nil && input.Receiver.HTTP != nil && input.Receiver.HTTP.Format == logging.FormatKubeAPIAudit {
						inputs = append(inputs, input.Name+`_input`)
					}
				}
			} else {
				inputs = append(inputs, inputName)
			}
		}

		if p.DetectMultilineErrors {
			d := normalize.DetectExceptions{
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

		filters := spec.FilterMap()
		for _, filterName := range p.FilterRefs {
			if f, ok := filters[filterName]; ok {
				if vrl, err := filter.RemapVRL(f); err != nil {
					log.Error(err, "bad filter", "filter", filterName)
				} else {
					vrls = append(vrls, vrl)
				}
			}
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
