package k8shandler

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kibanaServiceAccountName     = "kibana"
	kibanaOAuthRedirectReference = "{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"kibana\"}}"
	// The following strings are turned into JavaScript RegExps. Online tool to test them: https://regex101.com/
	nodesAndContainersNamespaceFilter = "^(openshift-.*|kube-.*|openshift$|kube$|default$)"
	appsNamespaceFilter               = "^((?!" + nodesAndContainersNamespaceFilter + ").)*$" // ^((?!^(openshift-.*|kube-.*|openshift$|kube$|default$)).)*$
)

var (
	kibanaServiceAccountAnnotations = map[string]string{
		"serviceaccounts.openshift.io/oauth-redirectreference.first": kibanaOAuthRedirectReference,
	}
)

// CreateOrUpdateVisualization reconciles visualization component for cluster logging
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateVisualization(proxyConfig *configv1.Proxy) (err error) {
	if clusterRequest.cluster.Spec.Visualization == nil || clusterRequest.cluster.Spec.Visualization.Type == "" {
		if err = clusterRequest.removeKibana(); err != nil {
			return
		}
		return nil
	}
	if clusterRequest.cluster.Spec.Visualization.Type == logging.VisualizationTypeKibana {

		if err = clusterRequest.CreateOrUpdateServiceAccount(kibanaServiceAccountName, &kibanaServiceAccountAnnotations); err != nil {
			return
		}

		if err = clusterRequest.RemoveClusterRoleBinding("kibana-proxy-oauth-delegator"); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaService(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaRoute(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaSecret(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaConsoleExternalLogLink(); err != nil {
			return
		}

		if err = clusterRequest.RemoveOAuthClient("kibana-proxy"); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaDeployment(proxyConfig); err != nil {
			return
		}

		clusterRequest.UpdateKibanaStatus()
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) UpdateKibanaStatus() (err error) {

	kibanaStatus, err := clusterRequest.getKibanaStatus()
	cluster := clusterRequest.cluster

	if err != nil {
		return fmt.Errorf("Failed to get Kibana status for %q: %v", cluster.Name, err)
	}

	printUpdateMessage := true
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if !compareKibanaStatus(kibanaStatus, cluster.Status.Visualization.KibanaStatus) {
			if printUpdateMessage {
				logrus.Infof("Updating status of Kibana")
				printUpdateMessage = false
			}
			cluster.Status.Visualization.KibanaStatus = kibanaStatus
			return clusterRequest.UpdateStatus(cluster)
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Kibana status for %q: %v", cluster.Name, retryErr)
	}

	return nil
}

func compareKibanaStatus(lhs, rhs []logging.KibanaStatus) bool {
	// there should only ever be a single kibana status object
	if len(lhs) != len(rhs) {
		return false
	}

	if len(lhs) > 0 {
		for index, _ := range lhs {
			if lhs[index].Deployment != rhs[index].Deployment {
				return false
			}

			if lhs[index].Replicas != rhs[index].Replicas {
				return false
			}

			if len(lhs[index].ReplicaSets) != len(rhs[index].ReplicaSets) {
				return false
			}

			if len(lhs[index].ReplicaSets) > 0 {
				if !reflect.DeepEqual(lhs[index].ReplicaSets, rhs[index].ReplicaSets) {
					return false
				}
			}

			if len(lhs[index].Pods) != len(rhs[index].Pods) {
				return false
			}

			if len(lhs[index].Pods) > 0 {
				if !reflect.DeepEqual(lhs[index].Pods, rhs[index].Pods) {
					return false
				}
			}

			if len(lhs[index].Conditions) != len(rhs[index].Conditions) {
				return false
			}

			if len(lhs[index].Conditions) > 0 {
				if !reflect.DeepEqual(lhs[index].Conditions, rhs[index].Conditions) {
					return false
				}
			}
		}
	}

	return true
}

func (clusterRequest *ClusterLoggingRequest) RestartKibana(proxyConfig *configv1.Proxy) (err error) {

	if err = clusterRequest.createOrUpdateKibanaDeployment(proxyConfig); err != nil {
		return
	}

	return clusterRequest.UpdateKibanaStatus()
}

func (clusterRequest *ClusterLoggingRequest) removeKibana() (err error) {
	if clusterRequest.isManaged() {
		name := "kibana"
		proxyName := "kibana-proxy"
		if err = clusterRequest.RemoveDeployment(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveOAuthClient(proxyName); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret(proxyName); err != nil {
			return
		}

		if err = clusterRequest.RemoveRoute(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("sharing-config"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(constants.KibanaTrustedCAName); err != nil {
			return
		}

		if err = clusterRequest.RemoveService(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveServiceAccount(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConsoleExternalLogLink(name); err != nil {
			return
		}

	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaDeployment(proxyConfig *configv1.Proxy) (err error) {

	kibanaTrustBundle := &v1.ConfigMap{}

	// Create cluster proxy trusted CA bundle.
	err = clusterRequest.createOrUpdateTrustedCABundleConfigMap(constants.KibanaTrustedCAName)
	if err != nil {
		return
	}

	kibanaPodSpec := newKibanaPodSpec(clusterRequest.cluster, "kibana", "elasticsearch.openshift-logging.svc.cluster.local", proxyConfig, kibanaTrustBundle)
	kibanaDeployment := NewDeployment(
		"kibana",
		clusterRequest.cluster.Namespace,
		"kibana",
		"kibana",
		kibanaPodSpec,
	)
	kibanaDeployment.Spec.Replicas = &clusterRequest.cluster.Spec.Visualization.KibanaSpec.Replicas

	// if we don't have the hash values we shouldn't start/create
	annotations, err := clusterRequest.getKibanaAnnotations(kibanaDeployment)
	if err != nil {
		return err
	}

	kibanaDeployment.Spec.Template.ObjectMeta.Annotations = annotations

	utils.AddOwnerRefToObject(kibanaDeployment, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(kibanaDeployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana deployment for %q: %v", clusterRequest.cluster.Name, err)
	}

	if clusterRequest.isManaged() {
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			current := &apps.Deployment{}

			if err := clusterRequest.Get(kibanaDeployment.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					logrus.Debugf("Returning nil. The deployment %q was not found even though create previously failed.  Was it culled?", kibanaDeployment.Name)
					return nil
				}
				return fmt.Errorf("Failed to get Kibana deployment: %v", err)
			}

			current, different := isDeploymentDifferent(current, kibanaDeployment)

			// Check trustedCA certs have been updated or not by comparing the hash values in annotation.
			if current.Spec.Template.ObjectMeta.Annotations[constants.TrustedCABundleHashName] != kibanaDeployment.Spec.Template.ObjectMeta.Annotations[constants.TrustedCABundleHashName] {
				different = true
			}

			// Check secret hash has been updated or not
			for _, secretName := range []string{"kibana", "kibana-proxy"} {

				hashKey := fmt.Sprintf("%s%s", constants.SecretHashPrefix, secretName)
				if kibanaDeployment.Spec.Template.ObjectMeta.Annotations[hashKey] != current.Spec.Template.ObjectMeta.Annotations[hashKey] {
					different = true
				}
			}

			if different {
				current.Spec = kibanaDeployment.Spec
				return clusterRequest.Update(current)
			}
			return nil
		})
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) getKibanaAnnotations(deployment *apps.Deployment) (map[string]string, error) {

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	annotations := deployment.Spec.Template.ObjectMeta.Annotations

	kibanaTrustBundle := &v1.ConfigMap{}
	kibanaTrustBundleName := types.NamespacedName{Name: constants.KibanaTrustedCAName, Namespace: constants.OpenshiftNS}
	if err := clusterRequest.client.Get(context.TODO(), kibanaTrustBundleName, kibanaTrustBundle); err != nil {
		if !errors.IsNotFound(err) {
			return annotations, err
		}
	}

	if _, ok := kibanaTrustBundle.Data[constants.TrustedCABundleKey]; !ok {
		return annotations, fmt.Errorf("%v does not yet contain expected key %v", kibanaTrustBundle.Name, constants.TrustedCABundleKey)
	}

	trustedCAHashValue, err := calcTrustedCAHashValue(kibanaTrustBundle)
	if err != nil {
		return annotations, fmt.Errorf("unable to calculate trusted CA value. E: %s", err.Error())
	}

	if trustedCAHashValue == "" {
		return annotations, fmt.Errorf("Did not receive hashvalue for trusted CA value")
	}

	annotations[constants.TrustedCABundleHashName] = trustedCAHashValue

	// generate secret hash
	for _, secretName := range []string{"kibana", "kibana-proxy"} {

		hashKey := fmt.Sprintf("%s%s", constants.SecretHashPrefix, secretName)

		secret, err := clusterRequest.GetSecret(secretName)
		if err != nil {
			return annotations, err
		}
		secretHashValue, err := calcSecretHashValue(secret)
		if err != nil {
			return annotations, err
		}

		annotations[hashKey] = secretHashValue
	}

	return annotations, nil
}

func isDeploymentDifferent(current *apps.Deployment, desired *apps.Deployment) (*apps.Deployment, bool) {

	different := false

	// is this needed?
	if !utils.AreMapsSame(current.Spec.Template.Spec.NodeSelector, desired.Spec.Template.Spec.NodeSelector) {
		logrus.Debugf("Visualization nodeSelector change found, updating '%s'", current.Name)
		current.Spec.Template.Spec.NodeSelector = desired.Spec.Template.Spec.NodeSelector
		different = true
	}

	// is this needed?
	if !utils.AreTolerationsSame(current.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		logrus.Debugf("Visualization tolerations change found, updating '%s'", current.Name)
		current.Spec.Template.Spec.Tolerations = desired.Spec.Template.Spec.Tolerations
		different = true
	}

	if isDeploymentImageDifference(current, desired) {
		logrus.Debugf("Visualization image change found, updating %q", current.Name)
		current = updateCurrentDeploymentImages(current, desired)
		different = true
	}

	if utils.AreResourcesDifferent(current, desired) {
		logrus.Debugf("Visualization resource(s) change found, updating %q", current.Name)
		different = true
	}

	if !utils.EnvValueEqual(current.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env) {
		current.Spec.Template.Spec.Containers[0].Env = desired.Spec.Template.Spec.Containers[0].Env
		different = true
	}

	return current, different
}

func isDeploymentImageDifference(current *apps.Deployment, desired *apps.Deployment) bool {

	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}

	return false
}

func updateCurrentDeploymentImages(current *apps.Deployment, desired *apps.Deployment) *apps.Deployment {

	containers := current.Spec.Template.Spec.Containers

	for index, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					containers[index].Image = des.Image
				}
			}
		}
	}

	return current
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaRoute() error {

	cluster := clusterRequest.cluster

	kibanaRoute := NewRoute(
		"kibana",
		cluster.Namespace,
		"kibana",
		utils.GetWorkingDirFilePath("ca.crt"),
	)

	utils.AddOwnerRefToObject(kibanaRoute, utils.AsOwner(cluster))

	if err := clusterRequest.CreateOrUpdateRoute(kibanaRoute); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	kibanaURL, err := clusterRequest.GetRouteURL("kibana")
	if err != nil {
		return err
	}

	sharedConfig := createSharedConfig(cluster.Namespace, kibanaURL, kibanaURL)
	utils.AddOwnerRefToObject(sharedConfig, utils.AsOwner(cluster))

	err = clusterRequest.Create(sharedConfig)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config: %v", err)
	}

	sharedRole := NewRole(
		"sharing-config-reader",
		cluster.Namespace,
		NewPolicyRules(
			NewPolicyRule(
				[]string{""},
				[]string{"configmaps"},
				[]string{"sharing-config"},
				[]string{"get"},
			),
		),
	)

	utils.AddOwnerRefToObject(sharedRole, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(sharedRole)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config role for %q: %v", cluster.Name, err)
	}

	sharedRoleBinding := NewRoleBinding(
		"openshift-logging-sharing-config-reader-binding",
		cluster.Namespace,
		"sharing-config-reader",
		NewSubjects(
			NewSubject(
				"Group",
				"system:authenticated",
			),
		),
	)

	utils.AddOwnerRefToObject(sharedRoleBinding, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(sharedRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config role binding for %q: %v", cluster.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaConsoleExternalLogLink() (err error) {
	cluster := clusterRequest.cluster

	kibanaURL, err := clusterRequest.GetRouteURL("kibana")
	if err != nil {
		return err
	}

	consoleExternalLogLinkForNodesAndContainers := NewConsoleExternalLogLink(
		"kibana-admin",
		cluster.Namespace,
		"Show in Kibana",
		strings.Join([]string{kibanaURL,
			"/app/kibana#/discover?_g=(time:(from:now-1w,mode:relative,to:now))&_a=(columns:!(kubernetes.container_name,message),index:",
			"'.operations.*'",
			",query:(query_string:(analyze_wildcard:!t,query:'",
			strings.Join([]string{
				"kubernetes.pod_name:\"${resourceName}\"",
				"kubernetes.namespace_name:\"${resourceNamespace}\"",
				"kubernetes.container_name.raw:\"${containerName}\"",
			}, " AND "),
			"')),sort:!('@timestamp',desc))"},
			""),
		nodesAndContainersNamespaceFilter,
	)

	consoleExternalLogLinkForApps := NewConsoleExternalLogLink(
		"kibana",
		cluster.Namespace,
		"Show in Kibana",
		strings.Join([]string{kibanaURL,
			"/app/kibana#/discover?_g=(time:(from:now-1w,mode:relative,to:now))&_a=(columns:!(kubernetes.container_name,message),index:",
			"'project.${resourceNamespace}.${resourceNamespaceUID}.*'",
			",query:(query_string:(analyze_wildcard:!t,query:'",
			strings.Join([]string{
				"kubernetes.pod_name:\"${resourceName}\"",
				"kubernetes.namespace_name:\"${resourceNamespace}\"",
				"kubernetes.container_name.raw:\"${containerName}\"",
			}, " AND "),
			"')),sort:!('@timestamp',desc))"},
			""),
		appsNamespaceFilter,
	)
	links := []*consolev1.ConsoleExternalLogLink{consoleExternalLogLinkForNodesAndContainers, consoleExternalLogLinkForApps}

	for _, externalLink := range links {
		utils.AddOwnerRefToObject(externalLink, utils.AsOwner(cluster))
		// In case the object already exists we delete it first
		if err = clusterRequest.RemoveConsoleExternalLogLink(externalLink.ObjectMeta.Name); err != nil {
			return
		}
		err = clusterRequest.Create(externalLink)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana console external %q log link for %q: %v",
				externalLink.ObjectMeta.Name, cluster.Name, err)
		}
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaService() error {

	kibanaService := NewService(
		"kibana",
		clusterRequest.cluster.Namespace,
		"kibana",
		[]v1.ServicePort{
			{Port: 443, TargetPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "oaproxy",
			}},
		})

	utils.AddOwnerRefToObject(kibanaService, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(kibanaService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana service for %q: %v", clusterRequest.cluster.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaSecret() error {

	kibanaSecret := NewSecret(
		"kibana",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"key":  utils.GetWorkingDirFileContents("system.logging.kibana.key"),
			"cert": utils.GetWorkingDirFileContents("system.logging.kibana.crt"),
		})

	utils.AddOwnerRefToObject(kibanaSecret, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.CreateOrUpdateSecret(kibanaSecret)
	if err != nil {
		return err
	}

	var sessionSecret []byte

	sessionSecret = utils.GetWorkingDirFileContents("kibana-session-secret")
	if sessionSecret == nil {
		sessionSecret = utils.GetRandomWord(32)
	}

	proxySecret := NewSecret(
		"kibana-proxy",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"session-secret": sessionSecret,
			"server-key":     utils.GetWorkingDirFileContents("kibana-internal.key"),
			"server-cert":    utils.GetWorkingDirFileContents("kibana-internal.crt"),
		})

	utils.AddOwnerRefToObject(proxySecret, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.CreateOrUpdateSecret(proxySecret)
	if err != nil {
		return err
	}

	return nil
}

func newKibanaPodSpec(cluster *logging.ClusterLogging, kibanaName string, elasticsearchName string, proxyConfig *configv1.Proxy, trustedCABundleCM *v1.ConfigMap) v1.PodSpec {
	visSpec := logging.VisualizationSpec{}
	if cluster.Spec.Visualization != nil {
		visSpec = *cluster.Spec.Visualization
	}
	var kibanaResources = visSpec.Resources
	if kibanaResources == nil {
		kibanaResources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultKibanaMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultKibanaMemory,
				v1.ResourceCPU:    defaultKibanaCpuRequest,
			},
		}
	}
	kibanaContainer := NewContainer(
		"kibana",
		"kibana",
		v1.PullIfNotPresent,
		*kibanaResources,
	)

	var endpoint bytes.Buffer

	endpoint.WriteString("https://")
	endpoint.WriteString(elasticsearchName)
	endpoint.WriteString(":9200")

	kibanaContainer.Env = []v1.EnvVar{
		{Name: "ELASTICSEARCH_URL", Value: endpoint.String()},
		{Name: "KIBANA_MEMORY_LIMIT",
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					ContainerName: "kibana",
					Resource:      "limits.memory",
				},
			},
		},
	}

	kibanaContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "kibana", ReadOnly: true, MountPath: "/etc/kibana/keys"},
	}

	kibanaContainer.ReadinessProbe = &v1.Probe{
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{
					"/usr/share/kibana/probe/readiness.sh",
				},
			},
		},
		InitialDelaySeconds: 5, TimeoutSeconds: 4, PeriodSeconds: 5,
	}

	var kibanaProxyResources = visSpec.ProxySpec.Resources
	if kibanaProxyResources == nil {
		kibanaProxyResources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultKibanaProxyMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultKibanaProxyMemory,
				v1.ResourceCPU:    defaultKibanaProxyCpuRequest,
			},
		}
	}
	kibanaProxyContainer := NewContainer(
		"kibana-proxy",
		"kibana-proxy",
		v1.PullIfNotPresent,
		*kibanaProxyResources,
	)

	kibanaProxyContainer.Args = []string{
		"--upstream-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		"--https-address=:3000",
		"-provider=openshift",
		"-client-id=system:serviceaccount:openshift-logging:kibana",
		"-client-secret-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
		"-cookie-secret-file=/secret/session-secret",
		"-upstream=http://localhost:5601",
		"-scope=user:info user:check-access user:list-projects",
		"--tls-cert=/secret/server-cert",
		"-tls-key=/secret/server-key",
		"-pass-access-token",
	}

	kibanaProxyContainer.Env = []v1.EnvVar{
		{Name: "OAP_DEBUG", Value: "false"},
		{Name: "OCP_AUTH_PROXY_MEMORY_LIMIT",
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					ContainerName: "kibana-proxy",
					Resource:      "limits.memory",
				},
			},
		},
	}

	proxyEnv := utils.SetProxyEnvVars(proxyConfig)
	kibanaProxyContainer.Env = append(kibanaProxyContainer.Env, proxyEnv...)

	kibanaProxyContainer.Ports = []v1.ContainerPort{
		{Name: "oaproxy", ContainerPort: 3000},
	}

	kibanaProxyContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "kibana-proxy", ReadOnly: true, MountPath: "/secret"},
	}

	addTrustedCAVolume := false
	// If trusted CA bundle ConfigMap exists and its hash value is non-zero, mount the bundle.
	if trustedCABundleCM != nil && hasTrustedCABundle(trustedCABundleCM) {
		addTrustedCAVolume = true
		kibanaProxyContainer.VolumeMounts = append(kibanaProxyContainer.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.KibanaTrustedCAName,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})
	}

	kibanaPodSpec := NewPodSpec(
		"kibana",
		[]v1.Container{kibanaContainer, kibanaProxyContainer},
		[]v1.Volume{
			{Name: "kibana", VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "kibana",
				},
			},
			},
			{Name: "kibana-proxy", VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "kibana-proxy",
				},
			},
			},
		},
		visSpec.NodeSelector,
		visSpec.Tolerations,
	)

	if addTrustedCAVolume {
		kibanaPodSpec.Volumes = append(kibanaPodSpec.Volumes,
			v1.Volume{
				Name: constants.KibanaTrustedCAName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: constants.KibanaTrustedCAName,
						},
						Items: []v1.KeyToPath{
							{
								Key:  constants.TrustedCABundleKey,
								Path: constants.TrustedCABundleMountFile,
							},
						},
					},
				},
			})
	}

	kibanaPodSpec.Affinity = &v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: v1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{Key: "logging-infra",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"kibana"},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}

	return kibanaPodSpec
}

func createSharedConfig(namespace, kibanaAppURL, kibanaInfraURL string) *v1.ConfigMap {
	return NewConfigMap(
		"sharing-config",
		namespace,
		map[string]string{
			"kibanaAppURL":   kibanaAppURL,
			"kibanaInfraURL": kibanaInfraURL,
		},
	)
}
