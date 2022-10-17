package k8shandler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"

	log "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/url"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func applyOutputDefaults(outputDefaults *logging.OutputDefaults, out logging.OutputSpec) logging.OutputSpec {

	if outputDefaults != nil && outputDefaults.Elasticsearch != nil {
		// TODO: fix - outputDefaults should only apply to default output
		if out.Elasticsearch == nil {
			out.Elasticsearch = outputDefaults.Elasticsearch
		} else {
			out.Elasticsearch = mergeDefaults(outputDefaults, out)
		}
	}
	out = setESVersion(out)
	return out
}

func mergeDefaults(outputDefaults *logging.OutputDefaults, out logging.OutputSpec) *logging.Elasticsearch {
	// Below is necessary since the output version has already been set for default
	outES := out.Elasticsearch
	defaults := outputDefaults.Elasticsearch

	if outES.StructuredTypeKey == "" && defaults.StructuredTypeKey != "" {
		outES.StructuredTypeKey = defaults.StructuredTypeKey
	}
	if outES.StructuredTypeName == "" && defaults.StructuredTypeName != "" {
		outES.StructuredTypeName = defaults.StructuredTypeName
	}
	if !outES.EnableStructuredContainerLogs && defaults.EnableStructuredContainerLogs {
		outES.EnableStructuredContainerLogs = defaults.EnableStructuredContainerLogs
	}
	return outES
}

