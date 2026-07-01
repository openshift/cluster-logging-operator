package reconcile_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("reconciling RoleBinding", func() {

	var (
		namespace = "test-namespace"
		name      = "test-rolebinding"
		oldRef    = rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "old-role"}
		newRef    = rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "new-role"}
		subject   = rbacv1.Subject{Kind: "ServiceAccount", Name: "test-sa", Namespace: namespace}
		ownerRef  = metav1.OwnerReference{APIVersion: "v1", Kind: "ConfigMap", Name: "owner", UID: "12345"}
	)

	newClient := func(objs ...client.Object) client.Client {
		globalScheme := k8sruntime.NewScheme()
		Expect(scheme.AddToScheme(globalScheme)).To(Succeed())
		return fake.NewClientBuilder().WithScheme(globalScheme).WithObjects(objs...).Build()
	}

	getRoleBinding := func(k8sClient client.Client) *rbacv1.RoleBinding {
		result := &rbacv1.RoleBinding{}
		key := client.ObjectKey{Namespace: namespace, Name: name}
		Expect(k8sClient.Get(context.TODO(), key, result)).To(Succeed())
		return result
	}

	It("should create the RoleBinding when it does not exist", func() {
		k8sClient := newClient()
		desired := runtime.NewRoleBinding(namespace, name, newRef, subject)
		desired.OwnerReferences = []metav1.OwnerReference{ownerRef}

		Expect(reconcile.RoleBinding(k8sClient, desired)).To(Succeed())

		result := getRoleBinding(k8sClient)
		Expect(result.RoleRef).To(Equal(newRef))
		Expect(result.Subjects).To(Equal([]rbacv1.Subject{subject}))
		Expect(result.OwnerReferences).To(Equal([]metav1.OwnerReference{ownerRef}))
	})

	It("should update Subjects and OwnerReferences when roleRef is unchanged", func() {
		existing := runtime.NewRoleBinding(namespace, name, newRef,
			rbacv1.Subject{Kind: "ServiceAccount", Name: "old-sa", Namespace: namespace})
		k8sClient := newClient(existing)

		desired := runtime.NewRoleBinding(namespace, name, newRef, subject)
		desired.OwnerReferences = []metav1.OwnerReference{ownerRef}

		Expect(reconcile.RoleBinding(k8sClient, desired)).To(Succeed())

		result := getRoleBinding(k8sClient)
		Expect(result.RoleRef).To(Equal(newRef))
		Expect(result.Subjects).To(Equal([]rbacv1.Subject{subject}))
		Expect(result.OwnerReferences).To(Equal([]metav1.OwnerReference{ownerRef}))
	})

	It("should delete and recreate the RoleBinding when roleRef changes", func() {
		existing := runtime.NewRoleBinding(namespace, name, oldRef,
			rbacv1.Subject{Kind: "ServiceAccount", Name: "old-sa", Namespace: namespace})
		k8sClient := newClient(existing)

		desired := runtime.NewRoleBinding(namespace, name, newRef, subject)
		desired.OwnerReferences = []metav1.OwnerReference{ownerRef}

		Expect(reconcile.RoleBinding(k8sClient, desired)).To(Succeed())

		result := getRoleBinding(k8sClient)
		Expect(result.RoleRef).To(Equal(newRef))
		Expect(result.Subjects).To(Equal([]rbacv1.Subject{subject}))
		Expect(result.OwnerReferences).To(Equal([]metav1.OwnerReference{ownerRef}))
	})
})
