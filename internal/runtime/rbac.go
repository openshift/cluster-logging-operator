package runtime

import rbacv1 "k8s.io/api/rbac/v1"

// NewPolicyRule stubs policy rule
func NewPolicyRule(apiGroups, resources, resourceNames, verbs []string) rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		APIGroups:     apiGroups,
		Resources:     resources,
		ResourceNames: resourceNames,
		Verbs:         verbs,
	}
}

// NewPolicyRules stubs policy rules
func NewPolicyRules(rules ...rbacv1.PolicyRule) []rbacv1.PolicyRule {
	return rules
}

// NewRole returns a role with namespace, names, rules
func NewRole(namespace, name string, rules ...rbacv1.PolicyRule) *rbacv1.Role {
	role := &rbacv1.Role{
		Rules: rules,
	}
	Initialize(role, namespace, name)
	return role
}

// NewClusterRole returns a role with namespace, names, rules
func NewClusterRole(name string, rules ...rbacv1.PolicyRule) *rbacv1.ClusterRole {
	role := &rbacv1.ClusterRole{
		Rules: rules,
	}
	Initialize(role, "", name)
	return role
}

// NewRoleBinding returns a role with namespace, names, rules
func NewRoleBinding(namespace, name string, roleRef rbacv1.RoleRef, subjects ...rbacv1.Subject) *rbacv1.RoleBinding {
	binding := &rbacv1.RoleBinding{
		RoleRef:  roleRef,
		Subjects: subjects,
	}
	Initialize(binding, namespace, name)
	return binding
}

// NewSubject stubs a new subject
func NewSubject(kind, name string) rbacv1.Subject {
	return rbacv1.Subject{
		Kind:     kind,
		Name:     name,
		APIGroup: rbacv1.GroupName,
	}
}

// NewSubjects stubs subjects
func NewSubjects(subjects ...rbacv1.Subject) []rbacv1.Subject {
	return subjects
}

// NewClusterRoleBinding returns a role with namespace, names, rules
func NewClusterRoleBinding(name string, roleRef rbacv1.RoleRef, subjects ...rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	binding := &rbacv1.ClusterRoleBinding{
		RoleRef:  roleRef,
		Subjects: subjects,
	}
	Initialize(binding, "", name)
	return binding
}
