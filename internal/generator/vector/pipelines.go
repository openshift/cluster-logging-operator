package vector

import (
	"encoding/json"
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
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
