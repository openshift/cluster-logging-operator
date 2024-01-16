package clusterlogforwarder

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/url"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ValidateInputsOutputsPipelines all inputs, outputs, and pipelines without mutating the spec
func ValidateInputsOutputsPipelines(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {

	status := &loggingv1.ClusterLogForwarderStatus{}

	// Check if any defined pipelines and if a clusterLogForwarder instance is available
	if len(clf.Spec.Pipelines) == 0 {
		log.V(3).Info("ClusterLogForwarder disabled")
		return errors.NewValidationError("ClusterLogForwarder disabled"), status
	}

	verifyInputs(&clf.Spec, status, extras)
	if !status.Inputs.IsAllReady() {
		log.V(3).Info("Input not Ready", "inputs", status.Inputs)
	}
	verifyOutputs(clf.Namespace, k8sClient, &clf.Spec, status, extras)
	if !status.Outputs.IsAllReady() {
		log.V(3).Info("Output not Ready", "outputs", status.Outputs)
	}
	verifyPipelines(clf.Name, &clf.Spec, status)
	if !status.Pipelines.IsAllReady() {
		log.V(3).Info("Pipeline not Ready", "pipelines", status.Pipelines)
	}

	// Check all pipeline statuses
	unready := []string{}
	for name, conds := range status.Pipelines {
		if !conds.IsTrueFor(loggingv1.ConditionReady) {
			unready = append(unready, name)
		}
	}

	// All pipelines have to be ready or invalid CLF
	if len(unready) > 0 {
		log.V(3).Info("validate clusterlogforwarder. Not all pipelines valid. Invalid CLF", "ForwarderSpec", clf.Spec)
		status.Conditions.SetCondition(CondInvalid("invalid clf spec; one or more errors present: %v", unready))
		return errors.NewValidationError("clusterlogforwarder is not ready"), status
	}
	status.Conditions.SetCondition(condReady)
	return nil, status
}

// verifyRefs returns the set of valid refs and a slice of error messages for bad refs.
func verifyRefs(what, forwarderName string, status loggingv1.ClusterLogForwarderStatus, refs []string, allowed sets.String, required bool) (sets.String, []string) {

	good, bad := sets.NewString(), sets.NewString()

	msg := []string{}
	for _, ref := range refs {
		switch what {
		case "inputs":
			if loggingv1.ReservedInputNames.Has(ref) ||
				(allowed.Has(ref) && status.Inputs[ref].IsTrueFor(loggingv1.ConditionReady)) {
				good.Insert(ref)
			} else {
				bad.Insert(ref)
			}
		case "outputs":
			// Check if custom CLF is forwarding to default store
			if forwarderName != constants.SingletonName && ref == loggingv1.OutputNameDefault {
				msg = append(msg, "custom ClusterLogForwarders cannot forward to the `default` log store")
				bad.Insert(ref)
			} else if allowed.Has(ref) && status.Outputs[ref].IsTrueFor(loggingv1.ConditionReady) {
				good.Insert(ref)
			} else {
				bad.Insert(ref)
			}
		case "filters":
			if allowed.Has(ref) {
				good.Insert(ref)
			} else {
				bad.Insert(ref)
			}
		}
	}

	if bad.Len() > 0 {
		msg = append(msg, fmt.Sprintf("unrecognized %s: %v", what, bad.List()))
	}
	if required && good.Len() == 0 {
		msg = append(msg, fmt.Sprintf("no valid %s", what))
	}
	return *good, msg
}

func verifyPipelines(forwarderName string, spec *loggingv1.ClusterLogForwarderSpec, status *loggingv1.ClusterLogForwarderStatus) {
	// Validate each pipeline and add a status object.
	status.Pipelines = loggingv1.NamedConditions{}
	names := sets.NewString() // Collect pipeline names

	// Known output names, note if "default" is enabled it will already be in the OutputMap()
	outputs := *sets.NewString()
	for k := range spec.OutputMap() {
		outputs.Insert(k)
	}
	// Known input names, reserved names not in InputMap() we don't expose default inputs.
	inputs := *sets.NewString()
	for k := range spec.InputMap() {
		inputs.Insert(k)
	}
	// Known filter names
	filters := *sets.NewString()
	for k := range spec.FilterMap() {
		filters.Insert(k)
	}
	for i, pipeline := range spec.Pipelines {

		// Don't allow empty names since this no longer mutates the spec
		if pipeline.Name == "" {
			pipeline.Name = fmt.Sprintf("pipeline_%v_", i)
			status.Pipelines.Set(pipeline.Name, CondInvalid("pipeline must have a name"))
			names.Insert(pipeline.Name)
			continue
		}

		if names.Has(pipeline.Name) {
			original := pipeline.Name
			pipeline.Name = fmt.Sprintf("pipeline_%v_", i)
			status.Pipelines.Set(pipeline.Name, CondInvalid("duplicate name %q", original))
			continue
		}
		names.Insert(pipeline.Name)

		// Verify pipeline labels
		if _, err := json.Marshal(pipeline.Labels); err != nil {
			status.Pipelines.Set(pipeline.Name, CondInvalid("invalid pipeline labels"))
			continue
		}

		_, msgIn := verifyRefs("inputs", forwarderName, *status, pipeline.InputRefs, inputs, true)
		_, msgOut := verifyRefs("outputs", forwarderName, *status, pipeline.OutputRefs, outputs, true)
		_, msgFilter := verifyRefs("filters", forwarderName, *status, pipeline.FilterRefs, filters, false)

		// Pipelines must all be valid for CLF to be considered valid
		// Partially valid pipelines invalidate CLF
		if msgs := append(msgIn, append(msgOut, msgFilter...)...); len(msgs) > 0 { // Something wrong
			msg := strings.Join(msgs, ", ")
			con := loggingv1.NewCondition(loggingv1.ValidationCondition, corev1.ConditionTrue, loggingv1.ValidationFailureReason, "invalid: %v", msg)
			status.Pipelines.Set(pipeline.Name, con)
			continue
		} else {
			status.Pipelines.Set(pipeline.Name, condReady) // Ready
		}
	}
}

// verifyInputs and set status.Inputs conditions
func verifyInputs(spec *loggingv1.ClusterLogForwarderSpec, status *loggingv1.ClusterLogForwarderStatus, extras map[string]bool) {
	// Collect input conditions
	status.Inputs = loggingv1.NamedConditions{}

	// Check input names
	for i, input := range spec.Inputs {
		i, input := i, input // Don't bind range variables.
		badInput := func(format string, args ...interface{}) {
			if input.Name == "" {
				input.Name = fmt.Sprintf("input_%v_", i)
			}
			status.Inputs.Set(input.Name, CondInvalid(format, args...))
		}

		validPort := func(port int32) bool {
			return port == 0 || (port > 1023 && port < 65536)
		}
		isHTTPReceiver := func(input loggingv1.InputSpec) bool {
			return input.Receiver != nil && input.Receiver.HTTP != nil
		}

		switch {
		case input.Name == "":
			badInput("input must have a name")
		case loggingv1.ReservedInputNames.Has(input.Name):
			badInput("input name %q is reserved", input.Name)
		case len(status.Inputs[input.Name]) > 0:
			badInput("duplicate name: %q", input.Name)
		// Check if inputspec has application, infrastructure, audit or receiver specs
		case input.Application == nil && input.Infrastructure == nil && input.Audit == nil && input.Receiver == nil:
			badInput("inputspec must define one or more of application, infrastructure, audit or receiver")
		case input.HasPolicy() && input.Application.ContainerLimit != nil && input.Application.GroupLimit != nil:
			badInput("inputspec must define only one of container or group limit")
		case input.HasPolicy() && input.GetMaxRecordsPerSecond() < 0:
			badInput("inputspec cannot have a negative limit threshold")
		case input.Receiver != nil && !extras[constants.VectorName]:
			badInput("ReceiverSpecs are only supported for the vector log collector")
		case input.Receiver != nil && input.Receiver.HTTP == nil:
			badInput("ReceiverSpec must define an HTTP receiver")
		case isHTTPReceiver(input) && input.Receiver.HTTP.Format != loggingv1.FormatKubeAPIAudit:
			badInput("invalid format specified for HTTP receiver")
		case isHTTPReceiver(input) && !validPort(input.Receiver.HTTP.Port):
			badInput("invalid port specified for HTTP receiver")
		default:
			status.Inputs.Set(input.Name, condReady)
		}
	}
}

func verifyOutputs(namespace string, clfClient client.Client, spec *loggingv1.ClusterLogForwarderSpec, status *loggingv1.ClusterLogForwarderStatus, extras map[string]bool) {
	outputRefs := sets.NewString()
	for _, p := range spec.Pipelines {
		outputRefs.Insert(p.OutputRefs...)
	}

	status.Outputs = loggingv1.NamedConditions{}
	names := sets.NewString() // Collect pipeline names
	for i, output := range spec.Outputs {
		i, output := i, output // Don't bind range variable.
		badName := func(format string, args ...interface{}) {
			output.Name = fmt.Sprintf("output_%v_", i)
			status.Outputs.Set(output.Name, CondInvalid(format, args...))
		}
		log.V(3).Info("Verifying", "outputs", output)
		switch {
		case output.Name == "":
			log.V(3).Info("verifyOutputs failed", "reason", "output must have a name")
			badName("output must have a name")
		case loggingv1.IsReservedOutputName(output.Name) && !extras[constants.MigrateDefaultOutput]:
			// adding check for our replaced spec during migration (ES only)
			log.V(3).Info("verifyOutputs failed", "reason", "output name is reserved", "output name", output.Name)
			badName("output name %q is reserved", output.Name)
		case names.Has(output.Name):
			log.V(3).Info("verifyOutputs failed", "reason", "output name is duplicated", "output name", output.Name)
			badName("duplicate name: %q", output.Name)
		case !loggingv1.IsOutputTypeName(output.Type):
			log.V(3).Info("verifyOutputs failed", "reason", "output type is invalid", "output name", output.Name, "output type", output.Type)
			status.Outputs.Set(output.Name, CondInvalid("output %q: unknown output type %q", output.Name, output.Type))
		case !verifyOutputURL(&output, status.Outputs):
			log.V(3).Info("verifyOutputs failed", "reason", "output URL is invalid", "output URL", output.URL)
		case !verifyOutputSecret(namespace, clfClient, &output, status.Outputs, extras):
			log.V(3).Info("verifyOutputs failed", "reason", "output secret is invalid")
		case output.Type == loggingv1.OutputTypeCloudwatch && output.Cloudwatch == nil:
			log.V(3).Info("verifyOutputs failed", "reason", "Cloudwatch output requires type spec", "output name", output.Name)
			status.Outputs.Set(output.Name, CondInvalid("output %q: Cloudwatch output requires type spec", output.Name))
		// Check googlecloudlogging specs, must only include one of the following
		case output.Type == loggingv1.OutputTypeGoogleCloudLogging && output.GoogleCloudLogging != nil && !verifyGoogleCloudLogging(output.GoogleCloudLogging):
			log.V(3).Info("verifyOutputs failed", "reason",
				"Exactly one of billingAccountId, folderId, organizationId, or projectId must be set.",
				"output name", output.Name, "output type", output.Type)
			status.Outputs.Set(output.Name,
				CondInvalid("output %q: Exactly one of billingAccountId, folderId, organizationId, or projectId must be set.",
					output.Name))
		case output.HasPolicy() && output.GetMaxRecordsPerSecond() < 0:
			status.Outputs.Set(output.Name, CondInvalid("output %q: Output cannot have negative limit threshold", output.Name))
		case !outputRefs.Has(output.Name):
			status.Outputs.Set(output.Name, CondInvalid("output %q: Output not referenced by any pipeline", output.Name))
		default:
			status.Outputs.Set(output.Name, condReady)
		}

		if output.Type == loggingv1.OutputTypeCloudwatch {
			if output.Cloudwatch != nil && output.Cloudwatch.GroupPrefix == nil {
				clusterName, err := readClusterName(clfClient)
				if err != nil {
					badName("outputprefix is not set and it can't be fetched from the cluster. Error: %s", err.Error())
				} else {
					output.Cloudwatch.GroupPrefix = &clusterName
				}
			}
		}
		names.Insert(output.Name)
	}
}

func verifyGoogleCloudLogging(gcl *loggingv1.GoogleCloudLogging) bool {
	i := 0
	if gcl.ProjectID != "" {
		i += 1
	}
	if gcl.FolderID != "" {
		i += 1
	}
	if gcl.BillingAccountID != "" {
		i += 1
	}
	if gcl.OrganizationID != "" {
		i += 1
	}
	return i == 1
}

func verifyOutputURL(output *loggingv1.OutputSpec, conds loggingv1.NamedConditions) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}

	if output.Type == loggingv1.OutputTypeKafka {
		brokerUrls := []string{}
		if output.URL != "" {
			brokerUrls = append(brokerUrls, output.URL)
		}
		if output.Kafka != nil { // Add optional extra broker URLs.
			brokerUrls = append(brokerUrls, output.Kafka.Brokers...)
		}
		if len(brokerUrls) == 0 {
			return fail(CondInvalid("no broker URLs specified"))
		}
		for _, b := range brokerUrls {
			u, err := url.Parse(b)
			if err == nil {
				err = url.CheckAbsolute(u)
			}
			if err != nil {
				return fail(CondInvalid("invalid URL: %v", err))
			}
		}
	} else {
		if output.URL == "" {
			// Some output types allow a missing URL
			// TODO (alanconway) move output-specific valiation to the output implementation.
			if output.Type == loggingv1.OutputTypeCloudwatch || output.Type == loggingv1.OutputTypeGoogleCloudLogging {
				return true
			} else {
				return fail(CondInvalid("URL is required for output type %v", output.Type))
			}
		}
		u, err := url.Parse(output.URL)
		if err == nil {
			err = url.CheckAbsolute(u)
		}
		if err != nil {
			return fail(CondInvalid("invalid URL: %v", err))
		}
		if output.Type == loggingv1.OutputTypeSyslog {
			scheme := strings.ToLower(u.Scheme)
			if !(scheme == `tcp` || scheme == `tls` || scheme == `udp`) {
				return fail(CondInvalid("invalid URL scheme: %v", u.Scheme))
			}
		}
	}
	return true
}

