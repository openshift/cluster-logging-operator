package console

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testClusterVersion = "v4.14.0"
)

var _ = Describe("Cluster Observability Operator compatibility", func() {
	var (
		clWithLoki = &loggingv1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLoggingSpec{
				LogStore: &loggingv1.LogStoreSpec{
					Type: loggingv1.LogStoreTypeLokiStack,
					LokiStack: loggingv1.LokiStackStoreSpec{
						Name: "testing-loki",
					},
				},
			},
		}

		clWithoutLoki = &loggingv1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
		}
	)

	It("should deploy normally when no COO is present", func() {
		c := fake.NewFakeClient()
		err := ReconcilePlugin(c, clWithLoki, clWithLoki, testClusterVersion)
		Expect(err).To(BeNil())

		Expect(checkResourceExists(ctx, c, &consolev1alpha1.ConsolePlugin{}, "", Name)).To(BeTrue())
		Expect(checkResourceExists(ctx, c, &appsv1.Deployment{}, constants.OpenshiftNS, Name)).To(BeTrue())
		Expect(checkResourceExists(ctx, c, &corev1.Service{}, constants.OpenshiftNS, Name)).To(BeTrue())
		Expect(checkResourceExists(ctx, c, &corev1.ConfigMap{}, constants.OpenshiftNS, Name)).To(BeTrue())
	})

	It("should remove the existing resources when the COO manages the ConsolePlugin", func() {
		cooConsolePlugin := &consolev1alpha1.ConsolePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: Name,
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: ptr.To(true),
						APIVersion: uiPluginAPIVersion,
						Kind:       uiPluginKind,
						Name:       "ui-logging",
					},
				},
			},
		}
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: constants.OpenshiftNS,
			},
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: constants.OpenshiftNS,
			},
		}
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: constants.OpenshiftNS,
			},
		}

		c := fake.NewFakeClient(cooConsolePlugin, deployment, service, configMap)
		err := ReconcilePlugin(c, clWithLoki, clWithLoki, testClusterVersion)
		Expect(err).To(BeNil())

		Expect(checkResourceExists(ctx, c, &consolev1alpha1.ConsolePlugin{}, "", Name)).To(BeTrue())
		Expect(checkResourceExists(ctx, c, &appsv1.Deployment{}, constants.OpenshiftNS, Name)).To(BeFalse())
		Expect(checkResourceExists(ctx, c, &corev1.Service{}, constants.OpenshiftNS, Name)).To(BeFalse())
		Expect(checkResourceExists(ctx, c, &corev1.ConfigMap{}, constants.OpenshiftNS, Name)).To(BeFalse())
	})

	It("should not remove ConsolePlugin if managed by COO", func() {
		cooConsolePlugin := &consolev1alpha1.ConsolePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: Name,
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: ptr.To(true),
						APIVersion: uiPluginAPIVersion,
						Kind:       uiPluginKind,
						Name:       "ui-logging",
					},
				},
			},
		}
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: constants.OpenshiftNS,
			},
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: constants.OpenshiftNS,
			},
		}
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: constants.OpenshiftNS,
			},
		}

		c := fake.NewFakeClient(cooConsolePlugin, deployment, service, configMap)
		err := ReconcilePlugin(c, clWithoutLoki, clWithoutLoki, testClusterVersion)
		Expect(err).To(BeNil())

		Expect(checkResourceExists(ctx, c, &consolev1alpha1.ConsolePlugin{}, "", Name)).To(BeTrue())
		Expect(checkResourceExists(ctx, c, &appsv1.Deployment{}, constants.OpenshiftNS, Name)).To(BeFalse())
		Expect(checkResourceExists(ctx, c, &corev1.Service{}, constants.OpenshiftNS, Name)).To(BeFalse())
		Expect(checkResourceExists(ctx, c, &corev1.ConfigMap{}, constants.OpenshiftNS, Name)).To(BeFalse())
	})
})

func checkResourceExists(ctx context.Context, c client.Client, obj client.Object, namespace, name string) bool {
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	err := c.Get(ctx, key, obj)
	return !apierrors.IsNotFound(err)
}
