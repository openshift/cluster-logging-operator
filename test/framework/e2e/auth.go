package e2e

import (
	"fmt"
	clolog "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

var (
	roleFmt                              = "collect-%s-logs"
	ClusterRoleCollectApplicationLogs    = fmt.Sprintf(roleFmt, obsv1.InputTypeApplication)
	ClusterRoleCollectInfrastructureLogs = fmt.Sprintf(roleFmt, obsv1.InputTypeInfrastructure)
	ClusterRoleCollectAuditLogs          = fmt.Sprintf(roleFmt, obsv1.InputTypeAudit)
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
		crb := runtime.NewClusterRoleBinding(fmt.Sprintf("%s-%s-%s", role, sa.Namespace, sa.Name),
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
			return b.tc.Test.Delete(crb)
		})
		clolog.V(3).Info("Creating", "clusterrolebinding", crb.Name, "serviceaccount.namespace", sa.Namespace, "serviceaccount.name", sa.Name)
		if err = b.tc.Test.Recreate(crb); err != nil {
			return nil, err
		}
	}

	return sa, nil
}

func (tc *E2ETestFramework) createServiceAccount(namespace, name string) (serviceAccount *corev1.ServiceAccount, err error) {

	serviceAccount = runtime.NewServiceAccount(namespace, name)
	clolog.V(3).Info("Creating serviceaccount", "serviceaccount", serviceAccount)
	tc.AddCleanup(func() error {
		return tc.Test.Delete(serviceAccount)
	})

	if err = tc.Test.Recreate(serviceAccount); err != nil {
		return nil, err
	}
	return serviceAccount, nil
}

func (tc *E2ETestFramework) createRbac(namespace, name string) (err error) {
	saRole := runtime.NewRole(
		namespace,
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
	tc.AddCleanup(func() error {
		return tc.Test.Delete(saRole)
	})
	if err = tc.Test.Recreate(saRole); err != nil {
		return err
	}
	subject := runtime.NewSubject(
		"ServiceAccount",
		name,
	)
	subject.APIGroup = ""
	roleBinding := runtime.NewRoleBinding(
		namespace,
		name,
		rbacv1.RoleRef{
			Kind:     "Role",
			Name:     saRole.Name,
			APIGroup: rbacv1.GroupName,
		}, runtime.NewSubjects(
			subject,
		)...,
	)
	tc.AddCleanup(func() error {
		return tc.Test.Delete(roleBinding)
	})

	return tc.Test.Recreate(roleBinding)
}
