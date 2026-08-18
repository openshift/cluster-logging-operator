package reconcile_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ClusterRoleBinding reconcile", func() {
	It("should recreate ClusterRoleBinding when roleRef changes", func() {
		c := fake.NewFakeClient()
		name := "crb-under-test"

		makeCRB := func(roleName string) func() *rbacv1.ClusterRoleBinding {
			return func() *rbacv1.ClusterRoleBinding {
				return runtime.NewClusterRoleBinding(name,
					rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: roleName},
					rbacv1.Subject{Kind: "ServiceAccount", Name: "sa", Namespace: "openshift-logging"},
				)
			}
		}

		err := reconcile.ClusterRoleBinding(c, name, makeCRB("metadata-reader"))
		Expect(err).To(Succeed())

		// roleRef is immutable in Kubernetes; a plain Update here would fail.
		err = reconcile.ClusterRoleBinding(c, name, makeCRB("system:auth-delegator"))
		Expect(err).To(Succeed())

		got := &rbacv1.ClusterRoleBinding{}
		Expect(c.Get(context.TODO(), client.ObjectKey{Name: name}, got)).To(Succeed())
		Expect(got.RoleRef.Name).To(Equal("system:auth-delegator"))
	})
})
