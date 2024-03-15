package e2e

import (
	"context"
	"fmt"
	clolog "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"time"
)

type AuthorizationBuilder struct {
	tc          *E2ETestFramework
	saNamespace string
	saName      string
	roleNames   []string
}

func (tc *E2ETestFramework) BuildAuthorizationFor(saNamespace, saName string) *AuthorizationBuilder {
	return &AuthorizationBuilder{
		tc:          tc,
		saNamespace: saNamespace,
		saName:      saName,
	}
}

func (b *AuthorizationBuilder) AllowClusterRole(roleName string) *AuthorizationBuilder {
	b.roleNames = append(b.roleNames, roleName)
	return b
}

func (b *AuthorizationBuilder) Create() (sa *corev1.ServiceAccount, err error) {
	sa, err = b.tc.createServiceAccount(b.saNamespace, b.saName)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}

	for _, role := range b.roleNames {
		crb := runtime.NewClusterRoleBinding(fmt.Sprintf("%s-%s-%d%d", sa.Namespace, sa.Name, time.Now().Unix(), rand.Intn(100)),
			rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     role,
			},
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		)
		b.tc.AddCleanup(func() error {
			var zerograce int64
			opts := metav1.DeleteOptions{
				GracePeriodSeconds: &zerograce,
			}
			return b.tc.KubeClient.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.GetName(), opts)
		})
		opts := metav1.CreateOptions{}
		clolog.V(3).Info("Creating", "clusterrolebinding", crb.Name, "namespace", sa.Namespace, "name", sa.Name)
		if _, err = b.tc.KubeClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb, opts); err != nil && !errors.IsAlreadyExists(err) {
			return nil, err
		}
	}

	return sa, nil
}

func (tc *E2ETestFramework) createServiceAccount(namespace, name string) (serviceAccount *corev1.ServiceAccount, err error) {
	opts := metav1.CreateOptions{}
	serviceAccount = runtime.NewServiceAccount(namespace, name)
	clolog.V(3).Info("Creating serviceaccount", "serviceaccount", serviceAccount)
	if serviceAccount, err = tc.KubeClient.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), serviceAccount, opts); err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), serviceAccount.Name, opts)
	})
	return serviceAccount, nil
}

func (tc *E2ETestFramework) createRbac(name string) (err error) {
	opts := metav1.CreateOptions{}
	saRole := runtime.NewRole(
		constants.OpenshiftNS,
		name,
		runtime.NewPolicyRules(
			runtime.NewPolicyRule(
				[]string{"security.openshift.io"},
				[]string{"securitycontextconstraints"},
				[]string{"privileged"},
				[]string{"use"},
			),
		)...,
	)
	if _, err = tc.KubeClient.RbacV1().Roles(constants.OpenshiftNS).Create(context.TODO(), saRole, opts); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.RbacV1().Roles(constants.OpenshiftNS).Delete(context.TODO(), name, opts)
	})
	subject := runtime.NewSubject(
		"ServiceAccount",
		name,
	)
	subject.APIGroup = ""
	roleBinding := runtime.NewRoleBinding(
		constants.OpenshiftNS,
		name,
		rbacv1.RoleRef{
			Kind:     "Role",
			Name:     saRole.Name,
			APIGroup: rbacv1.GroupName,
		}, runtime.NewSubjects(
			subject,
		)...,
	)
	if _, err = tc.KubeClient.RbacV1().RoleBindings(constants.OpenshiftNS).Create(context.TODO(), roleBinding, opts); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.RbacV1().RoleBindings(constants.OpenshiftNS).Delete(context.TODO(), name, opts)
	})
	return nil
}
