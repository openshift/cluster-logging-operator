package k8shandler

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	rbac "k8s.io/api/rbac/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (elasticsearchRequest *ElasticsearchRequest) CreateOrUpdateRBAC() error {

	dpl := elasticsearchRequest.cluster

	owner := getOwnerRef(dpl)

	// elasticsearch RBAC
	elasticsearchRole := newClusterRole(
		"elasticsearch-metrics",
		newPolicyRules(
			newPolicyRule(
				[]string{""},
				[]string{"pods", "services", "endpoints"},
				[]string{},
				[]string{"list", "watch"},
				[]string{},
			),
			newPolicyRule(
				[]string{},
				[]string{},
				[]string{},
				[]string{"get"},
				[]string{"/metrics"},
			),
		),
	)

	addOwnerRefToObject(elasticsearchRole, owner)

	if err := createOrUpdateClusterRole(elasticsearchRole, elasticsearchRequest.client); err != nil {
		return err
	}

	subject := newSubject(
		"ServiceAccount",
		"prometheus-k8s",
		"openshift-monitoring",
	)
	subject.APIGroup = ""

	elasticsearchRoleBinding := newClusterRoleBinding(
		"elasticsearch-metrics",
		"elasticsearch-metrics",
		newSubjects(
			subject,
		),
	)

	addOwnerRefToObject(elasticsearchRoleBinding, owner)

	if err := createOrUpdateClusterRoleBinding(elasticsearchRoleBinding, elasticsearchRequest.client); err != nil {
		return err
	}

	// proxy RBAC
	proxyRole := newClusterRole(
		"elasticsearch-proxy",
		newPolicyRules(
			newPolicyRule(
				[]string{"authentication.k8s.io"},
				[]string{"tokenreviews"},
				[]string{},
				[]string{"create"},
				[]string{},
			),
			newPolicyRule(
				[]string{"authorization.k8s.io"},
				[]string{"subjectaccessreviews"},
				[]string{},
				[]string{"create"},
				[]string{},
			),
		),
	)

	addOwnerRefToObject(proxyRole, owner)

	if err := createOrUpdateClusterRole(proxyRole, elasticsearchRequest.client); err != nil {
		return err
	}

	// Cluster role elasticsearch-proxy has to contain subjects for all ES instances
	esList := &v1.ElasticsearchList{}
	err := elasticsearchRequest.client.List(context.TODO(), &client.ListOptions{}, esList)
	if err != nil {
		return err
	}

	subjects := []rbac.Subject{}
	for _, es := range esList.Items {
		subject = newSubject(
			"ServiceAccount",
			es.Name,
			es.Namespace,
		)
		subject.APIGroup = ""
		subjects = append(subjects, subject)
	}

	proxyRoleBinding := newClusterRoleBinding(
		"elasticsearch-proxy",
		"elasticsearch-proxy",
		subjects,
	)

	addOwnerRefToObject(proxyRoleBinding, owner)

	return createOrUpdateClusterRoleBinding(proxyRoleBinding, elasticsearchRequest.client)
}

func createOrUpdateClusterRole(role *rbac.ClusterRole, client client.Client) error {
	if err := client.Create(context.TODO(), role); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create ClusterRole %s: %v", role.Name, err)
		}
		existingRole := role.DeepCopy()
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if getErr := client.Get(context.TODO(), types.NamespacedName{Name: existingRole.Name, Namespace: existingRole.Namespace}, existingRole); getErr != nil {
				logrus.Debugf("could not get ClusterRole %v: %v", existingRole.Name, getErr)
				return getErr
			}
			existingRole.Rules = role.Rules
			if updateErr := client.Update(context.TODO(), existingRole); updateErr != nil {
				logrus.Debugf("failed to update ClusterRole %v status: %v", existingRole.Name, updateErr)
				return updateErr
			}
			return nil
		})
	}
	return nil
}

func createOrUpdateClusterRoleBinding(roleBinding *rbac.ClusterRoleBinding, client client.Client) error {
	if err := client.Create(context.TODO(), roleBinding); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create ClusterRoleBindig %s: %v", roleBinding.Name, err)
		}
		existingRoleBinding := roleBinding.DeepCopy()
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if getErr := client.Get(context.TODO(), types.NamespacedName{Name: existingRoleBinding.Name, Namespace: existingRoleBinding.Namespace}, existingRoleBinding); getErr != nil {
				return fmt.Errorf("could not get ClusterRole %v: %v", existingRoleBinding.Name, getErr)
			}
			existingRoleBinding.Subjects = roleBinding.Subjects
			if updateErr := client.Update(context.TODO(), existingRoleBinding); updateErr != nil {
				return fmt.Errorf("failed to update ClusterRoleBinding %v status: %v", existingRoleBinding.Name, updateErr)
			}
			return nil
		})
	}
	return nil
}

func newPolicyRule(apiGroups, resources, resourceNames, verbs, urls []string) rbac.PolicyRule {
	return rbac.PolicyRule{
		APIGroups:       apiGroups,
		Resources:       resources,
		ResourceNames:   resourceNames,
		Verbs:           verbs,
		NonResourceURLs: urls,
	}
}

func newPolicyRules(rules ...rbac.PolicyRule) []rbac.PolicyRule {
	return rules
}

func newClusterRole(roleName string, rules []rbac.PolicyRule) *rbac.ClusterRole {
	return &rbac.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: rules,
	}
}

func newSubject(kind, name, namespace string) rbac.Subject {
	return rbac.Subject{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		APIGroup:  rbac.GroupName,
	}
}

func newSubjects(subjects ...rbac.Subject) []rbac.Subject {
	return subjects
}

func newClusterRoleBinding(bindingName, roleName string, subjects []rbac.Subject) *rbac.ClusterRoleBinding {
	return &rbac.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingName,
		},
		RoleRef: rbac.RoleRef{
			Kind:     "ClusterRole",
			Name:     roleName,
			APIGroup: rbac.GroupName,
		},
		Subjects: subjects,
	}
}