func verifyOutputSecret(namespace string, clfClient client.Client, output *loggingv1.OutputSpec, conds loggingv1.NamedConditions, extras map[string]bool) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}
	if output.Secret == nil {
		return true
	}
	if output.Secret.Name == "" {
		conds.Set(output.Name, CondInvalid("secret has empty name"))
		return false
	}
	// Only for ES. If default replaced, the "collector" secret will be created later
	if (output.Type == loggingv1.OutputTypeElasticsearch || output.Type == loggingv1.OutputTypeLoki) && extras[constants.MigrateDefaultOutput] {
		return true
	}
	log.V(3).Info("getting output secret", "output", output.Name, "secret", output.Secret.Name)
	secret, err := getOutputSecret(namespace, clfClient, output.Secret.Name)
	if err != nil {
		return fail(CondMissing("secret %q not found", output.Secret.Name))
	}

	switch output.Type {
	case loggingv1.OutputTypeCloudwatch:
		if !verifySecretKeysForCloudwatch(output, conds, secret) {
			return false
		}
	case loggingv1.OutputTypeSplunk:
		if !verifySecretKeysForSplunk(output, conds, secret) {
			return false
		}
	}
	return verifySecretKeysForTLS(output, conds, secret)
}

func getOutputSecret(namespace string, clfClient client.Client, secretName string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	namespacedName := types.NamespacedName{Name: secretName, Namespace: namespace}

	log.V(3).Info("Getting object", "namespacedName", namespacedName, "object", secret)

	err := clfClient.Get(context.TODO(), namespacedName, secret)
	return secret, err
}

