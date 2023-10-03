package next

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/filters/passthrough"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/registry"
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
				if input.Application != nil {
					inputs = append(inputs, fmt.Sprintf(common.UserDefinedInput, inputName))
				}
				if input.Receiver != nil && input.Receiver.HTTP != nil && input.Receiver.HTTP.Format == logging.FormatKubeAPIAudit {
					inputs = append(inputs, input.Name+`_input`)
				}
			} else {
				switch inputName {
				case logging.InputNameApplication:
					inputs = append(inputs, fmt.Sprintf("raw_%s_logs", logging.InputNameContainer))
				case logging.InputNameAudit:
					for _, a := range []string{common.RawK8sAuditLogs, common.RawHostAuditLogs, common.RawOpenshiftAuditLogs, common.RawOvnAuditLogs} {
						inputs = append(inputs, a)
					}
				case logging.InputNameInfrastructure:
					inputs = append(inputs, fmt.Sprintf("raw_%s_logs", logging.InputNameNode))
					inputs = append(inputs, fmt.Sprintf("raw_%s_logs", logging.InputNameContainer))
				default:
					inputs = append(inputs, fmt.Sprintf("raw_%s_logs", inputName))
				}
			}
		}
		log.V(4).Info("Filter refs", "pipeline", p.Name, "refs", p.FilterRefs)
		if len(p.FilterRefs) == 0 {
			el = append(el, passthrough.New(p.Name, inputs...))
		} else {
			fSpecs := spec.FilterMap()
			for _, filterName := range p.FilterRefs {
				if f := registry.LookupProto(filterName, fSpecs); f != nil {
					fElements := f.Elements(inputs, p, *spec, op)
					log.V(4).Info("filter Elements", "elements", fElements)
					el = append(el, fElements...)
					inputs = []string{f.TranformsName(p)}
				} else {
					log.V(4).Info("Filter is unknown", "name", filterName)
				}
			}
		}
	}
	return el
}