func setESVersion(out logging.OutputSpec) logging.OutputSpec {
	// Always set the version for our default output
	if out.Name == logging.OutputNameDefault {
		if out.Elasticsearch != nil {
			out.Elasticsearch.Version = logging.DefaultESVersion
		} else {
			out.Elasticsearch = &logging.Elasticsearch{
				Version: logging.DefaultESVersion,
			}
		}
	}

	return out
}

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {
	if clusterRequest.Cluster == nil || clusterRequest.Cluster.Spec.Collection == nil {
		log.V(2).Info("skipping collection config generation as 'collection' section is not specified in CLO's CR")
		return "", nil
	}
	switch clusterRequest.Cluster.Spec.Collection.Type {
	case logging.LogCollectionTypeFluentd:
		break
	case logging.LogCollectionTypeVector:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.Cluster.Spec.Collection.Type)
	}

	if clusterRequest.ForwarderRequest == nil {
		clusterRequest.ForwarderRequest = &logging.ClusterLogForwarder{}
	}

	spec, status := clusterRequest.NormalizeForwarder()
	clusterRequest.ForwarderSpec = *spec
	clusterRequest.ForwarderRequest.Status = *status

	tokenSecret, err := clusterRequest.getLogCollectorServiceAccountTokenSecret()
	if err == nil {
		clusterRequest.OutputSecrets[constants.LogCollectorToken] = tokenSecret
	}

	op := generator.Options{}
	if clusterRequest.useOldRemoteSyslogPlugin() {
		op[generator.UseOldRemoteSyslogPlugin] = ""
	}
	if debug, ok := clusterRequest.ForwarderRequest.Annotations[AnnotationDebugOutput]; ok && strings.ToLower(debug) == "true" {
		op[helpers.EnableDebugOutput] = "true"
	}

	var collectorType = clusterRequest.Cluster.Spec.Collection.Type
	g := forwardergenerator.New(collectorType)
	err = g.Verify(clusterRequest.Cluster.Spec.Collection, clusterRequest.OutputSecrets, &clusterRequest.ForwarderSpec, op)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		if updateError := clusterRequest.UpdateCondition(
			logging.CollectorDeadEnd,
			"ClusterLogForwarder input validation error",
			"ClusterLogForwarder input validation error",
			corev1.ConditionTrue,
		); updateError != nil {
			log.Error(updateError, "Unable to update the clusterlogging status", "conditionType", logging.CollectorDeadEnd)
		}
		return "", err
	}

	generatedConfig, err := g.GenerateConf(clusterRequest.Cluster.Spec.Collection, clusterRequest.OutputSecrets, &clusterRequest.ForwarderSpec, op)

	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		if updateError := clusterRequest.UpdateCondition(
			logging.CollectorDeadEnd,
			"Collectors are defined but there is no defined LogStore or LogForward destinations",
			"No defined logstore destination",
			corev1.ConditionTrue,
		); updateError != nil {
			log.Error(updateError, "Unable to update the clusterlogging status", "conditionType", logging.CollectorDeadEnd)
		}
		return "", err
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
	clusterRequest.OutputSecrets = make(map[string]*corev1.Secret, len(clusterRequest.ForwarderSpec.Outputs))

	// Check for default configuration
	logStore := clusterRequest.Cluster.Spec.LogStore
	if len(clusterRequest.ForwarderSpec.Pipelines) == 0 {
		if logStore != nil && logStore.Type != "" {
			log.V(2).Info("ClusterLogForwarder forwarding to default store")
			clusterRequest.ForwarderSpec.Pipelines = []logging.PipelineSpec{
				{
					InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
					OutputRefs: []string{logging.OutputNameDefault},
				},
			}
			// Continue with normalization to fill out spec and status.
		} else if clusterRequest.ForwarderRequest == nil {
			log.V(3).Info("ClusterLogForwarder disabled")
			return &logging.ClusterLogForwarderSpec{}, &logging.ClusterLogForwarderStatus{}
		}
	}

	if logStore != nil && logStore.Type == logging.LogStoreTypeLokiStack {
		outputs, pipelines := clusterRequest.processPipelinesForLokiStack(clusterRequest.ForwarderSpec.Pipelines)

		clusterRequest.ForwarderSpec.Outputs = append(clusterRequest.ForwarderSpec.Outputs, outputs...)
		clusterRequest.ForwarderSpec.Pipelines = pipelines
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
			Name:                  pipeline.Name,
			InputRefs:             goodIn.List(),
			OutputRefs:            goodOut.List(),
			Labels:                pipeline.Labels,
			Parse:                 pipeline.Parse,
			DetectMultilineErrors: pipeline.DetectMultilineErrors,
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

// LokiStackGatewayService returns the name of LokiStack gateway service.
// Returns an empty string if ClusterLogging is not configured for a LokiStack log store.
func (clusterRequest *ClusterLoggingRequest) LokiStackGatewayService() string {
	logStore := clusterRequest.Cluster.Spec.LogStore
	if logStore == nil || logStore.LokiStack.Name == "" {
		return ""
	}

	return fmt.Sprintf("%s-gateway-http", logStore.LokiStack.Name)
}

// LokiStackURL returns the URL of the LokiStack API for a specific tenant.
// Returns an empty string if ClusterLogging is not configured for a LokiStack log store.
func (clusterRequest *ClusterLoggingRequest) LokiStackURL(tenant string) string {
	service := clusterRequest.LokiStackGatewayService()
	if service == "" {
		return ""
	}

	return fmt.Sprintf("https://%s.%s.svc:8080/api/logs/v1/%s", service, clusterRequest.Cluster.Namespace, tenant)
}

func (clusterRequest *ClusterLoggingRequest) verifyOutputs(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	status.Outputs = logging.NamedConditions{}
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
		case output.Type == logging.OutputTypeCloudwatch && output.Cloudwatch == nil:
			log.V(3).Info("verifyOutputs failed", "reason", "Cloudwatch output requires type spec", "output name", output.Name)
			status.Outputs.Set(output.Name, condInvalid("output %q: Cloudwatch output requires type spec", output.Name))
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
			logStore := clusterRequest.Cluster.Spec.LogStore
			switch logStore.Type {
			case logging.LogStoreTypeElasticsearch:
				spec.Outputs = append(spec.Outputs, logging.OutputSpec{
					Name:   logging.OutputNameDefault,
					Type:   logging.OutputTypeElasticsearch,
					URL:    constants.LogStoreURL,
					Secret: &logging.OutputSecretSpec{Name: constants.CollectorSecretName},
					OutputTypeSpec: logging.OutputTypeSpec{
						Elasticsearch: &logging.Elasticsearch{
							Version: logging.DefaultESVersion,
						},
					},
				})
				status.Outputs.Set(name, condReady)
			case logging.LogStoreTypeLokiStack:
				// The outputs for LokiStack will already have been added at this point
			default:
				status.Outputs.Set(name, condInvalid(fmt.Sprintf("unknown log store type: %s", clusterRequest.Cluster.Spec.LogStore.Type)))
			}
		}
	}
	for i, out := range spec.Outputs {
		if out.Type == logging.OutputTypeElasticsearch {
			// applyOutputDefaults only applies to Elasticsearch
			out = applyOutputDefaults(clusterRequest.ForwarderSpec.OutputDefaults, out)
		}
		spec.Outputs[i] = out
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
		// Some output types allow a missing URL
		// TODO (alanconway) move output-specific valiation to the output implementation.
		if output.Type == logging.OutputTypeKafka || output.Type == logging.OutputTypeCloudwatch || output.Type == logging.OutputTypeGoogleCloudLogging {
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

	// Ensure we have secrets for valid cloudwatch config
	hasKeyID := len(secret.Data[constants.AWSAccessKeyID]) > 0
	hasSecretKey := len(secret.Data[constants.AWSSecretAccessKey]) > 0
	hasRoleArnKey := security.HasAwsRoleArnKey(secret)
	hasCredentialsKey := security.HasAwsCredentialsKey(secret)
	hasValidRoleArn := len(cloudwatch.ParseRoleArn(secret)) > 0
	switch {
	case hasValidRoleArn: // Sts secret format is the first check
		return true
	case hasRoleArnKey && !hasValidRoleArn, hasCredentialsKey && !hasValidRoleArn:
		return fail(condMissing("auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"))
	case !hasKeyID || !hasSecretKey:
		return fail(condMissing("auth keys: " + constants.AWSAccessKeyID + " and " + constants.AWSSecretAccessKey + " are required"))
	}
	return true
}

func (clusterRequest *ClusterLoggingRequest) getLogCollectorServiceAccountTokenSecret() (*corev1.Secret, error) {
	s := &corev1.Secret{}
	log.V(9).Info("Fetching Secret", "Name", constants.LogCollectorToken)
	if err := clusterRequest.Get(constants.LogCollectorToken, s); err != nil {
		log.V(3).Error(err, "Could not find Secret", "Name", constants.LogCollectorToken)
		return nil, errors.New("Could not retrieve ServiceAccount token")
	}

	if _, ok := s.Data[constants.TokenKey]; !ok {
		log.V(9).Info("did not find token in secret", "Name", s.Name)
		return nil, errors.New("logcollector secret is missing token")
	}

	return s, nil
}
