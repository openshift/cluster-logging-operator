package clusterlogforwarder

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/outputs"

	log "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/url"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/inputs"
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

	inputs.Verify(clf.Spec.Inputs, status, extras)
	if !status.Inputs.IsAllReady() {
		log.V(3).Info("Input not Ready", "inputs", status.Inputs)
	}
	verifyOutputs(clf.Namespace, k8sClient, &clf.Spec, status, extras)
	if !status.Outputs.IsAllReady() {
		log.V(3).Info("sink not Ready", "outputs", status.Outputs)
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
		status.Conditions.SetCondition(conditions.CondInvalid("invalid clf spec; one or more errors present: %v", unready))
		return errors.NewValidationError("clusterlogforwarder is not ready"), status
	}
	status.Conditions.SetCondition(conditions.CondReady)
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
			status.Pipelines.Set(pipeline.Name, conditions.CondInvalid("pipeline must have a name"))
			names.Insert(pipeline.Name)
			continue
		}

		if names.Has(pipeline.Name) {
			original := pipeline.Name
			pipeline.Name = fmt.Sprintf("pipeline_%v_", i)
			status.Pipelines.Set(pipeline.Name, conditions.CondInvalid("duplicate name %q", original))
			continue
		}
		names.Insert(pipeline.Name)

		// Verify pipeline labels
		if _, err := json.Marshal(pipeline.Labels); err != nil {
			status.Pipelines.Set(pipeline.Name, conditions.CondInvalid("invalid pipeline labels"))
			continue
		}

		// Verify prune filter does not prune `.hostname` for GCL if filter/output in pipeline
		if invMsg := verifyHostNameNotFilteredForGCL(pipeline.OutputRefs, pipeline.FilterRefs, spec.OutputMap(), spec.FilterMap()); len(invMsg) != 0 {
			status.Pipelines.Set(pipeline.Name, conditions.CondInvalid("googleCloudLogging cannot prune `.hostname` field. %q", invMsg))
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
			status.Pipelines.Set(pipeline.Name, conditions.CondReady) // Ready
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
			status.Outputs.Set(output.Name, conditions.CondInvalid(format, args...))
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
			status.Outputs.Set(output.Name, conditions.CondInvalid("output %q: unknown output type %q", output.Name, output.Type))
		case !verifyOutputURL(&output, status.Outputs):
			log.V(3).Info("verifyOutputs failed", "reason", "output URL is invalid", "output URL", output.URL)
		case !verifyOutputSecret(namespace, clfClient, &output, status.Outputs, extras):
			log.V(3).Info("verifyOutputs failed", "reason", "output secret is invalid")
		case output.Type == loggingv1.OutputTypeCloudwatch && output.Cloudwatch == nil:
			log.V(3).Info("verifyOutputs failed", "reason", "Cloudwatch output requires type spec", "output name", output.Name)
			status.Outputs.Set(output.Name, conditions.CondInvalid("output %q: Cloudwatch output requires type spec", output.Name))
		case output.Type == loggingv1.OutputTypeAzureMonitor:
			if output.AzureMonitor == nil {
				log.V(3).Info("verifyOutputs failed", "reason", "Azure Monitor Logs output requires type spec", "output name", output.Name)
				status.Outputs.Set(output.Name, conditions.CondInvalid("output %q: Azure Monitor Logs output requires type spec", output.Name))
			} else {
				valid, con := outputs.VerifyAzureMonitorLog(output.Name, output.AzureMonitor)
				if !valid {
					log.V(3).Info("verifyOutputs failed", "reason", con.Reason, "output name", output.Name)
				}
				status.Outputs.Set(output.Name, con)
			}
		// Check googlecloudlogging specs, must only include one of the following
		case output.Type == loggingv1.OutputTypeGoogleCloudLogging && output.GoogleCloudLogging != nil && !verifyGoogleCloudLogging(output.GoogleCloudLogging):
			log.V(3).Info("verifyOutputs failed", "reason",
				"Exactly one of billingAccountId, folderId, organizationId, or projectId must be set.",
				"output name", output.Name, "output type", output.Type)
			status.Outputs.Set(output.Name,
				conditions.CondInvalid("output %q: Exactly one of billingAccountId, folderId, organizationId, or projectId must be set.",
					output.Name))
		case output.Type == loggingv1.OutputTypeSplunk:
			valid, con := outputs.VerifySplunk(output.Name, output.Splunk)
			if !valid {
				log.V(3).Info("verifyOutputs failed", "reason", con.Reason, "output name", output.Name)
			}
			status.Outputs.Set(output.Name, con)
		case output.HasPolicy() && output.GetMaxRecordsPerSecond() < 0:
			status.Outputs.Set(output.Name, conditions.CondInvalid("output %q: sink cannot have negative limit threshold", output.Name))
		case !outputRefs.Has(output.Name):
			status.Outputs.Set(output.Name, conditions.CondInvalid("output %q: sink not referenced by any pipeline", output.Name))
		default:
			status.Outputs.Set(output.Name, conditions.CondReady)
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

		if valid, msg := outputs.VerifyTuning(output); !valid {
			log.V(3).Info("verify output tuning failed", "output name", output.Name, "message", msg)
			status.Outputs.Set(output.Name, conditions.CondInvalid("output %q: %s", output.Name, msg))
			status.Outputs.Set(output.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
				corev1.ConditionTrue,
				loggingv1.ValidationFailureReason,
				msg))
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
	notSecureUrlMsg := "invalid configuration: secure URL is required with TLS configurations"
	if output.Type == loggingv1.OutputTypeKafka {
		brokerUrls := []string{}
		if output.URL != "" {
			if u, _ := url.Parse(output.URL); output.TLS != nil && !urlhelper.IsTLSScheme(u.Scheme) {
				return fail(conditions.CondInvalid(notSecureUrlMsg))
			}
			brokerUrls = append(brokerUrls, output.URL)
		}
		if output.Kafka != nil { // Add optional extra broker URLs.
			brokerUrls = append(brokerUrls, output.Kafka.Brokers...)
		}
		if len(brokerUrls) == 0 {
			return fail(conditions.CondInvalid("no broker URLs specified"))
		}
		for _, b := range brokerUrls {
			u, err := url.Parse(b)
			if err == nil {
				err = url.CheckAbsolute(u)
			}
			if err != nil {
				return fail(conditions.CondInvalid("invalid URL: %v", err))
			}
		}
	} else {
		if output.URL == "" {
			// Some output types allow a missing URL
			// TODO (alanconway) move output-specific valiation to the output implementation.
			if output.Type == loggingv1.OutputTypeCloudwatch || output.Type == loggingv1.OutputTypeAzureMonitor ||
				output.Type == loggingv1.OutputTypeGoogleCloudLogging {
				return true
			} else {
				return fail(conditions.CondInvalid("URL is required for output type %v", output.Type))
			}
		}
		u, err := url.Parse(output.URL)
		if err == nil {
			err = url.CheckAbsolute(u)
		}
		if err != nil {
			return fail(conditions.CondInvalid("invalid URL: %v", err))
		}
		if output.TLS != nil && !urlhelper.IsTLSScheme(u.Scheme) {
			return fail(conditions.CondInvalid(notSecureUrlMsg))
		}
		if output.Type == loggingv1.OutputTypeSyslog {
			scheme := strings.ToLower(u.Scheme)
			if !(scheme == `tcp` || scheme == `tls` || scheme == `udp`) {
				return fail(conditions.CondInvalid("invalid URL scheme: %v", u.Scheme))
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
		if output.Type == loggingv1.OutputTypeCloudwatch || output.Type == loggingv1.OutputTypeSplunk {
			return fail(conditions.CondMissing("secret must be provided for %s output", output.Type))
		}
		return true
	}

	if output.Secret.Name == "" {
		conds.Set(output.Name, conditions.CondInvalid("secret has empty name"))
		return false
	}
	// Only for ES. If default replaced, the "collector" secret will be created later
	if (output.Type == loggingv1.OutputTypeElasticsearch || output.Type == loggingv1.OutputTypeLoki) && extras[constants.MigrateDefaultOutput] {
		return true
	}
	log.V(3).Info("getting output secret", "output", output.Name, "secret", output.Secret.Name)
	secret, err := getOutputSecret(namespace, clfClient, output.Secret.Name)
	if err != nil {
		return fail(conditions.CondMissing("secret %q not found", output.Secret.Name))
	}

	switch output.Type {
	case loggingv1.OutputTypeCloudwatch:
		if !verifySecretKeysForCloudwatch(output, conds, secret) {
			return false
		}
	case loggingv1.OutputTypeSplunk:
		if !outputs.VerifySecretKeysForSplunk(output, conds, secret) {
			return false
		}
	case loggingv1.OutputTypeAzureMonitor:
		if !outputs.VerifySharedKeysForAzure(output, conds, secret) {
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
		return fail(conditions.CondMissing("cannot have %v without %v", constants.ClientCertKey, constants.ClientPrivateKey))
	case !haveCert && haveKey:
		return fail(conditions.CondMissing("cannot have %v without %v", constants.ClientPrivateKey, constants.ClientCertKey))
	case haveUsername && !havePassword:
		return fail(conditions.CondMissing("cannot have %v without %v", constants.ClientUsername, constants.ClientPassword))
	case !haveUsername && havePassword:
		return fail(conditions.CondMissing("cannot have %v without %v", constants.ClientPassword, constants.ClientUsername))
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
	hasRoleArnKey := common.HasAwsRoleArnKey(secret)
	hasCredentialsKey := common.HasAwsCredentialsKey(secret)
	// TODO: FIXME
	hasValidRoleArn := false //len(cloudwatch.ParseRoleArn(secret)) > 0
	switch {
	case hasValidRoleArn: // Sts secret format is the first check
		return true
	case hasRoleArnKey && !hasValidRoleArn, hasCredentialsKey && !hasValidRoleArn:
		return fail(conditions.CondMissing("auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"))
	case !hasKeyID || !hasSecretKey:
		return fail(conditions.CondMissing("auth keys: " + constants.AWSAccessKeyID + " and " + constants.AWSSecretAccessKey + " are required"))
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

// verifyHostNameNotFilteredForGCL verifies that within a pipeline featuring a GCL sink and prune filters, the `.hostname` field is exempted from pruning.
func verifyHostNameNotFilteredForGCL(outputRefs []string, filterRefs []string, outputs map[string]*loggingv1.OutputSpec, filters map[string]*loggingv1.FilterSpec) []string {
	if len(filterRefs) == 0 {
		return nil
	}

	var errMsgs []string

	for _, out := range outputRefs {
		if outputs[out].Type == loggingv1.OutputTypeGoogleCloudLogging {
			for _, f := range filterRefs {
				filterSpec := filters[f]
				if prunesHostName(*filterSpec) {
					errMsgs = append(errMsgs, fmt.Sprintf("filter: %s prunes the `.hostname` field which is required for output: %s of type googleCloudLogging.", filterSpec.Name, outputs[out].Name))
				}
			}
		}
	}
	return errMsgs
}

// prunesHostName checks if a prune filter prunes the `.hostname` field
func prunesHostName(filter loggingv1.FilterSpec) bool {
	if filter.Type != loggingv1.FilterPrune {
		return false
	}

	hostName := ".hostname"

	inListPrunes := false
	notInListPrunes := false

	if filter.PruneFilterSpec.NotIn != nil {
		found := false
		for _, field := range filter.PruneFilterSpec.NotIn {
			if field == hostName {
				found = true
				break
			}
		}
		if !found {
			inListPrunes = true
		}
	}

	if filter.PruneFilterSpec.In != nil {
		for _, field := range filter.PruneFilterSpec.In {
			if field == hostName {
				notInListPrunes = true
				break
			}
		}
	}

	return inListPrunes || notInListPrunes
}
