package admission

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"sort"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Protected collector ServiceAccount admission.
//
// These policies enforce that a Pod (or a pod-templating workload) referencing a
// protected collector ServiceAccount may only be created by the Cluster Logging
// Operator (workloads) or the built-in controllers that propagate from an
// operator-created workload (pods). The trust anchor is request.userInfo, the
// authenticated identity set by the API server; it cannot be spoofed by copying
// collector Pod metadata (labels/annotations/names are all reproducible).
//
// The set of protected ServiceAccounts and the allowed creator identities are
// supplied to CEL through a param ConfigMap that the operator keeps in sync with
// the set of ClusterLogForwarders (see SyncProtectedServiceAccounts).
const (
	ProtectedSAConfigMapName = "clo-protected-serviceaccounts"

	ProtectedSAPodsPolicyName       = "clo-protected-sa-pods"
	ProtectedSAPodsBindingName      = "clo-protected-sa-pods-binding"
	ProtectedSAWorkloadsPolicyName  = "clo-protected-sa-workloads"
	ProtectedSAWorkloadsBindingName = "clo-protected-sa-workloads-binding"
	// protectedSAKeyPrefix and the '_' separator below build the ConfigMap data
	// key "sa_<namespace>_<name>". ConfigMap keys may not contain '/', and both
	// namespaces (DNS-1123 label) and ServiceAccount names (DNS-1123 subdomain)
	// forbid '_', so the encoding is valid and collision-free. The same key is
	// reconstructed in CEL (see the saKey variable in the policy manifests).
	protectedSAKeyPrefix               = "sa_"
	protectedSAPodCreatorsKey          = "podCreators"
	protectedSAWorkloadCreatorsKey     = "workloadCreators"
	serviceAccountNamespaceFile        = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	kubeSystemDaemonSetControllerUser  = "system:serviceaccount:kube-system:daemon-set-controller"
	kubeSystemReplicaSetControllerUser = "system:serviceaccount:kube-system:replicaset-controller"
	kubeSystemDeploymentControllerUser = "system:serviceaccount:kube-system:deployment-controller"
)

//go:embed manifests/protected-sa-pods.yaml
var protectedSAPodsPolicyYAML []byte

//go:embed manifests/protected-sa-pods-binding.yaml
var protectedSAPodsBindingYAML []byte

//go:embed manifests/protected-sa-workloads.yaml
var protectedSAWorkloadsPolicyYAML []byte

//go:embed manifests/protected-sa-workloads-binding.yaml
var protectedSAWorkloadsBindingYAML []byte

// OperatorNamespace returns the namespace the operator runs in, used both to
// locate the param ConfigMap and to build the operator ServiceAccount username.
func OperatorNamespace() string {
	if data, err := os.ReadFile(serviceAccountNamespaceFile); err == nil {
		if ns := strings.TrimSpace(string(data)); ns != "" {
			return ns
		}
	}
	if ns := strings.TrimSpace(os.Getenv("POD_NAMESPACE")); ns != "" {
		return ns
	}
	return constants.OpenshiftNS
}

func operatorServiceAccountUser(operatorNS string) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", operatorNS, constants.ClusterLoggingOperator)
}

// ReconcileProtectedSAPolicies ensures the two ValidatingAdmissionPolicies and
// their bindings exist, and that the param ConfigMap exists with the allowed
// creator identities populated. Protected SA membership is (re)computed by
// SyncProtectedServiceAccounts.
func ReconcileProtectedSAPolicies(ctx context.Context, k8sClient client.Client, operatorNS string) error {
	if err := ensureProtectedSAConfigMap(ctx, k8sClient, operatorNS); err != nil {
		return err
	}
	if err := SyncProtectedServiceAccounts(ctx, k8sClient, operatorNS); err != nil {
		// A failure to enumerate CLFs must not block policy installation; the
		// controller will resync. Log and continue.
		log.V(1).Info("initial protected ServiceAccount sync failed; will resync on CLF events", "error", err)
	}

	policies := []struct {
		policyYAML  []byte
		bindingYAML []byte
	}{
		{protectedSAPodsPolicyYAML, protectedSAPodsBindingYAML},
		{protectedSAWorkloadsPolicyYAML, protectedSAWorkloadsBindingYAML},
	}
	for _, p := range policies {
		policy, err := decodeValidatingAdmissionPolicy(p.policyYAML)
		if err != nil {
			return err
		}
		if err := reconcileValidatingAdmissionPolicy(ctx, k8sClient, policy); err != nil {
			return err
		}
		binding, err := decodeValidatingAdmissionPolicyBinding(p.bindingYAML)
		if err != nil {
			return err
		}
		// Point the paramRef at the ConfigMap in the operator's own namespace.
		if binding.Spec.ParamRef != nil {
			binding.Spec.ParamRef.Namespace = operatorNS
		}
		if err := reconcileValidatingAdmissionPolicyBinding(ctx, k8sClient, binding); err != nil {
			return err
		}
	}
	return nil
}

func ensureProtectedSAConfigMap(ctx context.Context, k8sClient client.Client, operatorNS string) error {
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: ProtectedSAConfigMapName, Namespace: operatorNS}}
	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, cm, func() error {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		setCreatorKeys(cm.Data, operatorNS)
		return nil
	})
	if err != nil {
		return fmt.Errorf("ensure protected SA ConfigMap %s/%s: %w", operatorNS, ProtectedSAConfigMapName, err)
	}
	return nil
}

func setCreatorKeys(data map[string]string, operatorNS string) {
	data[protectedSAPodCreatorsKey] = strings.Join([]string{
		kubeSystemDaemonSetControllerUser,
		kubeSystemReplicaSetControllerUser,
	}, ",")
	data[protectedSAWorkloadCreatorsKey] = strings.Join([]string{
		operatorServiceAccountUser(operatorNS),
		kubeSystemDeploymentControllerUser,
	}, ",")
}

// SyncProtectedServiceAccounts rebuilds the param ConfigMap's protected-SA
// membership from the full set of ClusterLogForwarders. Rebuilding from a list
// (rather than incremental add/remove) makes deletion handling automatic and
// keeps the ConfigMap self-healing.
func SyncProtectedServiceAccounts(ctx context.Context, k8sClient client.Client, operatorNS string) error {
	clfList := &obsv1.ClusterLogForwarderList{}
	if err := k8sClient.List(ctx, clfList); err != nil {
		return fmt.Errorf("list ClusterLogForwarders: %w", err)
	}

	saKeys := map[string]string{}
	for i := range clfList.Items {
		clf := &clfList.Items[i]
		sa := strings.TrimSpace(clf.Spec.ServiceAccount.Name)
		if sa == "" {
			continue
		}
		saKeys[protectedSAKeyPrefix+clf.Namespace+"_"+sa] = ""
	}

	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: ProtectedSAConfigMapName, Namespace: operatorNS}}
	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, cm, func() error {
		data := map[string]string{}
		for k, v := range saKeys {
			data[k] = v
		}
		setCreatorKeys(data, operatorNS)
		cm.Data = data
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync protected SA ConfigMap: %w", err)
	}
	log.V(3).Info("synced protected collector ServiceAccounts", "count", len(saKeys), "serviceAccounts", sortedKeys(saKeys))
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, strings.TrimPrefix(k, protectedSAKeyPrefix))
	}
	sort.Strings(keys)
	return keys
}
