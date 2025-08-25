package observability

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obscontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils/json"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"time"
)

var _ = Describe("#Reconcile", func() {
	Context("LOG-6758", func() {
		It("should pass validation when the spec is valid", func() {
			clfYaml := `
apiVersion: observability.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: collector
  namespace: openshift-logging
spec:
  managementState: Managed
  outputs:
    - lokiStack:
        authentication:
          token:
            from: serviceAccount
        target:
          name: logging-loki
          namespace: openshift-logging
      name: default-lokistack
      tls:
        ca:
          configMapName: openshift-service-ca.crt
          key: service-ca.crt
      type: lokiStack
  pipelines:
    - inputRefs:
        - audit
      name: syslog
      outputRefs:
        - default-lokistack
    - inputRefs:
        - infrastructure
        - application
      name: logging-loki
      outputRefs:
        - default-lokistack
    - inputRefs:
        - application
      name: container-logs
      outputRefs:
        - default-lokistack
  serviceAccount:
    name: collector
`
			clf := obs.ClusterLogForwarder{}
			test.MustUnmarshal(clfYaml, &clf)
			sub := v1.Subject{Kind: "ServiceAccount", Name: clf.Spec.ServiceAccount.Name, Namespace: clf.Namespace}
			cm := runtime.NewConfigMap(clf.Namespace, "openshift-service-ca.crt", map[string]string{
				"service-ca.crt": "--service-ca--",
			})
			updateStatus := func(obj client.Object) error {
				aCLF, ok := obj.(*obs.ClusterLogForwarder)
				if !ok {
					return nil
				}
				clf.Status = aCLF.Status
				return nil
			}
			fakeClient := fake.NewClientBuilder().WithRuntimeObjects(
				&clf,
				runtime.NewClusterRoleBinding("collect-app-logs", v1.RoleRef{Name: "collect-application-logs"}, sub),
				runtime.NewClusterRoleBinding("collect-infra-logs", v1.RoleRef{Name: "collect-infrastructure-logs"}, sub),
				runtime.NewClusterRoleBinding("collect-audit-logs", v1.RoleRef{Name: "collect-audit-logs"}, sub),
				runtime.NewServiceAccount(clf.Namespace, clf.Spec.ServiceAccount.Name),
				auth.NewSCC(),
				cm,
			).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						sar, ok := obj.(*authv1.SubjectAccessReview)
						if !ok {
							return nil
						}
						sar.Status.Allowed = true
						return nil
					},
					SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
						return updateStatus(obj)
					},
					SubResourcePatch: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
						return updateStatus(obj)
					},
				}).
				Build()
			fContext := obscontext.ForwarderContext{
				Client:    fakeClient,
				Reader:    fakeClient,
				Forwarder: &clf,
				ConfigMaps: map[string]*corev1.ConfigMap{
					cm.Name: cm,
				},
			}
			r := ClusterLogForwarderReconciler{
				PollInterval: 500 * time.Millisecond,
				TimeOut:      1 * time.Second,
				NewForwarderContext: func() obscontext.ForwarderContext {
					return fContext
				},
			}

			_, err := r.Reconcile(context.TODO(), ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: clf.Namespace, Name: clf.Name},
			})
			Expect(err).To(Succeed())
			Expect(fContext.Forwarder.Status.Conditions).To(HaveCondition(obs.ConditionTypeReady, true, "", ""), json.MustMarshal(fContext.Forwarder.Status.Conditions))
		})
	})

})
