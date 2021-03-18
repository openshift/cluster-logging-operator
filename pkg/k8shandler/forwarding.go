package k8shandler

import (
	"fmt"
	"strings"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/status"
	"github.com/openshift/cluster-logging-operator/pkg/url"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {

	if clusterRequest.Cluster == nil || clusterRequest.Cluster.Spec.Collection == nil {
		log.V(2).Info("skipping collection config generation as 'collection' section is not specified in the CLO's CR")
		return "", nil
	}

	switch clusterRequest.Cluster.Spec.Collection.Logs.Type {
	case logging.LogCollectionTypeFluentd:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.Cluster.Spec.Collection.Logs.Type)
	}

	if clusterRequest.ForwarderRequest == nil {
		clusterRequest.ForwarderRequest = &logging.ClusterLogForwarder{}
	}

	spec, status := clusterRequest.NormalizeForwarder()
	clusterRequest.ForwarderSpec = *spec
	clusterRequest.ForwarderRequest.Status = *status

	generator, err := forwarding.NewConfigGenerator(
		clusterRequest.Cluster.Spec.Collection.Logs.Type,
		clusterRequest.includeLegacyForwardConfig(),
		clusterRequest.includeLegacySyslogConfig(),
		clusterRequest.useOldRemoteSyslogPlugin(),
	)

	if err != nil {
		log.Error(err, "Unable to create collector config generator")
		return "",
			clusterRequest.UpdateCondition(
				logging.CollectorDeadEnd,
				"Unable to generate collector configuration",
				"No defined logstore destination",
				corev1.ConditionTrue,
			)
	}

	clfSpec := &clusterRequest.ForwarderSpec
	fwSpec := clusterRequest.Cluster.Spec.Forwarder
	generatedConfig, err := generator.Generate(clfSpec, clusterRequest.OutputSecrets, fwSpec)

	if err != nil {
		log.Error(err, "Unable to generate log configuration")
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
	log.V(3).Info("ClusterLogForwarder generated config", generatedConfig)
	return generatedConfig, err
}

// NormalizeForwarder normalizes the clusterRequest.ForwarderSpec, returns a normalized spec and status.
func (clusterRequest *ClusterLoggingRequest) NormalizeForwarder() (*logging.ClusterLogForwarderSpec, *logging.ClusterLogForwarderStatus) {

	// Check for default configuration
	if len(clusterRequest.ForwarderSpec.Pipelines) == 0 {
		if clusterRequest.Cluster.Spec.LogStore != nil && clusterRequest.Cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {
			log.V(2).Info("ClusterLogForwarder forwarding to default store")
			defaultPipeline := logging.PipelineSpec{
				InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
				OutputRefs: []string{logging.OutputNameDefault},
			}
			clusterRequest.ForwarderSpec.Pipelines = []logging.PipelineSpec{defaultPipeline}
			if clusterRequest.includeLegacySyslogConfig() {
				defaultPipeline.OutputRefs = append(defaultPipeline.OutputRefs, constants.LegacySyslog)
			}
			if clusterRequest.includeLegacyForwardConfig() {
				defaultPipeline.OutputRefs = append(defaultPipeline.OutputRefs, constants.LegacySecureforward)
			}
			// Continue with normalization to fill out spec and status.
		} else if clusterRequest.ForwarderRequest == nil {
			log.V(3).Info("ClusterLogForwarder disabled")
			return &logging.ClusterLogForwarderSpec{}, &logging.ClusterLogForwarderStatus{}
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
		}
		status.Conditions.SetCondition(condReady)
	}
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

func condMissing(format string, args ...interface{}) status.Condition {
	return condNotReady(logging.ReasonMissingResource, format, args...)
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
			pipeline.Name = fmt.Sprintf("pipeline_%v_", i)
		}
		if names.Has(pipeline.Name) {
			original := pipeline.Name
			pipeline.Name = fmt.Sprintf("pipeline_%v_", i)
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
			Name:       pipeline.Name,
			InputRefs:  goodIn.List(),
			OutputRefs: goodOut.List(),
			Labels:     pipeline.Labels,
			Parse:      pipeline.Parse,
		})
	}
}

