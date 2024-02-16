package telemetry

import (
	"strconv"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func UpdateInfofromCLF(forwarder logging.ClusterLogForwarder) {

	var npipelines = 0
	var output *logging.OutputSpec
	var found bool

	//CLO got two custom resources CL, CFL, CLF here is meant for forwarding logs to third party systems

	//CLO CLF pipelines and set of output specs
	lgpipeline := forwarder.Spec.Pipelines
	outputs := forwarder.Spec.OutputMap()
	inputs := forwarder.Spec.InputMap()
	log.V(1).Info("OutputMap", "outputs", outputs)

	for _, pipeline := range lgpipeline {
		npipelines++
		log.V(1).Info("pipelines", "npipelines", npipelines)
		inref := pipeline.InputRefs
		outref := pipeline.OutputRefs

		Data.CLFInputType.Range(func(labelname, value interface{}) bool {
			log.V(1).Info("iter over labelnames", "labelname", labelname)
			Data.CLFInputType.Set(labelname.(string), constants.IsNotPresent) //reset to zero
			for _, inputtype := range inref {
				log.V(1).Info("iter over inputtype", "inputtype", inputtype)
				hasInput := false
				if inputtype == labelname {
					hasInput = true
					// Check for receiver input types if not app, infra, or audit
				} else {
					if inputSpec, ok := inputs[inputtype]; ok && inputSpec.Receiver != nil && labelname == inputSpec.Receiver.Type {
						hasInput = true
					}
				}
				if hasInput {
					log.V(1).Info("labelname and inputtype", "labelname", labelname, "inputtype", inputtype) //when matched print matched labelname with input type stated in CLF spec
					Data.CLFInputType.Set(labelname.(string), constants.IsPresent)                           //input type present in CLF spec
				}
			}
			return true // continue iterating
		})

		Data.CLFOutputType.Range(func(labelname, value interface{}) bool {
			log.V(1).Info("iter over labelnames", "labelname", labelname)
			Data.CLFOutputType.Set(labelname.(string), constants.IsNotPresent) //reset to zero
			for _, outputname := range outref {
				log.V(1).Info("iter over outref", "outputname", outputname)
				if outputname == "default" {
					Data.CLFOutputType.Set("default", constants.IsPresent)
					continue
				}
				output, found = outputs[outputname]
				if found {
					outputtype := output.Type
					if outputtype == labelname {
						log.V(1).Info("labelname and outputtype", "labelname", labelname, "outputtype", outputtype)
						Data.CLFOutputType.Set(labelname.(string), constants.IsPresent) //when matched print matched labelname with output type stated in CLF spec
					}
				}
			}
			return true // continue iterating
		})
		log.V(1).Info("post updating inputtype and outputtype")
		Data.CLFInfo.Set("pipelineInfo", strconv.Itoa(npipelines))
	}
}
