package lokistack

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sort"
	"strings"

	"github.com/ViaQ/logerr/v2/kverrors"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const (
	lokiStackFinalizer = "logging.openshift.io/lokistack-rbac"
)

var (
	DefaultLokiOuputNames sets.String
)

func init() {
	DefaultLokiOuputNames = *sets.NewString()
	for _, input := range loggingv1.ReservedInputNames.List() {
		DefaultLokiOuputNames.Insert(FormatOutputNameFromInput(input))
	}
}

// CheckFinalizer checks if the finalizer used for tracking the cluster-wide RBAC resources
// is attached to the provided object and removes it, if present.
func CheckFinalizer(ctx context.Context, client client.Client, obj client.Object) error {
	if controllerutil.RemoveFinalizer(obj, lokiStackFinalizer) {
		if err := client.Update(ctx, obj); err != nil {
			return kverrors.Wrap(err, "Failed to remove finalizer from ClusterLogging.")
		}
	}

	return nil
}

func ProcessForwarderPipelines(logStore *loggingv1.LogStoreSpec, namespace string, spec loggingv1.ClusterLogForwarderSpec, extras map[string]bool, saTokenSecret string) ([]loggingv1.OutputSpec, []loggingv1.PipelineSpec, map[string]bool) {
	needOutput := make(map[string]bool)
	inPipelines := spec.Pipelines
	pipelines := []loggingv1.PipelineSpec{}

	for _, p := range inPipelines {
		if !slices.Contains(p.OutputRefs, loggingv1.OutputNameDefault) {
			// Skip pipelines that do not reference "default" output
			pipelines = append(pipelines, p)
			continue
		}

		for _, i := range p.InputRefs {
			needOutput[i] = true
		}

		for i, input := range p.InputRefs {
			pOut := p.DeepCopy()
			pOut.InputRefs = []string{input}

			for i, output := range pOut.OutputRefs {
				if output != loggingv1.OutputNameDefault {
					// Leave non-default output names as-is
					continue
				}

				pOut.OutputRefs[i] = FormatOutputNameFromInput(input)
				// For loki we don't want to set 'extras[constants.MigrateDefaultOutput] = true'
				// we want 'default' output to fail per LOG-3437 since we did not create it
			}

			// Can no longer have empty pipeline names
			if pOut.Name == "" {
				pOut.Name = fmt.Sprintf("%s_%d_", "default_loki_pipeline", i)
				// Generate new name for named pipelines as duplicate names are not allowed
			} else if pOut.Name != "" && i > 0 {
				pOut.Name = fmt.Sprintf("%s-%d", pOut.Name, i)
			}

			pipelines = append(pipelines, *pOut)
		}
	}

	outputs := []loggingv1.OutputSpec{}
	if spec.Outputs != nil {
		outputs = spec.Outputs
	}
	// Now create output from each input
	for input := range needOutput {
		tenant := getInputTypeFromName(spec, input)
		outputs = append(outputs, loggingv1.OutputSpec{
			Name: FormatOutputNameFromInput(input),
			Type: loggingv1.OutputTypeLoki,
			URL:  lokiStackURL(logStore, namespace, tenant),
			Secret: &loggingv1.OutputSecretSpec{
				Name: saTokenSecret,
			},
		})
	}

	// Sort outputs, because we have tests depending on the exact generated configuration
	sort.Slice(outputs, func(i, j int) bool {
		return strings.Compare(outputs[i].Name, outputs[j].Name) < 0
	})

	return outputs, pipelines, extras
}

func getInputTypeFromName(spec loggingv1.ClusterLogForwarderSpec, inputName string) string {
	if loggingv1.ReservedInputNames.Has(inputName) {
		// use name as type
		return inputName
	}

	for _, input := range spec.Inputs {
		if input.Name == inputName {
			if input.Application != nil {
				return loggingv1.InputNameApplication
			}
			if input.Infrastructure != nil || loggingv1.IsSyslogReceiver(&input) {
				return loggingv1.InputNameInfrastructure
			}
			if input.Audit != nil || loggingv1.IsAuditHttpReceiver(&input) {
				return loggingv1.InputNameAudit
			}
		}
	}
	log.V(3).Info("unable to get input type from name", "inputName", inputName)
	return ""
}

// lokiStackURL returns the URL of the LokiStack API for a specific tenant.
// Returns an empty string if ClusterLogging is not configured for a LokiStack log store.
func lokiStackURL(logStore *loggingv1.LogStoreSpec, namespace, tenant string) string {
	service := LokiStackGatewayService(logStore)
	if service == "" {
		return ""
	}
	if !loggingv1.ReservedInputNames.Has(tenant) {
		log.V(3).Info("url tenant must be one of our reserved input names", "tenant", tenant)
		return ""
	}
	return fmt.Sprintf("https://%s.%s.svc:8080/api/logs/v1/%s", service, namespace, tenant)
}

// LokiStackGatewayService returns the name of LokiStack gateway service.
// Returns an empty string if ClusterLogging is not configured for a LokiStack log store.
func LokiStackGatewayService(logStore *loggingv1.LogStoreSpec) string {
	if logStore == nil || logStore.LokiStack.Name == "" {
		return ""
	}

	return fmt.Sprintf("%s-gateway-http", logStore.LokiStack.Name)
}

// FormatOutputNameFromInput takes an clf.input and formats the output name for  'default' output
func FormatOutputNameFromInput(inputName string) string {
	switch inputName {
	case loggingv1.InputNameApplication:
		return loggingv1.OutputNameDefault + "-loki-apps"
	case loggingv1.InputNameInfrastructure:
		return loggingv1.OutputNameDefault + "-loki-infra"
	case loggingv1.InputNameAudit:
		return loggingv1.OutputNameDefault + "-loki-audit"
	}

	return loggingv1.OutputNameDefault + "-" + inputName
}
