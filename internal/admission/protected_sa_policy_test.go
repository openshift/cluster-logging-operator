package admission

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Protected collector SA ValidatingAdmissionPolicies", func() {
	const operatorNS = "openshift-logging"

	var (
		fakeClient client.Client
		ctx        = context.Background()
	)

	newClient := func(objs ...client.Object) client.Client {
		scheme := apiruntime.NewScheme()
		Expect(admissionregistrationv1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(obsv1.AddToScheme(scheme)).To(Succeed())
		return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	}

	clf := func(name, ns, sa string) *obsv1.ClusterLogForwarder {
		return &obsv1.ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Spec:       obsv1.ClusterLogForwarderSpec{ServiceAccount: obsv1.ServiceAccount{Name: sa}},
		}
	}

	getCM := func(c client.Client) *corev1.ConfigMap {
		cm := &corev1.ConfigMap{}
		Expect(c.Get(ctx, client.ObjectKey{Name: ProtectedSAConfigMapName, Namespace: operatorNS}, cm)).To(Succeed())
		return cm
	}

	It("reconciles both policies and points the paramRef at the operator namespace", func() {
		fakeClient = newClient()
		Expect(ReconcileProtectedSAPolicies(ctx, fakeClient, operatorNS)).To(Succeed())

		for _, name := range []string{ProtectedSAPodsPolicyName, ProtectedSAWorkloadsPolicyName} {
			policy := &admissionregistrationv1.ValidatingAdmissionPolicy{}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: name}, policy)).To(Succeed())
			Expect(*policy.Spec.FailurePolicy).To(Equal(admissionregistrationv1.Fail))
			Expect(policy.Spec.ParamKind.Kind).To(Equal("ConfigMap"))
		}

		for _, name := range []string{ProtectedSAPodsBindingName, ProtectedSAWorkloadsBindingName} {
			binding := &admissionregistrationv1.ValidatingAdmissionPolicyBinding{}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: name}, binding)).To(Succeed())
			Expect(binding.Spec.ParamRef).ToNot(BeNil())
			Expect(binding.Spec.ParamRef.Namespace).To(Equal(operatorNS))
			Expect(binding.Spec.ParamRef.Name).To(Equal(ProtectedSAConfigMapName))
		}
	})

	It("populates the param ConfigMap with protected SAs and allowed creators", func() {
		fakeClient = newClient(
			clf("app-logging", "team-a", "collector-a"),
			clf("infra-logging", "team-b", "collector-b"),
			clf("no-sa", "team-c", ""), // ignored
		)
		Expect(SyncProtectedServiceAccounts(ctx, fakeClient, operatorNS)).To(Succeed())

		data := getCM(fakeClient).Data
		Expect(data).To(HaveKey("sa_team-a_collector-a"))
		Expect(data).To(HaveKey("sa_team-b_collector-b"))
		Expect(data).ToNot(HaveKey("sa_team-c_")) // empty SA name skipped

		Expect(strings.Split(data[protectedSAPodCreatorsKey], ",")).To(ConsistOf(
			kubeSystemDaemonSetControllerUser,
			kubeSystemReplicaSetControllerUser,
			kubeSystemStatefulSetControllerUser,
			kubeSystemJobControllerUser,
			kubeSystemReplicationControllerUser,
		))
		Expect(strings.Split(data[protectedSAWorkloadCreatorsKey], ",")).To(ConsistOf(
			"system:serviceaccount:openshift-logging:cluster-logging-operator",
			kubeSystemDeploymentControllerUser,
			kubeSystemCronJobControllerUser,
		))
	})

	It("removes an SA from the ConfigMap when its CLF is deleted (rebuild-from-list)", func() {
		fakeClient = newClient(clf("app-logging", "team-a", "collector-a"))
		Expect(SyncProtectedServiceAccounts(ctx, fakeClient, operatorNS)).To(Succeed())
		Expect(getCM(fakeClient).Data).To(HaveKey("sa_team-a_collector-a"))

		Expect(fakeClient.Delete(ctx, clf("app-logging", "team-a", "collector-a"))).To(Succeed())
		Expect(SyncProtectedServiceAccounts(ctx, fakeClient, operatorNS)).To(Succeed())
		Expect(getCM(fakeClient).Data).ToNot(HaveKey("sa_team-a_collector-a"))
	})
})
