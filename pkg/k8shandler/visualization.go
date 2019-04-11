package k8shandler

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateVisualization reconciles visualization component for cluster logging
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateVisualization() (err error) {

	if clusterRequest.cluster.Spec.Visualization.Type == logging.VisualizationTypeKibana {

		cluster := clusterRequest.cluster

		if err = clusterRequest.CreateOrUpdateServiceAccount("kibana"); err != nil {
			return
		}

		if err = clusterRequest.createKibanaProxyClusterRoleBinding(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaService(); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaRoute(); err != nil {
			return
		}

		oauthSecret := utils.GetWorkingDirFileContents("kibana-proxy-oauth.secret")
		if oauthSecret == nil {
			oauthSecret = utils.GetRandomWord(64)
			utils.WriteToWorkingDirFile("kibana-proxy-oauth.secret", oauthSecret)
		}

		if err = clusterRequest.createOrUpdateKibanaSecret(oauthSecret); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateOauthClient(string(oauthSecret)); err != nil {
			return
		}

		if err = clusterRequest.createOrUpdateKibanaDeployment(); err != nil {
			return
		}

		kibanaStatus, err := clusterRequest.getKibanaStatus()

		if err != nil {
			return fmt.Errorf("Failed to get Kibana status for %q: %v", cluster.Name, err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if !reflect.DeepEqual(kibanaStatus, cluster.Status.Visualization.KibanaStatus) {
				if printUpdateMessage {
					logrus.Infof("Updating status of Kibana for %q", cluster.Name)
					printUpdateMessage = false
				}
				cluster.Status.Visualization.KibanaStatus = kibanaStatus
				return clusterRequest.Update(cluster)
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Kibana status for %q: %v", cluster.Name, retryErr)
		}
	} else {
		clusterRequest.removeKibana()
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeKibana() (err error) {
	if clusterRequest.isManaged() {
		name := "kibana"
		proxyName := "kibana-proxy"
		if err = clusterRequest.RemoveServiceAccount(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("sharing-config"); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret(proxyName); err != nil {
			return
		}

		if err = clusterRequest.RemoveService(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveRoute(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveDeployment(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveOAuthClient(proxyName); err != nil {
			return
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createKibanaProxyClusterRoleBinding() (err error) {

	subject := NewSubject(
		"ServiceAccount",
		"kibana",
	)
	subject.Namespace = clusterRequest.cluster.Namespace
	subject.APIGroup = ""

	clusterRoleBinding := NewClusterRoleBinding(
		"kibana-proxy-oauth-delegator",
		"system:auth-delegator",
		NewSubjects(subject),
	)

	utils.AddOwnerRefToObject(clusterRoleBinding, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(clusterRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana clusterrolebinding %q: %v", clusterRoleBinding.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaDeployment() (err error) {

	kibanaPodSpec := newKibanaPodSpec(clusterRequest.cluster, "kibana", "elasticsearch")
	kibanaDeployment := NewDeployment(
		"kibana",
		clusterRequest.cluster.Namespace,
		"kibana",
		"kibana",
		kibanaPodSpec,
	)

	kibanaDeployment.Spec.Replicas = &clusterRequest.cluster.Spec.Visualization.KibanaSpec.Replicas

	utils.AddOwnerRefToObject(kibanaDeployment, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(kibanaDeployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana deployment for %q: %v", clusterRequest.cluster.Name, err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateKibanaIfRequired(kibanaDeployment)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateOauthClient(oauthSecret string) (err error) {

	if err != nil {
		return nil
	}

	redirectURIs := []string{}

	kibanaURL, err := clusterRequest.GetRouteURL("kibana")
	if err != nil {
		return err
	}

	redirectURIs = []string{
		kibanaURL,
	}

	oauthClient := NewOAuthClient(
		"kibana-proxy",
		clusterRequest.cluster.Namespace,
		oauthSecret,
		redirectURIs,
		[]string{
			"user:info",
			"user:check-access",
			"user:list-projects",
		},
	)

	utils.AddOwnerRefToObject(oauthClient, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.Create(oauthClient)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing %v oauthclient for %q: %v", oauthClient.Name, clusterRequest.cluster.Name, err)
		}

		current := oauthClient.DeepCopy()

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(oauthClient.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v oauthclient for %q: %v", oauthClient.Name, clusterRequest.cluster.Name, err)
			}

			current.RedirectURIs = oauthClient.RedirectURIs
			current.Secret = oauthClient.Secret
			current.ScopeRestrictions = oauthClient.ScopeRestrictions
			if err = clusterRequest.Update(current); err != nil {
				return err
			}
			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
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

	err := clusterRequest.Create(kibanaRoute)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route for %q: %v", cluster.Name, err)
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

func (clusterRequest *ClusterLoggingRequest) createOrUpdateKibanaSecret(oauthSecret []byte) error {

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

	proxySecret := NewSecret(
		"kibana-proxy",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"oauth-secret":   oauthSecret,
			"session-secret": utils.GetRandomWord(32),
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

func newKibanaPodSpec(cluster *logging.ClusterLogging, kibanaName string, elasticsearchName string) v1.PodSpec {

	var kibanaResources = cluster.Spec.Visualization.KibanaSpec.Resources
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

	var kibanaProxyResources = cluster.Spec.Visualization.ProxySpec.Resources
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
		"-client-id=kibana-proxy",
		"-client-secret-file=/secret/oauth-secret",
		"-cookie-secret-file=/secret/session-secret",
		"-upstream=http://localhost:5601",
		"-scope=user:info user:check-access user:list-projects",
		"--tls-cert=/secret/server-cert",
		"-tls-key=/secret/server-key",
		"-pass-access-token",
		"-skip-provider-button",
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

	kibanaProxyContainer.Ports = []v1.ContainerPort{
		{Name: "oaproxy", ContainerPort: 3000},
	}

	kibanaProxyContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "kibana-proxy", ReadOnly: true, MountPath: "/secret"},
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
		cluster.Spec.Visualization.KibanaSpec.NodeSelector,
	)

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

func (clusterRequest *ClusterLoggingRequest) updateKibanaIfRequired(desired *apps.Deployment) (err error) {
	current := desired.DeepCopy()

	if err = clusterRequest.Get(desired.Name, current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Kibana deployment: %v", err)
	}

	current, different := isKibanaDifferent(current, desired)

	if different {
		if err = clusterRequest.Update(current); err != nil {
			return err
		}
	}

	return nil
}

func isKibanaDifferent(current *apps.Deployment, desired *apps.Deployment) (*apps.Deployment, bool) {

	different := false

	if !utils.AreSelectorsSame(current.Spec.Template.Spec.NodeSelector, desired.Spec.Template.Spec.NodeSelector) {
		logrus.Infof("Invalid Kibana nodeSelector change found, updating '%s'", current.Name)
		current.Spec.Template.Spec.NodeSelector = desired.Spec.Template.Spec.NodeSelector
		different = true
	}

	if *current.Spec.Replicas != *desired.Spec.Replicas {
		logrus.Infof("Invalid Kibana replica count found, updating %q", current.Name)
		current.Spec.Replicas = desired.Spec.Replicas
		different = true
	}

	if isDeploymentImageDifference(current, desired) {
		logrus.Infof("Kibana image(s) change found, updating %q", current.Name)
		current = updateCurrentImages(current, desired)
		different = true
	}

	if utils.AreResourcesDifferent(current, desired) {
		logrus.Infof("Kibana resource(s) change found, updating %q", current.Name)
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

func updateCurrentImages(current *apps.Deployment, desired *apps.Deployment) *apps.Deployment {

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
