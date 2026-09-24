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
	"github.com/openshift/cluster-logging-operator/internal/runtime/clusterlogforwarder"
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
var (
	// workloadControllers lists kube-system controllers that create intermediate
	// workload resources (e.g. Deployment → ReplicaSet, CronJob → Job).
	workloadControllers = []string{
		"system:serviceaccount:kube-system:deployment-controller",
		"system:serviceaccount:kube-system:cronjob-controller",
	}
	// podCreatorsValue is the pre-joined comma-separated list of pod creators.
	podCreatorsValue string
	// workloadCreatorsTemplate is the pre-joined workload controllers, to be
	// prefixed with the operator SA at runtime.
	workloadCreatorsTemplate string
	//go:embed protected-sa-pods.yaml
	protectedSAPodsPolicyYAML string
	//go:embed protected-sa-pods-binding.yaml
	protectedSAPodsBindingYAML string
	//go:embed protected-sa-workloads.yaml
	protectedSAWorkloadsPolicyYAML string

	//go:embed protected-sa-workloads-binding.yaml
	protectedSAWorkloadsBindingYAML string
	protectedSAPodsPolicy           *admissionregistrationv1.ValidatingAdmissionPolicy
	protectedSAPodsBinding          *admissionregistrationv1.ValidatingAdmissionPolicyBinding
	protectedSAWorkloadsPolicy      *admissionregistrationv1.ValidatingAdmissionPolicy
	protectedSAWorkloadsBinding     *admissionregistrationv1.ValidatingAdmissionPolicyBinding
)

func init() {
	protectedSAPodsPolicy = internalruntime.Decode(protectedSAPodsPolicyYAML).(*admissionregistrationv1.ValidatingAdmissionPolicy)
	protectedSAPodsBinding = internalruntime.Decode(protectedSAPodsBindingYAML).(*admissionregistrationv1.ValidatingAdmissionPolicyBinding)
	protectedSAWorkloadsPolicy = internalruntime.Decode(protectedSAWorkloadsPolicyYAML).(*admissionregistrationv1.ValidatingAdmissionPolicy)
	protectedSAWorkloadsBinding = internalruntime.Decode(protectedSAWorkloadsBindingYAML).(*admissionregistrationv1.ValidatingAdmissionPolicyBinding)

	for _, obj := range []internalruntime.Object{
		protectedSAPodsPolicy, protectedSAPodsBinding,
		protectedSAWorkloadsPolicy, protectedSAWorkloadsBinding,
	} {
		internalruntime.SetCommonLabels(obj, constants.ClusterLogging, ProtectedSAConfigMapName, "admission")
	}

	// Pre-compute static creator identity lists.
	podCreatorsValue = strings.Join(podControllers, ",")
	workloadCreatorsTemplate = strings.Join(workloadControllers, ",")
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
			return err
		}
		binding := p.binding.DeepCopy()
		if binding.Spec.ParamRef != nil {
			binding.Spec.ParamRef.Namespace = operatorNS
		}
		if err := internalreconcile.ValidatingAdmissionPolicyBinding(ctx, k8sClient, binding); err != nil {
			return err
		}
	}
	return nil
}

func setCreatorKeys(data map[string]string, operatorNS string) {
	data[protectedSAPodCreatorsKey] = podCreatorsValue
	data[protectedSAWorkloadCreatorsKey] = operatorServiceAccountUser(operatorNS) + "," + workloadCreatorsTemplate
}

// SyncProtectedServiceAccounts rebuilds the param ConfigMap from the full set
// of ClusterLogForwarders. The ConfigMap is created if it does not exist.
func SyncProtectedServiceAccounts(ctx context.Context, k8sClient client.Client, operatorNS string) error {
	refs, err := clusterlogforwarder.ListServiceAccounts(ctx, k8sClient)
	if err != nil {
		return err
	}

	data := map[string]string{}
	for _, ref := range refs {
		// Use ConfigMap data as a set: the VAP CEL expression checks key presence
		// (saKey in params.data), not value, so empty string is sufficient.
		data[ref.String()] = ""
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
