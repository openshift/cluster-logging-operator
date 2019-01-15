package utils

// factory.go is a set of factory methods for initializing
// various OpenShift structs

import (
	oauth "github.com/openshift/api/oauth/v1"
	route "github.com/openshift/api/route/v1"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	scheduling "k8s.io/api/scheduling/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CronJob(cronjobName string, namespace string, loggingComponent string, component string, cronjobSpec batch.CronJobSpec) *batch.CronJob {
	return &batch.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: batch.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronjobName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
		},
		Spec: cronjobSpec,
	}
}

func Secret(secretName string, namespace string, data map[string][]byte) *core.Secret {
	return &core.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: "Opaque",
		Data: data,
	}
}

func NewServiceAccount(accountName string, namespace string) *core.ServiceAccount {
	return ServiceAccount(accountName, namespace)
}
func ServiceAccount(accountName string, namespace string) *core.ServiceAccount {
	return &core.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      accountName,
			Namespace: namespace,
		},
	}
}

func Service(serviceName string, namespace string, selectorComponent string, servicePorts []core.ServicePort) *core.Service {
	return &core.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"logging-infra": "support",
			},
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{
				"component": selectorComponent,
				"provider":  "openshift",
			},
			Ports: servicePorts,
		},
	}
}

func Route(routeName string, namespace string, serviceName string, cafilePath string) *route.Route {
	return &route.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: route.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeName,
			Namespace: namespace,
			Labels: map[string]string{
				"component":     "support",
				"logging-infra": "support",
				"provider":      "openshift",
			},
		},
		Spec: route.RouteSpec{
			To: route.RouteTargetReference{
				Name: serviceName,
				Kind: "Service",
			},
			TLS: &route.TLSConfig{
				Termination:                   route.TLSTerminationReencrypt,
				InsecureEdgeTerminationPolicy: route.InsecureEdgeTerminationPolicyRedirect,
				CACertificate:                 string(GetFileContents(cafilePath)),
				DestinationCACertificate:      string(GetFileContents(cafilePath)),
			},
		},
	}
}

//NewPodSpec is a constructor to instansiate a new PodSpec
func NewPodSpec(serviceAccountName string, containers []core.Container, volumes []core.Volume) core.PodSpec {
	return PodSpec(serviceAccountName, containers, volumes)
}
func PodSpec(serviceAccountName string, containers []core.Container, volumes []core.Volume) core.PodSpec {
	return core.PodSpec{
		Containers:         containers,
		ServiceAccountName: serviceAccountName,
		Volumes:            volumes,
	}
}

func Container(containerName string, pullPolicy core.PullPolicy, resources core.ResourceRequirements) core.Container {
	return core.Container{
		Name:            containerName,
		Image:           GetComponentImage(containerName),
		ImagePullPolicy: pullPolicy,
		Resources:       resources,
	}
}

// loggingComponent = kibana
// component = kibana{,-ops}
func Deployment(deploymentName string, namespace string, loggingComponent string, component string, podSpec core.PodSpec) *apps.Deployment {
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: GetInt32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"provider":      "openshift",
					"component":     component,
					"logging-infra": loggingComponent,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: deploymentName,
					Labels: map[string]string{
						"provider":      "openshift",
						"component":     component,
						"logging-infra": loggingComponent,
					},
				},
				Spec: podSpec,
			},
			Strategy: apps.DeploymentStrategy{
				Type: apps.RollingUpdateDeploymentStrategyType,
				//RollingUpdate: {}
			},
		},
	}
}

func DaemonSet(daemonsetName string, namespace string, loggingComponent string, component string, podSpec core.PodSpec) *apps.DaemonSet {
	return &apps.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonsetName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
		},
		Spec: apps.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"provider":      "openshift",
					"component":     component,
					"logging-infra": loggingComponent,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: daemonsetName,
					Labels: map[string]string{
						"provider":      "openshift",
						"component":     component,
						"logging-infra": loggingComponent,
					},
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: podSpec,
			},
			UpdateStrategy: apps.DaemonSetUpdateStrategy{
				Type:          apps.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &apps.RollingUpdateDaemonSet{},
			},
			MinReadySeconds: 600,
		},
	}
}

func ConfigMap(configmapName string, namespace string, data map[string]string) *core.ConfigMap {
	return &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: data,
	}
}

func OAuthClient(oauthClientName, namespace, oauthSecret string, redirectURIs, scopeRestrictions []string) *oauth.OAuthClient {

	return &oauth.OAuthClient{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OAuthClient",
			APIVersion: oauth.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: oauthClientName,
			Labels: map[string]string{
				"logging-infra": "support",
				"namespace":     namespace,
			},
		},
		Secret:       oauthSecret,
		RedirectURIs: redirectURIs,
		GrantMethod:  oauth.GrantHandlerPrompt,
		ScopeRestrictions: []oauth.ScopeRestriction{
			oauth.ScopeRestriction{
				ExactValues: scopeRestrictions,
			},
		},
	}

}

//NewPriorityClass is a constructor to create a PriorityClass
func NewPriorityClass(priorityclassName string, priorityValue int32, globalDefault bool, description string) *scheduling.PriorityClass {
	return PriorityClass(priorityclassName, priorityValue, globalDefault, description)
}
func PriorityClass(priorityclassName string, priorityValue int32, globalDefault bool, description string) *scheduling.PriorityClass {
	return &scheduling.PriorityClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PriorityClass",
			APIVersion: scheduling.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: priorityclassName,
		},
		Value:         priorityValue,
		GlobalDefault: globalDefault,
		Description:   description,
	}
}

func NewPolicyRule(apiGroups, resources, resourceNames, verbs []string) rbac.PolicyRule {
	return rbac.PolicyRule{
		APIGroups:     apiGroups,
		Resources:     resources,
		ResourceNames: resourceNames,
		Verbs:         verbs,
	}
}

func NewPolicyRules(rules ...rbac.PolicyRule) []rbac.PolicyRule {
	return rules
}

func NewRole(roleName, namespace string, rules []rbac.PolicyRule) *rbac.Role {
	return &rbac.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
		Rules: rules,
	}
}

func NewSubject(kind, name string) rbac.Subject {
	return rbac.Subject{
		Kind:     kind,
		Name:     name,
		APIGroup: rbac.GroupName,
	}
}

func NewSubjects(subjects ...rbac.Subject) []rbac.Subject {
	return subjects
}

func NewRoleBinding(bindingName, namespace, roleName string, subjects []rbac.Subject) *rbac.RoleBinding {
	return &rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
		},
		RoleRef: rbac.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: rbac.GroupName,
		},
		Subjects: subjects,
	}
}

func NewClusterRoleBinding(bindingName, roleName string, subjects []rbac.Subject) *rbac.ClusterRoleBinding {
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
