package admission

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// This suite exercises the protected-SA ValidatingAdmissionPolicies against a
// real kube-apiserver (envtest), so the CEL is actually compiled and evaluated
// -- unlike the fake-client unit tests, which only assert the generated objects.
// It catches CEL/key regressions (e.g. an invalid ConfigMap key) without needing
// a full cluster.
//
// It requires the envtest binaries; when KUBEBUILDER_ASSETS is unset the whole
// suite skips, so `make test-unit` is unaffected. Run it via:
//
//	make test-admission-envtest
var _ = Describe("Protected SA VAP enforcement (envtest)", Ordered, func() {
	const (
		podNS          = "collector-ns"
		operatorNS     = "operator-ns"
		collectorSA    = "log-collector"
		unprotectedSA  = "plain-sa"
		restrictedUser = "system:serviceaccount:collector-ns:restricted-user"
	)

	var (
		testEnv     *envtest.Environment
		cfg         *rest.Config
		scheme      *apiruntime.Scheme
		adminClient client.Client
		restricted  client.Client
		ctx         context.Context
	)

	BeforeAll(func() {
		if os.Getenv("KUBEBUILDER_ASSETS") == "" {
			Skip("KUBEBUILDER_ASSETS not set; run via `make test-admission-envtest`")
		}
		ctx = context.Background()

		scheme = apiruntime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		Expect(rbacv1.AddToScheme(scheme)).To(Succeed())
		Expect(admissionregistrationv1.AddToScheme(scheme)).To(Succeed())

		// Use envtest defaults (RBAC authz; the ValidatingAdmissionPolicy admission
		// plugin is enabled by default on the served apiserver version). The default
		// admin credential is in system:masters, a superuser that can impersonate
		// any identity.
		testEnv = &envtest.Environment{Scheme: scheme, ControlPlaneStartTimeout: time.Minute}

		var err error
		cfg, err = testEnv.Start()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())

		adminClient, err = client.New(cfg, client.Options{Scheme: scheme})
		Expect(err).ToNot(HaveOccurred())
		restricted = clientAsUser(cfg, scheme, restrictedUser)

		for _, ns := range []string{podNS, operatorNS} {
			Expect(adminClient.Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: ns},
			})).To(Succeed())
		}

		// Grant each impersonated identity RBAC to create the workloads, so requests
		// reach the ValidatingAdmissionPolicy stage rather than being stopped earlier
		// by authorization. The VAP -- not RBAC -- is what this suite exercises.
		grantWorkloadRBAC(ctx, adminClient,
			restrictedUser,
			kubeSystemDaemonSetControllerUser,
			operatorServiceAccountUser(operatorNS))

		// Param ConfigMap: creator allow-lists (from the real setCreatorKeys) plus
		// the protected SA key the operator would have written for the CLF.
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: ProtectedSAConfigMapName, Namespace: operatorNS},
			Data:       map[string]string{},
		}
		setCreatorKeys(cm.Data, operatorNS)
		cm.Data[protectedSAKeyPrefix+podNS+"_"+collectorSA] = ""
		Expect(adminClient.Create(ctx, cm)).To(Succeed())

		// Install both policies + bindings from the real embedded manifests.
		installPolicyAndBinding(ctx, adminClient, protectedSAPodsPolicyYAML, protectedSAPodsBindingYAML, operatorNS)
		installPolicyAndBinding(ctx, adminClient, protectedSAWorkloadsPolicyYAML, protectedSAWorkloadsBindingYAML, operatorNS)

		// A freshly-created VAP is not enforced instantly; the apiserver compiles
		// and loads it asynchronously. Poll (with unique names) until a protected
		// Pod is actually denied before running the deterministic specs below.
		Eventually(func(g Gomega) {
			err := restricted.Create(ctx, newPod(podNS, uniqueName("canary"), collectorSA))
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("protected collector ServiceAccount"))
		}, 90*time.Second, time.Second).Should(Succeed(), "protected-SA Pod policy never became active")
	})

	AfterAll(func() {
		if testEnv != nil {
			Expect(testEnv.Stop()).To(Succeed())
		}
	})

	It("denies a Pod referencing the protected SA created by a restricted user", func() {
		err := restricted.Create(ctx, newPod(podNS, uniqueName("evil-pod"), collectorSA))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("protected collector ServiceAccount"))
	})

	It("denies a Deployment referencing the protected SA created by a restricted user", func() {
		err := restricted.Create(ctx, newDeployment(podNS, uniqueName("evil-deploy"), collectorSA))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("protected collector ServiceAccount"))
	})

	It("allows a Pod referencing an unprotected SA", func() {
		Expect(restricted.Create(ctx, newPod(podNS, uniqueName("plain-pod"), unprotectedSA))).To(Succeed())
	})

	It("allows a Pod referencing the protected SA when created by an allowed controller", func() {
		daemonSetController := clientAsUser(cfg, scheme, kubeSystemDaemonSetControllerUser)
		Expect(daemonSetController.Create(ctx, newPod(podNS, uniqueName("collector-pod"), collectorSA))).To(Succeed())
	})

	It("allows a Deployment referencing the protected SA when created by the operator", func() {
		operator := clientAsUser(cfg, scheme, operatorServiceAccountUser(operatorNS))
		Expect(operator.Create(ctx, newDeployment(podNS, uniqueName("collector-deploy"), collectorSA))).To(Succeed())
	})
})

