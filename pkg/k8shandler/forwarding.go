package k8shandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/ViaQ/logerr/log"
	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/status"
	"github.com/openshift/cluster-logging-operator/pkg/url"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {

	if clusterRequest.Cluster == nil || clusterRequest.Cluster.Spec.Collection == nil {
		log.V(2).Info("skipping collection config generation as 'collection' section is not specified in the CLO's CR")
		return "", nil
	}
	fmt.Println("cluster is not nil")
	switch clusterRequest.Cluster.Spec.Collection.Logs.Type {
	case logging.LogCollectionTypeFluentd:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.Cluster.Spec.Collection.Logs.Type)
	}

	fmt.Println("type is fluentd")
	if clusterRequest.ForwarderRequest == nil {
		clusterRequest.ForwarderRequest = &logging.ClusterLogForwarder{}
	}

	spec, status := clusterRequest.NormalizeForwarder()
	fmt.Println("forwarder normalized", spec)
	clusterRequest.ForwarderSpec = *spec
	clusterRequest.ForwarderRequest.Status = *status
	fmt.Printf("!!! DEBUG FORWARDER SPEC AFTER NORMALIZER: %+v\r\n", *spec)

	if clusterRequest.ForwarderSpec.OutputDefaults != nil {
		defaultOutputSpecs := clusterRequest.ForwarderSpec.OutputDefaults
		if defaultOutputSpecs.Elasticsearch != nil {
			for i, out := range clusterRequest.ForwarderSpec.Outputs {
				// copy from defaults if output specific spec not present
				if out.Type == logging.OutputTypeElasticsearch && out.Elasticsearch == nil {
					out.Elasticsearch = defaultOutputSpecs.Elasticsearch
					out.Secret = &logging.OutputSecretSpec{
						Name: constants.CollectorSecretName,
					}
					secret, err := clusterRequest.GetSecret(constants.CollectorSecretName)
					if err != nil {
						log.V(2).Info("no secret for default output type")
					}
					clusterRequest.OutputSecrets[out.Name] = secret
					clusterRequest.ForwarderSpec.Outputs[i] = out
				}
			}
		}
	}

	generator, err := forwarding.NewConfigGenerator(
		clusterRequest.Cluster.Spec.Collection.Logs.Type,
		clusterRequest.includeLegacyForwardConfig(),
		clusterRequest.includeLegacySyslogConfig(),
		clusterRequest.useOldRemoteSyslogPlugin(),
	)

	fmt.Printf("generator result %+v\r\n", generator)
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

	fmt.Printf("!!! DEBUG OUTPUT SECRETS: %+v\r\n", clusterRequest.OutputSecrets)
	generatedConfig, err := generator.Generate(clfSpec, clusterRequest.OutputSecrets, fwSpec)
	fmt.Printf("!!! DEBUG OUTPUT GENERATOR RESULT: %s\r\n", generatedConfig)

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

func (clusterRequest *ClusterLoggingRequest) readClusterName() (string, error) {
	infra := configv1.Infrastructure{}
	err := clusterRequest.Client.Get(context.Background(), client.ObjectKey{Name: constants.ClusterInfrastructureInstance}, &infra)
	if err != nil {
		return "", err
	}

	return infra.Status.InfrastructureName, nil
}

