package k8shandler

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateVisualization reconciles visualization component for cluster logging
func (cluster *ClusterLogging) CreateOrUpdateVisualization() (err error) {

	if cluster.Spec.Visualization.Type == logging.VisualizationTypeKibana {
		if err = cluster.createOrUpdateKibanaServiceAccount(); err != nil {
			return
		}

		if err = cluster.createOrUpdateKibanaService(); err != nil {
			return
		}

		if err = cluster.createOrUpdateKibanaRoute(); err != nil {
			return
		}

		oauthSecret := utils.GetWorkingDirFileContents("kibana-proxy-oauth.secret")
		if oauthSecret == nil {
			oauthSecret = utils.GetRandomWord(64)
			utils.WriteToWorkingDirFile("kibana-proxy-oauth.secret", oauthSecret)
		}

		if err = cluster.createOrUpdateKibanaSecret(oauthSecret); err != nil {
			return
		}

		if err = cluster.createOrUpdateOauthClient(string(oauthSecret)); err != nil {
			return
		}

		if err = cluster.createOrUpdateKibanaDeployment(); err != nil {
			return
		}

		kibanaStatus, err := cluster.getKibanaStatus()

		if err != nil {
			return fmt.Errorf("Failed to get Kibana status for %q: %v", cluster.Name, err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if exists := cluster.Exists(); exists {
				if !reflect.DeepEqual(kibanaStatus, cluster.Status.Visualization.KibanaStatus) {
					if printUpdateMessage {
						logrus.Infof("Updating status of Kibana for %q", cluster.Name)
						printUpdateMessage = false
					}
					cluster.Status.Visualization.KibanaStatus = kibanaStatus
					return sdk.Update(cluster)
				}
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Kibana status for %q: %v", cluster.Name, retryErr)
		}
	} else {
		cluster.removeKibana()
	}

	return nil
}

func (cluster *ClusterLogging) removeKibana() (err error) {
	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		name := "kibana"
		proxyName := "kibana-proxy"
		if err = utils.RemoveServiceAccount(cluster.Namespace, name); err != nil {
			return
		}

		if err = utils.RemoveConfigMap(cluster.Namespace, name); err != nil {
			return
		}

		if err = utils.RemoveConfigMap(cluster.Namespace, "sharing-config"); err != nil {
			return
		}

		if err = utils.RemoveSecret(cluster.Namespace, name); err != nil {
			return
		}

		if err = utils.RemoveSecret(cluster.Namespace, proxyName); err != nil {
			return
		}

		if err = utils.RemoveService(cluster.Namespace, name); err != nil {
			return
		}

		if err = utils.RemoveRoute(cluster.Namespace, name); err != nil {
			return
		}

		if err = utils.RemoveDeployment(cluster.Namespace, name); err != nil {
			return
		}

		if err = utils.RemoveOAuthClient(cluster.Namespace, proxyName); err != nil {
			return
		}
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateKibanaServiceAccount() error {

	kibanaServiceAccount := utils.ServiceAccount("kibana", cluster.Namespace)

	cluster.AddOwnerRefTo(kibanaServiceAccount)

	err := sdk.Create(kibanaServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana service account for %q: %v", cluster.Name, err)
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateKibanaDeployment() (err error) {

	kibanaPodSpec := cluster.newKibanaPodSpec("kibana", "elasticsearch")
	kibanaDeployment := utils.Deployment(
		"kibana",
		cluster.Namespace,
		"kibana",
		"kibana",
		kibanaPodSpec,
	)

	cluster.AddOwnerRefTo(kibanaDeployment)

	err = sdk.Create(kibanaDeployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana deployment for %q: %v", cluster.Name, err)
	}

	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return updateKibanaIfRequired(kibanaDeployment)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateOauthClient(oauthSecret string) (err error) {

	if err != nil {
		return nil
	}

	redirectURIs := []string{}

	kibanaURL, err := utils.GetRouteURL("kibana", cluster.Namespace)
	if err != nil {
		return err
	}

	redirectURIs = []string{
		kibanaURL,
	}

	oauthClient := utils.OAuthClient(
		"kibana-proxy",
		cluster.Namespace,
		oauthSecret,
		redirectURIs,
		[]string{
			"user:info",
			"user:check-access",
			"user:list-projects",
		},
	)

	cluster.AddOwnerRefTo(oauthClient)

	err = sdk.Create(oauthClient)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing %v oauthclient for %q: %v", oauthClient.Name, cluster.Name, err)
		}

		current := oauthClient.DeepCopy()
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = sdk.Get(current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v oauthclient for %q: %v", oauthClient.Name, cluster.Name, err)
			}

			current.RedirectURIs = oauthClient.RedirectURIs
			current.Secret = oauthClient.Secret
			current.ScopeRestrictions = oauthClient.ScopeRestrictions
			if err = sdk.Update(current); err != nil {
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

func (cluster *ClusterLogging) createOrUpdateKibanaRoute() error {

	kibanaRoute := utils.Route(
		"kibana",
		cluster.Namespace,
		"kibana",
		utils.GetWorkingDirFilePath("ca.crt"),
	)

	utils.AddOwnerRefToObject(kibanaRoute, utils.AsOwner(cluster))

	err := sdk.Create(kibanaRoute)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route for %q: %v", cluster.Name, err)
	}

	kibanaURL, err := utils.GetRouteURL("kibana", cluster.Namespace)
	if err != nil {
		return err
	}

	sharedConfig := createSharedConfig(cluster.Namespace, kibanaURL, kibanaURL)
	utils.AddOwnerRefToObject(sharedConfig, utils.AsOwner(cluster))

	err = sdk.Create(sharedConfig)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config: %v", err)
	}

	sharedRole := utils.NewRole(
		"sharing-config-reader",
		cluster.Namespace,
		utils.NewPolicyRules(
			utils.NewPolicyRule(
				[]string{""},
				[]string{"configmaps"},
				[]string{"sharing-config"},
				[]string{"get"},
			),
		),
	)

	cluster.AddOwnerRefTo(sharedRole)

	err = sdk.Create(sharedRole)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config role for %q: %v", cluster.Name, err)
	}

	sharedRoleBinding := utils.NewRoleBinding(
		"openshift-logging-sharing-config-reader-binding",
		cluster.Namespace,
		"sharing-config-reader",
		utils.NewSubjects(
			utils.NewSubject(
				"Group",
				"system:authenticated",
			),
		),
	)

	cluster.AddOwnerRefTo(sharedRoleBinding)

	err = sdk.Create(sharedRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config role binding for %q: %v", cluster.Name, err)
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateKibanaService() error {

	kibanaService := utils.Service(
		"kibana",
		cluster.Namespace,
		"kibana",
		[]v1.ServicePort{
			{Port: 443, TargetPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "oaproxy",
			}},
		})

	cluster.AddOwnerRefTo(kibanaService)

	err := sdk.Create(kibanaService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana service for %q: %v", cluster.Name, err)
	}

	return nil
}

func (cluster *ClusterLogging) createOrUpdateKibanaSecret(oauthSecret []byte) error {

	kibanaSecret := utils.Secret(
		"kibana",
		cluster.Namespace,
		map[string][]byte{
			"ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"key":  utils.GetWorkingDirFileContents("system.logging.kibana.key"),
			"cert": utils.GetWorkingDirFileContents("system.logging.kibana.crt"),
		})

	cluster.AddOwnerRefTo(kibanaSecret)

	err := utils.CreateOrUpdateSecret(kibanaSecret)
	if err != nil {
		return err
	}

	proxySecret := utils.Secret(
		"kibana-proxy",
		cluster.Namespace,
		map[string][]byte{
			"oauth-secret":   oauthSecret,
			"session-secret": utils.GetRandomWord(32),
			"server-key":     utils.GetWorkingDirFileContents("kibana-internal.key"),
			"server-cert":    utils.GetWorkingDirFileContents("kibana-internal.crt"),
		})

	cluster.AddOwnerRefTo(proxySecret)

	err = utils.CreateOrUpdateSecret(proxySecret)
	if err != nil {
		return err
	}

	return nil
}

func (cluster *ClusterLogging) newKibanaPodSpec(kibanaName string, elasticsearchName string) v1.PodSpec {

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
	kibanaContainer := utils.Container(
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
	kibanaProxyContainer := utils.Container(
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

	kibanaPodSpec := utils.PodSpec(
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

func updateKibanaIfRequired(desired *apps.Deployment) (err error) {
	current := desired.DeepCopy()

	if err = sdk.Get(current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Kibana deployment: %v", err)
	}

	current, different := isKibanaDifferent(current, desired)

	if different {
		if err = sdk.Update(current); err != nil {
			return err
		}
	}

	return nil
}

func isKibanaDifferent(current *apps.Deployment, desired *apps.Deployment) (*apps.Deployment, bool) {

	different := false

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
	return utils.ConfigMap(
		"sharing-config",
		namespace,
		map[string]string{
			"kibanaAppURL":   kibanaAppURL,
			"kibanaInfraURL": kibanaInfraURL,
		},
	)
}
