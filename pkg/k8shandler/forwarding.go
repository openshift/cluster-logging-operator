package k8shandler

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {
	switch clusterRequest.cluster.Spec.Collection.Logs.Type {
	case logging.LogCollectionTypeFluentd:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.cluster.Spec.Collection.Logs.Type)
	}

	spec, status := clusterRequest.normalizeForwarder()
	clusterRequest.ForwarderSpec = *spec
	clusterRequest.ForwarderRequest.Status = *status

	generator, err := forwarding.NewConfigGenerator(
		clusterRequest.cluster.Spec.Collection.Logs.Type,
		clusterRequest.includeLegacyForwardConfig(),
		clusterRequest.includeLegacySyslogConfig(),
		clusterRequest.useOldRemoteSyslogPlugin())
	if err != nil {
		logger.Warnf("Unable to create collector config generator: %v", err)
		return "",
			clusterRequest.UpdateCondition(
				logging.CollectorDeadEnd,
				"Unable to generate collector configuration",
				"No defined logstore destination",
				corev1.ConditionTrue,
			)
	}
	generatedConfig, err := generator.Generate(&clusterRequest.ForwarderSpec)
	if err != nil {
		logger.Warnf("Unable to generate log configuration: %v", err)
		return "",
			clusterRequest.UpdateCondition(
				logging.CollectorDeadEnd,
				"Collectors are defined but there is no defined LogStore or LogForward destinations",
				"No defined logstore destination",
				corev1.ConditionTrue,
			)
	}
	// else
	err = clusterRequest.UpdateCondition(
		logging.CollectorDeadEnd,
		"",
		"",
		corev1.ConditionFalse,
	)

	return generatedConfig, err
}

// normalizeForwarder normalizes the clusterRequest.ForwarderSpec, returns a normalized spec and status.
func (clusterRequest *ClusterLoggingRequest) normalizeForwarder() (*logging.ClusterLogForwarderSpec, *logging.ClusterLogForwarderStatus) {
	logger.DebugObject("Normalizing ClusterLogForwarder from request: %v", clusterRequest)

	// Check for default configuration
	if len(clusterRequest.ForwarderSpec.Pipelines) == 0 {
		if clusterRequest.cluster.Spec.LogStore != nil && clusterRequest.cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {
			logger.Debug("ClusterLogForwarder forwarding to default store")
			clusterRequest.ForwarderSpec.Pipelines = []logging.PipelineSpec{
				{
					InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
					OutputRefs: []string{logging.OutputNameDefault},
				},
			}
			// Continue with normalization to fill out spec and status.
		} else {
			if clusterRequest.ForwarderRequest == nil {
				logger.Debug("ClusterLogForwarder disabled")
				return &logging.ClusterLogForwarderSpec{}, &logging.ClusterLogForwarderStatus{}
			}
		}
	}

	spec := &logging.ClusterLogForwarderSpec{}
	status := &logging.ClusterLogForwarderStatus{}

	clusterRequest.verifyInputs(spec, status)
	clusterRequest.verifyOutputs(spec, status)
	clusterRequest.verifyPipelines(spec, status)

	routes := logging.NewRoutes(spec.Pipelines) // Compute used inputs/outputs

	// Add Ready=true status for all surviving inputs.
	status.Inputs = logging.NamedConditions{}
	inRefs := sets.StringKeySet(routes.ByInput).List()
	for _, inRef := range inRefs {
		status.Inputs.Set(inRef, condReady)
	}

	// Determine overall health
	degraded := []string{}
	unready := []string{}
	for name, conds := range status.Pipelines {
		if !conds.IsTrueFor(logging.ConditionReady) {
			unready = append(unready, name)
		}
		if conds.IsTrueFor(logging.ConditionDegraded) {
			degraded = append(degraded, name)
		}
	}
	if len(unready) == len(status.Pipelines) {
		status.Conditions.SetCondition(condInvalid("all pipelines invalid: %v", unready))
	} else {
		if len(unready)+len(degraded) > 0 {
			status.Conditions.SetCondition(condDegraded(logging.ReasonInvalid, "degraded pipelines: invalid %v, degraded %v", unready, degraded))
			logger.Infof("ClusterLogForwarder degraded")
		}
		status.Conditions.SetCondition(condReady)
		logger.Infof("ClusterLogForwarder is ready")
	}
	logger.DebugObject("ClusterLogForwarder normalized spec: %v", spec)
	logger.DebugObject("ClusterLogForwarder normalized status: %v", status)
	return spec, status
}

func condNotReady(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return logging.NewCondition(logging.ConditionReady, corev1.ConditionFalse, r, format, args...)
}

func condDegraded(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return logging.NewCondition(logging.ConditionDegraded, corev1.ConditionTrue, r, format, args...)
}

func condInvalid(format string, args ...interface{}) status.Condition {
	return condNotReady(logging.ReasonInvalid, format, args...)
}

var condReady = status.Condition{Type: logging.ConditionReady, Status: corev1.ConditionTrue}

// verifyRefs returns the set of valid refs and a slice of error messages for bad refs.
func verifyRefs(what string, refs []string, allowed sets.String) (sets.String, []string) {
	good, bad := sets.NewString(), sets.NewString()
	for _, ref := range refs {
		if allowed.Has(ref) {
			good.Insert(ref)
		} else {
			bad.Insert(ref)
		}
	}
	msg := []string{}
	if len(bad) > 0 {
		msg = append(msg, fmt.Sprintf("unrecognized %s: %v", what, bad.List()))
	}
	if len(good) == 0 {
		msg = append(msg, fmt.Sprintf("no valid %s", what))
	}
	return good, msg
}