// NormalizeForwarder normalizes the clusterRequest.ForwarderSpec, returns a normalized spec and status.
func (clusterRequest *ClusterLoggingRequest) NormalizeForwarder() (*logging.ClusterLogForwarderSpec, *logging.ClusterLogForwarderStatus) {
	if clusterRequest.CLFVerifier.VerifyOutputSecret == nil {
		clusterRequest.CLFVerifier.VerifyOutputSecret = clusterRequest.verifyOutputSecret
	}

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
	if !status.Inputs.IsAllReady() {
		log.V(3).Info("Input not Ready", "inputs", status.Inputs)
	}
	clusterRequest.verifyOutputs(spec, status)
	if !status.Outputs.IsAllReady() {
		log.V(3).Info("Output not Ready", "outputs", status.Outputs)
	}
	clusterRequest.verifyPipelines(spec, status)
	if !status.Pipelines.IsAllReady() {
		log.V(3).Info("Pipeline not Ready", "pipelines", status.Pipelines)
	}

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
		log.V(3).Info("NormalizeForwarder. All pipelines invalid", "ForwarderSpec", clusterRequest.ForwarderSpec)
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
			log.V(3).Info("verifyOutputs failed", "reason", "output must have a name")
			badName("output must have a name")
		case logging.IsReservedOutputName(output.Name):
			log.V(3).Info("verifyOutputs failed", "reason", "output name is reserved", "output name", output.Name)
			badName("output name %q is reserved", output.Name)
		case names.Has(output.Name):
			log.V(3).Info("verifyOutputs failed", "reason", "output name is duplicated", "output name", output.Name)
			badName("duplicate name: %q", output.Name)
		case !logging.IsOutputTypeName(output.Type):
			log.V(3).Info("verifyOutputs failed", "reason", "output type is invalid", "output name", output.Name, "output type", output.Type)
			status.Outputs.Set(output.Name, condInvalid("output %q: unknown output type %q", output.Name, output.Type))
		case !clusterRequest.verifyOutputURL(&output, status.Outputs):
			log.V(3).Info("verifyOutputs failed", "reason", "output URL is invalid", "output URL", output.URL)
		case !clusterRequest.verifyOutputSecret(&output, status.Outputs):
			log.V(3).Info("verifyOutputs failed", "reason", "output secret is invalid")
		case !clusterRequest.CLFVerifier.VerifyOutputSecret(&output, status.Outputs):
			break
		default:
			status.Outputs.Set(output.Name, condReady)
			spec.Outputs = append(spec.Outputs, output)
		}
		if output.Type == logging.OutputTypeCloudwatch {
			if output.Cloudwatch != nil && output.Cloudwatch.GroupPrefix == nil {
				clusterName, err := clusterRequest.readClusterName()
				if err != nil {
					badName("outputprefix is not set and it can't be fetched from the cluster. Error: %s", err.Error())
				} else {
					output.Cloudwatch.GroupPrefix = &clusterName
				}
			}
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

	if clusterRequest.ForwarderSpec.OutputDefaults != nil {
		spec.OutputDefaults = clusterRequest.ForwarderSpec.OutputDefaults
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
		if output.Type == logging.OutputTypeKafka || output.Type == logging.OutputTypeCloudwatch {
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
	verifySecret := verifySecretKeysForTLS
	if output.Type == logging.OutputTypeCloudwatch {
		verifySecret = verifySecretKeysForCloudwatch
	}
	if !verifySecret(output, conds, secret) {
		return false
	}
	clusterRequest.OutputSecrets[output.Name] = secret
	return true
}

func verifySecretKeysForTLS(output *logging.OutputSpec, conds logging.NamedConditions, secret *corev1.Secret) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}
	// Make sure we have secrets for a valid TLS configuration.
	haveCert := len(secret.Data[constants.ClientCertKey]) > 0
	haveKey := len(secret.Data[constants.ClientPrivateKey]) > 0
	haveUsername := len(secret.Data[constants.ClientUsername]) > 0
	havePassword := len(secret.Data[constants.ClientPassword]) > 0
	switch {
	case haveCert && !haveKey:
		return fail(condMissing("cannot have %v without %v", constants.ClientCertKey, constants.ClientPrivateKey))
	case !haveCert && haveKey:
		return fail(condMissing("cannot have %v without %v", constants.ClientPrivateKey, constants.ClientCertKey))
	case haveUsername && !havePassword:
		return fail(condMissing("cannot have %v without %v", constants.ClientUsername, constants.ClientPassword))
	case !haveUsername && havePassword:
		return fail(condMissing("cannot have %v without %v", constants.ClientPassword, constants.ClientUsername))
	}
	return true
}
func verifySecretKeysForCloudwatch(output *logging.OutputSpec, conds logging.NamedConditions, secret *corev1.Secret) bool {
	log.V(3).Info("V")
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}
	hasID := len(secret.Data[constants.AWSAccessKeyID]) > 0
	hasKey := len(secret.Data[constants.AWSSecretAccessKey]) > 0
	missingMessage := "aws_access_key_id and aws_secret_access_key are required"
	if !hasID || !hasKey {
		return fail(condMissing(missingMessage))
	}
	return true
}