var envtestNameCounter int

func uniqueName(prefix string) string {
	envtestNameCounter++
	return fmt.Sprintf("%s-%d", prefix, envtestNameCounter)
}

func clientAsUser(cfg *rest.Config, scheme *apiruntime.Scheme, username string) client.Client {
	impersonated := rest.CopyConfig(cfg)
	impersonated.Impersonate = rest.ImpersonationConfig{UserName: username}
	c, err := client.New(impersonated, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	return c
}

// grantWorkloadRBAC lets the given impersonated identities create/delete the
// workloads under test, so the VAP (not RBAC) is the sole gate the specs
// observe. The identities are bound explicitly as Users (an impersonated
// UserName is not guaranteed to carry the system:authenticated group).
func grantWorkloadRBAC(ctx context.Context, c client.Client, users ...string) {
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "workload-creator"},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"", "apps"},
			Resources: []string{"pods", "deployments"},
			Verbs:     []string{"create", "get", "list", "delete"},
		}},
	}
	Expect(c.Create(ctx, role)).To(Succeed())

	subjects := make([]rbacv1.Subject, 0, len(users))
	for _, u := range users {
		subjects = append(subjects, rbacv1.Subject{APIGroup: rbacv1.GroupName, Kind: "User", Name: u})
	}
	Expect(c.Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "workload-creator-binding"},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: role.Name},
		Subjects:   subjects,
	})).To(Succeed())
}

func installPolicyAndBinding(ctx context.Context, c client.Client, policyYAML, bindingYAML []byte, operatorNS string) {
	policy, err := decodeValidatingAdmissionPolicy(policyYAML)
	Expect(err).ToNot(HaveOccurred())
	Expect(c.Create(ctx, policy)).To(Succeed())

	binding, err := decodeValidatingAdmissionPolicyBinding(bindingYAML)
	Expect(err).ToNot(HaveOccurred())
	if binding.Spec.ParamRef != nil {
		binding.Spec.ParamRef.Namespace = operatorNS
	}
	Expect(c.Create(ctx, binding)).To(Succeed())
}

// newPod copies the collector's visible metadata to prove that spoofing
// labels/annotations/name does not bypass the policy -- enforcement keys on
// request.userInfo, not on Pod metadata.
func newPod(namespace, name, sa string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "vector",
				"app.kubernetes.io/component":  "collector",
				"app.kubernetes.io/managed-by": "cluster-logging-operator",
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: sa,
			Containers: []corev1.Container{
				{Name: "c", Image: "registry.redhat.io/ubi9/ubi-minimal:latest"},
			},
		},
	}
}

func newDeployment(namespace, name, sa string) *appsv1.Deployment {
	labels := map[string]string{"app": name}
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					ServiceAccountName: sa,
					Containers: []corev1.Container{
						{Name: "c", Image: "registry.redhat.io/ubi9/ubi-minimal:latest"},
					},
				},
			},
		},
	}
}