func verifySecretKeysForTLS(output *loggingv1.OutputSpec, conds loggingv1.NamedConditions, secret *corev1.Secret) bool {
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
		return fail(CondMissing("cannot have %v without %v", constants.ClientCertKey, constants.ClientPrivateKey))
	case !haveCert && haveKey:
		return fail(CondMissing("cannot have %v without %v", constants.ClientPrivateKey, constants.ClientCertKey))
	case haveUsername && !havePassword:
		return fail(CondMissing("cannot have %v without %v", constants.ClientUsername, constants.ClientPassword))
	case !haveUsername && havePassword:
		return fail(CondMissing("cannot have %v without %v", constants.ClientPassword, constants.ClientUsername))
	}
	return true
}

func verifySecretKeysForCloudwatch(output *loggingv1.OutputSpec, conds loggingv1.NamedConditions, secret *corev1.Secret) bool {
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
		return fail(CondMissing("auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"))
	case !hasKeyID || !hasSecretKey:
		return fail(CondMissing("auth keys: " + constants.AWSAccessKeyID + " and " + constants.AWSSecretAccessKey + " are required"))
	}
	return true
}

func verifySecretKeysForSplunk(output *loggingv1.OutputSpec, conds loggingv1.NamedConditions, secret *corev1.Secret) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}

	if len(secret.Data[constants.SplunkHECTokenKey]) > 0 {
		return true
	} else {
		return fail(CondMissing("A non-empty " + constants.SplunkHECTokenKey + " entry is required"))
	}
}

func readClusterName(clfClient client.Client) (string, error) {
	infra := configv1.Infrastructure{}
	err := clfClient.Get(context.Background(), client.ObjectKey{Name: constants.ClusterInfrastructureInstance}, &infra)
	if err != nil {
		return "", err
	}

	return infra.Status.InfrastructureName, nil
}
