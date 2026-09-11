package admission

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	internalreconcile "github.com/openshift/cluster-logging-operator/internal/reconcile"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	runtimeobs "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProtectedSAConfigMapName = "clo-protected-serviceaccounts"

	ProtectedSAPodsPolicyName       = "clo-protected-sa-pods"
	ProtectedSAPodsBindingName      = "clo-protected-sa-pods-binding"
	ProtectedSAWorkloadsPolicyName  = "clo-protected-sa-workloads"
	ProtectedSAWorkloadsBindingName = "clo-protected-sa-workloads-binding"

	protectedSAKeyPrefix           = "sa_"
	protectedSAPodCreatorsKey      = "podCreators"
	protectedSAWorkloadCreatorsKey = "workloadCreators"
)

// podControllers lists kube-system controllers that create Pods from
// higher-level workload resources matched by the protected-SA VAPs.
var podControllers = []string{
	"system:serviceaccount:kube-system:daemon-set-controller",
	"system:serviceaccount:kube-system:replicaset-controller",
	"system:serviceaccount:kube-system:statefulset-controller",
	"system:serviceaccount:kube-system:job-controller",
	"system:serviceaccount:kube-system:replication-controller",
}

// workloadControllers lists kube-system controllers that create intermediate
// workload resources (e.g. Deployment → ReplicaSet, CronJob → Job).
var workloadControllers = []string{
	"system:serviceaccount:kube-system:deployment-controller",
	"system:serviceaccount:kube-system:cronjob-controller",
}

//go:embed protected-sa-pods.yaml
var protectedSAPodsPolicyYAML string

//go:embed protected-sa-pods-binding.yaml
var protectedSAPodsBindingYAML string

//go:embed protected-sa-workloads.yaml
var protectedSAWorkloadsPolicyYAML string

//go:embed protected-sa-workloads-binding.yaml
var protectedSAWorkloadsBindingYAML string

var (
	protectedSAPodsPolicy       *admissionregistrationv1.ValidatingAdmissionPolicy
	protectedSAPodsBinding      *admissionregistrationv1.ValidatingAdmissionPolicyBinding
	protectedSAWorkloadsPolicy  *admissionregistrationv1.ValidatingAdmissionPolicy
	protectedSAWorkloadsBinding *admissionregistrationv1.ValidatingAdmissionPolicyBinding
)

func init() {
	protectedSAPodsPolicy = internalruntime.Decode(protectedSAPodsPolicyYAML).(*admissionregistrationv1.ValidatingAdmissionPolicy)
	protectedSAPodsBinding = internalruntime.Decode(protectedSAPodsBindingYAML).(*admissionregistrationv1.ValidatingAdmissionPolicyBinding)
	protectedSAWorkloadsPolicy = internalruntime.Decode(protectedSAWorkloadsPolicyYAML).(*admissionregistrationv1.ValidatingAdmissionPolicy)
	protectedSAWorkloadsBinding = internalruntime.Decode(protectedSAWorkloadsBindingYAML).(*admissionregistrationv1.ValidatingAdmissionPolicyBinding)
}

// OperatorNamespace returns the namespace the operator pod runs in.
// It reads the projected ServiceAccount namespace (set by the kubelet, always
// present in-cluster) so the result is independent of WATCH_NAMESPACE /
// olm.targetNamespaces, which may list namespaces the operator watches rather
// than the one it is deployed in.
func OperatorNamespace() string {
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); ns != "" {
			return ns
		}
	}
	if ns := os.Getenv("WATCH_NAMESPACE"); ns != "" {
		return strings.Split(ns, ",")[0]
	}
	return constants.OpenshiftNS
}

func operatorServiceAccountUser(operatorNS string) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", operatorNS, constants.ClusterLoggingOperator)
}

