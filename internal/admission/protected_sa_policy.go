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
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	ProtectedSAConfigMapName = "clo-protected-serviceaccounts"

	ProtectedSAPodsPolicyName       = "clo-protected-sa-pods"
	ProtectedSAPodsBindingName      = "clo-protected-sa-pods-binding"
	ProtectedSAWorkloadsPolicyName  = "clo-protected-sa-workloads"
	ProtectedSAWorkloadsBindingName = "clo-protected-sa-workloads-binding"

	protectedSAKeyPrefix               = "sa_"
	protectedSAPodCreatorsKey          = "podCreators"
	protectedSAWorkloadCreatorsKey     = "workloadCreators"
	kubeSystemDaemonSetControllerUser  = "system:serviceaccount:kube-system:daemon-set-controller"
	kubeSystemReplicaSetControllerUser = "system:serviceaccount:kube-system:replicaset-controller"
	kubeSystemDeploymentControllerUser = "system:serviceaccount:kube-system:deployment-controller"
)

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

// OperatorNamespace returns the namespace the operator runs in.
func OperatorNamespace() string {
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
	if err := ensureProtectedSAConfigMap(ctx, k8sClient, operatorNS); err != nil {
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
		if err := reconcileValidatingAdmissionPolicy(ctx, k8sClient, p.policy); err != nil {
			return err
		}
		binding := p.binding.DeepCopy()
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
	cm := internalruntime.NewConfigMap(operatorNS, ProtectedSAConfigMapName, nil)
	internalruntime.SetCommonLabels(cm, constants.ClusterLogging, ProtectedSAConfigMapName, "admission")
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
// membership from the full set of ClusterLogForwarders.
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

	cm := internalruntime.NewConfigMap(operatorNS, ProtectedSAConfigMapName, nil)
	internalruntime.SetCommonLabels(cm, constants.ClusterLogging, ProtectedSAConfigMapName, "admission")
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
