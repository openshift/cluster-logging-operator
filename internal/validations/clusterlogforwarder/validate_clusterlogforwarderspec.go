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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Validate all inputs, outputs, and pipelines without mutating the spec
func ValidateInputsOutputsPipelines(clusterlogging *loggingv1.ClusterLogging, clfClient client.Client,
	clfInstance *loggingv1.ClusterLogForwarder, clfSpec loggingv1.ClusterLogForwarderSpec, extras map[string]bool) *loggingv1.ClusterLogForwarderStatus {

	status := &loggingv1.ClusterLogForwarderStatus{}

	// Check if any defined pipelines and if a clusterLogForwarder instance is available
	if len(clfSpec.Pipelines) == 0 && clfInstance == nil {
		log.V(3).Info("ClusterLogForwarder disabled")
		return status
	}

	verifyInputs(&clfSpec, status)
	if !status.Inputs.IsAllReady() {
		log.V(3).Info("Input not Ready", "inputs", status.Inputs)
	}
	verifyOutputs(clusterlogging, clfClient, &clfSpec, status, extras)
	if !status.Outputs.IsAllReady() {
		log.V(3).Info("Output not Ready", "outputs", status.Outputs)
	}
	verifyPipelines(&clfSpec, status)
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
		log.V(3).Info("validate clusterlogforwarder. Not all pipelines valid. Invalid CLF", "ForwarderSpec", clfSpec)
		status.Conditions.SetCondition(CondInvalid("invalid clf spec; one or more errors present: %v", unready))
		// Everything was valid
	} else {
		status.Conditions.SetCondition(condReady)
	}

	return status
}

// verifyRefs returns the set of valid refs and a slice of error messages for bad refs.
func verifyRefs(what string, status loggingv1.ClusterLogForwarderStatus, refs []string, allowed sets.String) (sets.String, []string) {

	good, bad := sets.NewString(), sets.NewString()

	for _, ref := range refs {
		if what == "inputs" {
			if loggingv1.ReservedInputNames.Has(ref) ||
				(allowed.Has(ref) && status.Inputs[ref].IsTrueFor(loggingv1.ConditionReady)) {
				good.Insert(ref)
			} else {
				bad.Insert(ref)
			}
		} else {
			if allowed.Has(ref) && status.Outputs[ref].IsTrueFor(loggingv1.ConditionReady) {
				good.Insert(ref)
			} else {
				bad.Insert(ref)
			}
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

func verifyPipelines(spec *loggingv1.ClusterLogForwarderSpec, status *loggingv1.ClusterLogForwarderStatus) {
	// Validate each pipeline and add a status object.
	status.Pipelines = loggingv1.NamedConditions{}
	names := sets.NewString() // Collect pipeline names

	// Known output names, note if "default" is enabled it will already be in the OutputMap()
	outputs := sets.StringKeySet(spec.OutputMap())
	// Known input names, reserved names not in InputMap() we don't expose default inputs.
	inputs := sets.StringKeySet(spec.InputMap())

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

		_, msgIn := verifyRefs("inputs", *status, pipeline.InputRefs, inputs)
		_, msgOut := verifyRefs("outputs", *status, pipeline.OutputRefs, outputs)

		// Pipelines must all be valid for CLF to be considered valid
		// Partially valid pipelines invalidate CLF
		if msgs := append(msgIn, msgOut...); len(msgs) > 0 { // Something wrong
			msg := strings.Join(msgs, ", ")
			status.Pipelines.Set(pipeline.Name, CondInvalid("invalid: %v", msg))
			continue
		} else {
			status.Pipelines.Set(pipeline.Name, condReady) // Ready
		}
	}
}

// verifyInputs and set status.Inputs conditions
func verifyInputs(spec *loggingv1.ClusterLogForwarderSpec, status *loggingv1.ClusterLogForwarderStatus) {
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
		switch {
		case input.Name == "":
			badInput("input must have a name")
		case loggingv1.ReservedInputNames.Has(input.Name):
			badInput("input name %q is reserved", input.Name)
		case len(status.Inputs[input.Name]) > 0:
			badInput("duplicate name: %q", input.Name)
		// Check if inputspec has either application, infrastructure, or audit specs
		case input.Application == nil && input.Infrastructure == nil && input.Audit == nil:
			badInput("inputspec must define one or more of application, infrastructure, or audit")
		default:
			status.Inputs.Set(input.Name, condReady)
		}
	}
}

func verifyOutputs(clusterlogging *loggingv1.ClusterLogging, clfClient client.Client, spec *loggingv1.ClusterLogForwarderSpec, status *loggingv1.ClusterLogForwarderStatus, extras map[string]bool) {
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
		case !verifyOutputSecret(clusterlogging, clfClient, &output, status.Outputs, extras):
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
	if i > 1 {
		return false
	}
	return true
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
			if output.Type == loggingv1.OutputTypeCloudwatch ||
				output.Type == loggingv1.OutputTypeGoogleCloudLogging || output.Type == loggingv1.OutputTypeLoki {
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

func verifyOutputSecret(clusterlogging *loggingv1.ClusterLogging, clfClient client.Client, output *loggingv1.OutputSpec, conds loggingv1.NamedConditions, extras map[string]bool) bool {
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
	if output.Type == loggingv1.OutputTypeElasticsearch && extras[constants.MigrateDefaultOutput] {
		return true
	}
	log.V(3).Info("getting output secret", "output", output.Name, "secret", output.Secret.Name)
	secret, err := getOutputSecret(clusterlogging, clfClient, output.Secret.Name)
	if err != nil {
		return fail(CondMissing("secret %q not found", output.Secret.Name))
	}
	verifySecret := verifySecretKeysForTLS
	if output.Type == loggingv1.OutputTypeCloudwatch {
		verifySecret = verifySecretKeysForCloudwatch
	}
	if !verifySecret(output, conds, secret) {
		return false
	}
	return true
}

func getOutputSecret(clusterlogging *loggingv1.ClusterLogging, clfClient client.Client, secretName string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	namespacedName := types.NamespacedName{Name: secretName, Namespace: clusterlogging.Namespace}

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
		return fail(CondMissing("auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"))
	case !hasKeyID || !hasSecretKey:
		return fail(CondMissing("auth keys: " + constants.AWSAccessKeyID + " and " + constants.AWSSecretAccessKey + " are required"))
	}
	return true
}

func readClusterName(clfClient client.Client) (string, error) {
	infra := configv1.Infrastructure{}
	err := clfClient.Get(context.Background(), client.ObjectKey{Name: constants.ClusterInfrastructureInstance}, &infra)
	if err != nil {
		return "", err
	}

	return infra.Status.InfrastructureName, nil
}