// ReconcileProtectedSAPolicies ensures the two ValidatingAdmissionPolicies and
// their bindings exist, and that the param ConfigMap exists with the allowed
// creator identities populated.
func ReconcileProtectedSAPolicies(ctx context.Context, k8sClient client.Client, operatorNS string) error {
	if err := ensureProtectedSAConfigMap(k8sClient, operatorNS); err != nil {
		return err
	}
	if err := SyncProtectedServiceAccounts(ctx, k8sClient, operatorNS); err != nil {
		log.V(1).Info("initial protected ServiceAccount sync failed; will resync on CLF events", "error", err)
	}

	for _, p := range []struct {
		policy  *admissionregistrationv1.ValidatingAdmissionPolicy
		binding *admissionregistrationv1.ValidatingAdmissionPolicyBinding
	}{
		{protectedSAPodsPolicy, protectedSAPodsBinding},
		{protectedSAWorkloadsPolicy, protectedSAWorkloadsBinding},
	} {
		if err := internalreconcile.ValidatingAdmissionPolicy(ctx, k8sClient, p.policy); err != nil {
			if internalreconcile.IsUnsupportedAdmissionPolicyAPI(err) {
				log.Info("ValidatingAdmissionPolicy API is unavailable; skipping", "name", p.policy.Name)
				return nil
			}
			return err
		}
		binding := p.binding.DeepCopy()
		if binding.Spec.ParamRef != nil {
			binding.Spec.ParamRef.Namespace = operatorNS
		}
		if err := internalreconcile.ValidatingAdmissionPolicyBinding(ctx, k8sClient, binding); err != nil {
			if internalreconcile.IsUnsupportedAdmissionPolicyAPI(err) {
				log.Info("ValidatingAdmissionPolicyBinding API is unavailable; skipping", "name", binding.Name)
				return nil
			}
			return err
		}
	}
	return nil
}

func ensureProtectedSAConfigMap(k8sClient client.Client, operatorNS string) error {
	cm := internalruntime.NewConfigMap(operatorNS, ProtectedSAConfigMapName, nil)
	internalruntime.SetCommonLabels(cm, constants.ClusterLogging, ProtectedSAConfigMapName, "admission")
	cm.Data = map[string]string{}
	setCreatorKeys(cm.Data, operatorNS)
	if err := internalreconcile.Configmap(k8sClient, k8sClient, cm, comparators.CompareLabels); err != nil {
		return fmt.Errorf("ensure protected SA ConfigMap %s/%s: %w", operatorNS, ProtectedSAConfigMapName, err)
	}
	return nil
}

func setCreatorKeys(data map[string]string, operatorNS string) {
	data[protectedSAPodCreatorsKey] = strings.Join(podControllers, ",")
	data[protectedSAWorkloadCreatorsKey] = strings.Join(
		append([]string{operatorServiceAccountUser(operatorNS)}, workloadControllers...), ",")
}

// SyncProtectedServiceAccounts rebuilds the param ConfigMap's protected-SA
// membership from the full set of ClusterLogForwarders.
func SyncProtectedServiceAccounts(ctx context.Context, k8sClient client.Client, operatorNS string) error {
	refs, err := runtimeobs.CollectorServiceAccounts(ctx, k8sClient)
	if err != nil {
		return err
	}

	data := map[string]string{}
	for _, ref := range refs {
		data[protectedSAKeyPrefix+ref.Namespace+"_"+ref.Name] = ""
	}
	setCreatorKeys(data, operatorNS)

	cm := internalruntime.NewConfigMap(operatorNS, ProtectedSAConfigMapName, nil)
	internalruntime.SetCommonLabels(cm, constants.ClusterLogging, ProtectedSAConfigMapName, "admission")
	cm.Data = data
	if err := internalreconcile.Configmap(k8sClient, k8sClient, cm, comparators.CompareLabels); err != nil {
		return fmt.Errorf("sync protected SA ConfigMap: %w", err)
	}
	log.V(3).Info("synced protected collector ServiceAccounts", "count", len(refs), "serviceAccounts", refs)
	return nil
}