func (clusterRequest *ClusterLoggingRequest) verifyPipelines(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	// Validate each pipeline and add a status object.
	status.Pipelines = logging.NamedConditions{}
	names := sets.NewString() // Collect pipeline names

	// Known output names, note if "default" is enabled it will already be in the OutputMap()
	outputs := sets.StringKeySet(spec.OutputMap())
	// Known input names, reserved names not in InputMap() we don't expose default inputs.
	inputs := sets.StringKeySet(spec.InputMap()).Union(logging.ReservedInputNames)

	for i, pipeline := range clusterRequest.ForwarderSpec.Pipelines {
		if pipeline.Name == "" {
			pipeline.Name = fmt.Sprintf("pipeline[%v]", i)
		}
		if names.Has(pipeline.Name) {
			original := pipeline.Name
			pipeline.Name = fmt.Sprintf("pipeline[%v]", i)
			status.Pipelines.Set(pipeline.Name, condInvalid("duplicate name %q", original))
			continue
		}
		names.Insert(pipeline.Name)

		goodIn, msgIn := verifyRefs("inputs", pipeline.InputRefs, inputs)
		goodOut, msgOut := verifyRefs("outputs", pipeline.OutputRefs, outputs)
		if msgs := append(msgIn, msgOut...); len(msgs) > 0 { // Something wrong
			msg := strings.Join(msgs, ", ")
			if len(goodIn) == 0 || len(goodOut) == 0 { // All bad, disabled
				status.Pipelines.Set(pipeline.Name, condInvalid("invalid: %v", msg))
				continue
			} else { // Some good some bad, degrade the pipeline.
				status.Pipelines.Set(pipeline.Name, condDegraded(logging.ReasonInvalid, "invalid: %v", msg))
			}
		}
		status.Pipelines.Set(pipeline.Name, condReady) // Ready, possibly degraded.
		spec.Pipelines = append(spec.Pipelines, logging.PipelineSpec{
			Name: pipeline.Name, InputRefs: goodIn.List(), OutputRefs: goodOut.List(),
		})
	}
}

// verifyInputs and set status.Inputs conditions
func (clusterRequest *ClusterLoggingRequest) verifyInputs(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	// Collect input conditions
	status.Inputs = logging.NamedConditions{}
	for i, input := range clusterRequest.ForwarderSpec.Inputs {
		badName := func(format string, args ...interface{}) {
			input.Name = fmt.Sprintf("input[%v]", i)
			status.Inputs.Set(input.Name, condInvalid(format, args...))
		}
		switch {
		case input.Name == "":
			badName("input must have a name")
		case logging.ReservedInputNames.Has(input.Name):
			badName("input name %q is reserved", input.Name)
		case len(status.Inputs[input.Name]) > 0:
			badName("duplicate name: %q", input.Name)
		default:
			status.Inputs.Set(input.Name, condReady)
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) verifyOutputs(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	status.Outputs = logging.NamedConditions{}
	names := sets.NewString() // Collect pipeline names
	for i, output := range clusterRequest.ForwarderSpec.Outputs {
		badName := func(format string, args ...interface{}) {
			output.Name = fmt.Sprintf("output[%v]", i)
			status.Outputs.Set(output.Name, condInvalid(format, args...))
		}
		switch {
		case output.Name == "":
			badName("output must have a name")
		case logging.IsReservedOutputName(output.Name):
			badName("output name %q is reserved", output.Name)
		case names.Has(output.Name):
			badName("duplicate name: %q", output.Name)
		case !logging.IsOutputTypeName(output.Type):
			status.Outputs.Set(output.Name, condInvalid("output %q: unknown output type %q", output.Name, output.Type))
		case output.URL == "":
			status.Outputs.Set(output.Name, condInvalid("output %q: missing URL", output.Name))
		case !clusterRequest.verifyOutputSecret(&output, status.Outputs):
			break
		default:
			status.Outputs.Set(output.Name, condReady)
			spec.Outputs = append(spec.Outputs, output)
		}
		names.Insert(output.Name)
	}
	// Add the default output if required and available.
	routes := logging.NewRoutes(clusterRequest.ForwarderSpec.Pipelines)
	name := logging.OutputNameDefault
	if _, ok := routes.ByOutput[name]; ok {
		if clusterRequest.cluster.Spec.LogStore == nil {
			status.Outputs.Set(name, condNotReady(logging.ReasonMissingResource, "no default log store specified"))
		} else {
			spec.Outputs = append(spec.Outputs, logging.OutputSpec{
				Name:   logging.OutputNameDefault,
				Type:   logging.OutputTypeElasticsearch,
				URL:    constants.LogStoreURL,
				Secret: &logging.OutputSecretSpec{Name: constants.CollectorSecretName},
			})
			status.Outputs.Set(name, condReady)
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) verifyOutputSecret(output *logging.OutputSpec, conds logging.NamedConditions) bool {
	if output.Secret == nil {
		return true
	}
	if output.Secret.Name == "" {
		conds.Set(output.Name, condInvalid("secret has empty name"))
		return false
	}
	if _, err := clusterRequest.GetSecret(output.Secret.Name); err != nil {
		conds.Set(output.Name, condNotReady(logging.ReasonMissingResource, "output %q: secret %q not found", output.Name, output.Secret.Name))
		return false
	}
	return true
}
