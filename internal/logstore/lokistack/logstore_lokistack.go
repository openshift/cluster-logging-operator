package lokistack

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ViaQ/logerr/v2/kverrors"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	lokiStackFinalizer = "logging.openshift.io/lokistack-rbac"

	lokiStackWriterClusterRoleName        = "logging-collector-logs-writer"
	lokiStackWriterClusterRoleBindingName = "logging-collector-logs-writer"

	lokiStackAppReaderClusterRoleName        = "logging-application-logs-reader"
	lokiStackAppReaderClusterRoleBindingName = "logging-all-authenticated-application-logs-reader"
)

var (
	DefaultLokiOuputNames sets.String
)

func init() {
	DefaultLokiOuputNames = sets.NewString()
	for _, input := range loggingv1.ReservedInputNames.List() {
		DefaultLokiOuputNames.Insert(FormatOutputNameFromInput(input))
	}
}

func ReconcileLokiStackLogStore(k8sClient client.Client, deletionTimestamp *v1.Time, appendFinalizer func(identifier string) error) error {
	if deletionTimestamp != nil {
		// Skip creation if deleting
		return nil
	}

	if err := appendFinalizer(lokiStackFinalizer); err != nil {
		return kverrors.Wrap(err, "Failed to set finalizer for LokiStack RBAC rules.")
	}

	if err := reconcile.ClusterRole(k8sClient, lokiStackWriterClusterRoleName, newLokiStackWriterClusterRole); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRole for LokiStack collector.")
	}

	if err := reconcile.ClusterRoleBinding(k8sClient, lokiStackWriterClusterRoleBindingName, newLokiStackWriterClusterRoleBinding); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRoleBinding for LokiStack collector.")
	}

	if err := reconcile.ClusterRole(k8sClient, lokiStackAppReaderClusterRoleName, newLokiStackAppReaderClusterRole); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRole for reading application logs.")
	}

	if err := reconcile.ClusterRoleBinding(k8sClient, lokiStackAppReaderClusterRoleBindingName, newLokiStackAppReaderClusterRoleBinding); err != nil {
		return kverrors.Wrap(err, "Failed to create or update ClusterRoleBinding for reading application logs.")
	}

	return nil
}

func RemoveRbac(k8sClient client.Client, removeFinalizer func(string) error) error {
	if err := reconcile.DeleteClusterRoleBinding(k8sClient, lokiStackAppReaderClusterRoleBindingName); err != nil {
		return err
	}

	if err := reconcile.DeleteClusterRole(k8sClient, lokiStackAppReaderClusterRoleName); err != nil {
		return err
	}

	if err := reconcile.DeleteClusterRoleBinding(k8sClient, lokiStackWriterClusterRoleBindingName); err != nil {
		return err
	}

	if err := reconcile.DeleteClusterRole(k8sClient, lokiStackWriterClusterRoleName); err != nil {
		return err
	}

	if err := removeFinalizer(lokiStackFinalizer); err != nil {
		return err
	}

	return nil
}

func newLokiStackWriterClusterRole() *rbacv1.ClusterRole {

	return runtime.NewClusterRole(lokiStackWriterClusterRoleName,
		rbacv1.PolicyRule{
			APIGroups: []string{
				"loki.grafana.com",
			},
			Resources: []string{
				"application",
				"audit",
				"infrastructure",
			},
			ResourceNames: []string{
				"logs",
			},
			Verbs: []string{
				"create",
			},
		},
	)
}

func newLokiStackWriterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return runtime.NewClusterRoleBinding(
		lokiStackWriterClusterRoleBindingName,
		rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     lokiStackWriterClusterRoleName,
		}, rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      "logcollector",
			Namespace: "openshift-logging",
		},
	)
}

func newLokiStackAppReaderClusterRole() *rbacv1.ClusterRole {
	return runtime.NewClusterRole(
		lokiStackAppReaderClusterRoleName,
		rbacv1.PolicyRule{
			APIGroups: []string{
				"loki.grafana.com",
			},
			Resources: []string{
				"application",
			},
			ResourceNames: []string{
				"logs",
			},
			Verbs: []string{
				"get",
			},
		},
	)
}

func newLokiStackAppReaderClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return runtime.NewClusterRoleBinding(
		lokiStackAppReaderClusterRoleBindingName,
		rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     lokiStackAppReaderClusterRoleName,
		},
		rbacv1.Subject{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Group",
			Name:     "system:authenticated",
		},
	)
}

func ProcessForwarderPipelines(logStore *loggingv1.LogStoreSpec, namespace string, spec loggingv1.ClusterLogForwarderSpec, extras map[string]bool) ([]loggingv1.OutputSpec, []loggingv1.PipelineSpec, map[string]bool) {
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
			if input.Infrastructure != nil {
				return loggingv1.InputNameInfrastructure
			}
			if input.Audit != nil {
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