// verifyInputs and set status.Inputs conditions
func (clusterRequest *ClusterLoggingRequest) verifyInputs(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	// Collect input conditions
	status.Inputs = logging.NamedConditions{}
	for i, input := range clusterRequest.ForwarderSpec.Inputs {
		i, input := i, input // Don't bind range variables.
		badName := func(format string, args ...interface{}) {
			input.Name = fmt.Sprintf("input_%v_", i)
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
			spec.Inputs = append(spec.Inputs, input)
			status.Inputs.Set(input.Name, condReady)
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) verifyOutputs(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	status.Outputs = logging.NamedConditions{}
	clusterRequest.OutputSecrets = make(map[string]*corev1.Secret, len(clusterRequest.ForwarderSpec.Outputs))
	names := sets.NewString() // Collect pipeline names
	for i, output := range clusterRequest.ForwarderSpec.Outputs {
		i, output := i, output // Don't bind range variable.
		badName := func(format string, args ...interface{}) {
			output.Name = fmt.Sprintf("output_%v_", i)
			status.Outputs.Set(output.Name, condInvalid(format, args...))
		}
		log.V(3).Info("Verifying", "outputs", output)
		switch {
		case output.Name == "":
			badName("output must have a name")
		case logging.IsReservedOutputName(output.Name):
			badName("output name %q is reserved", output.Name)
		case names.Has(output.Name):
			badName("duplicate name: %q", output.Name)
		case !logging.IsOutputTypeName(output.Type):
			status.Outputs.Set(output.Name, condInvalid("output %q: unknown output type %q", output.Name, output.Type))
		case !clusterRequest.verifyOutputURL(&output, status.Outputs):
			break
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
		if clusterRequest.Cluster.Spec.LogStore == nil {
			status.Outputs.Set(name, condMissing("no default log store specified"))
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

func (clusterRequest *ClusterLoggingRequest) verifyOutputURL(output *logging.OutputSpec, conds logging.NamedConditions) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}
	if output.URL == "" {
		// Some output types (currently just kafka) allow a missing URL
		// TODO (alanconway) move output-specific valiation to the output implementation.
		if output.Type == "kafka" {
			return true
		} else {
			return fail(condInvalid("URL is required for output type %v", output.Type))
		}
	}
	u, err := url.Parse(output.URL)
	if err != nil {
		return fail(condInvalid("invalid URL: %v", err))
	}
	if err := url.CheckAbsolute(u); err != nil {
		return fail(condInvalid("invalid URL: %v", err))
	}
	return true
}

func (clusterRequest *ClusterLoggingRequest) verifyOutputSecret(output *logging.OutputSpec, conds logging.NamedConditions) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}
	if output.Secret == nil {
		return true
	}
	if output.Secret.Name == "" {
		conds.Set(output.Name, condInvalid("secret has empty name"))
		return false
	}
	log.V(3).Info("getting output secret", "output", output.Name, "secret", output.Secret.Name)
	secret, err := clusterRequest.GetSecret(output.Secret.Name)
	if err != nil {
		return fail(condMissing("secret %q not found", output.Secret.Name))
	}
	// Make sure we have secrets for a valid TLS configuration.
	haveCert := len(secret.Data[constants.ClientCertKey]) > 0
	haveKey := len(secret.Data[constants.ClientPrivateKey]) > 0
	switch {
	case haveCert && !haveKey:
		return fail(condMissing("cannot have %v without %v", constants.ClientCertKey, constants.ClientPrivateKey))
	case !haveCert && haveKey:
		return fail(condMissing("cannot have %v without %v", constants.ClientPrivateKey, constants.ClientCertKey))
	}
	clusterRequest.OutputSecrets[output.Name] = secret
	return true
}
