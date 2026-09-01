package admission

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
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

		testEnv = &envtest.Environment{Scheme: scheme, ControlPlaneStartTimeout: time.Minute}

		var err error
		cfg, err = testEnv.Start()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())

		adminClient, err = client.New(cfg, client.Options{Scheme: scheme})
		Expect(err).ToNot(HaveOccurred())
		restricted = clientAsUser(cfg, scheme, restrictedUser)

		for _, ns := range []string{podNS, operatorNS} {
			Expect(adminClient.Create(ctx, internalruntime.NewNamespace(ns))).To(Succeed())
		}

		grantWorkloadRBAC(ctx, adminClient,
			restrictedUser,
			kubeSystemDaemonSetControllerUser,
			operatorServiceAccountUser(operatorNS))

		cm := internalruntime.NewConfigMap(operatorNS, ProtectedSAConfigMapName, nil)
		setCreatorKeys(cm.Data, operatorNS)
		cm.Data[protectedSAKeyPrefix+podNS+"_"+collectorSA] = ""
		Expect(adminClient.Create(ctx, cm)).To(Succeed())

		installPolicyAndBinding(ctx, adminClient, protectedSAPodsPolicy, protectedSAPodsBinding, operatorNS)
		installPolicyAndBinding(ctx, adminClient, protectedSAWorkloadsPolicy, protectedSAWorkloadsBinding, operatorNS)

		Eventually(func(g Gomega) {
			err := restricted.Create(ctx, newTestPod(podNS, uniqueName("canary"), collectorSA))
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("protected ServiceAccount"))
		}, 90*time.Second, time.Second).Should(Succeed(), "protected-SA Pod policy never became active")
	})

	AfterAll(func() {
		if testEnv != nil {
			Expect(testEnv.Stop()).To(Succeed())
		}
	})

	It("denies a Pod referencing the protected SA created by a restricted user", func() {
		err := restricted.Create(ctx, newTestPod(podNS, uniqueName("evil-pod"), collectorSA))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("protected ServiceAccount"))
	})

	It("denies a Deployment referencing the protected SA created by a restricted user", func() {
		err := restricted.Create(ctx, newTestDeployment(podNS, uniqueName("evil-deploy"), collectorSA))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("protected ServiceAccount"))
	})

	It("allows a Pod referencing an unprotected SA", func() {
		Expect(restricted.Create(ctx, newTestPod(podNS, uniqueName("plain-pod"), unprotectedSA))).To(Succeed())
	})

	It("allows a Pod referencing the protected SA when created by an allowed controller", func() {
		daemonSetController := clientAsUser(cfg, scheme, kubeSystemDaemonSetControllerUser)
		Expect(daemonSetController.Create(ctx, newTestPod(podNS, uniqueName("collector-pod"), collectorSA))).To(Succeed())
	})

	It("allows a Deployment referencing the protected SA when created by the operator", func() {
		operator := clientAsUser(cfg, scheme, operatorServiceAccountUser(operatorNS))
		Expect(operator.Create(ctx, newTestDeployment(podNS, uniqueName("collector-deploy"), collectorSA))).To(Succeed())
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

func grantWorkloadRBAC(ctx context.Context, c client.Client, users ...string) {
	role := internalruntime.NewClusterRole("workload-creator",
		internalruntime.NewPolicyRules(
			internalruntime.NewPolicyRule(
				[]string{"", "apps"},
				[]string{"pods", "deployments"},
				nil,
				[]string{"create", "get", "list", "delete"},
			),
		)...,
	)
	Expect(c.Create(ctx, role)).To(Succeed())

	subjects := make([]rbacv1.Subject, 0, len(users))
	for _, u := range users {
		subjects = append(subjects, rbacv1.Subject{APIGroup: rbacv1.GroupName, Kind: "User", Name: u})
	}
	crb := internalruntime.NewClusterRoleBinding("workload-creator-binding",
		rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: role.Name},
		subjects...,
	)
	Expect(c.Create(ctx, crb)).To(Succeed())
}

func installPolicyAndBinding(ctx context.Context, c client.Client, policy *admissionregistrationv1.ValidatingAdmissionPolicy, binding *admissionregistrationv1.ValidatingAdmissionPolicyBinding, operatorNS string) {
	Expect(c.Create(ctx, policy.DeepCopy())).To(Succeed())
	b := binding.DeepCopy()
	if b.Spec.ParamRef != nil {
		b.Spec.ParamRef.Namespace = operatorNS
	}
	Expect(c.Create(ctx, b)).To(Succeed())
}

func newTestPod(namespace, name, sa string) *corev1.Pod {
	pod := internalruntime.NewPod(namespace, name,
		*internalruntime.NewContainer("c", "registry.redhat.io/ubi9/ubi-minimal:latest", corev1.PullIfNotPresent, nil),
	)
	pod.Spec.ServiceAccountName = sa
	internalruntime.SetCommonLabels(pod, constants.VectorName, name, constants.CollectorName)
	return pod
}

func newTestDeployment(namespace, name, sa string) *appsv1.Deployment {
	labels := map[string]string{"app": name}
	replicas := int32(1)
	deploy := internalruntime.NewDeployment(namespace, name)
	deploy.Spec = appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{MatchLabels: labels},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: labels},
			Spec: corev1.PodSpec{
				ServiceAccountName: sa,
				Containers: []corev1.Container{
					*internalruntime.NewContainer("c", "registry.redhat.io/ubi9/ubi-minimal:latest", corev1.PullIfNotPresent, nil),
				},
			},
		},
	}
	return deploy
}
