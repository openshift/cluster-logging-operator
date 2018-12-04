package k8shandler

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	oauth "github.com/openshift/api/oauth/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateOrUpdateVisualization(cluster *logging.ClusterLogging) (err error) {

	if cluster.Spec.Visualization.Type == logging.VisualizationTypeKibana {
		if err = createOrUpdateKibanaServiceAccount(cluster); err != nil {
			return
		}

		if err = createOrUpdateKibanaService(cluster); err != nil {
			return
		}

		if err = createOrUpdateKibanaRoute(cluster); err != nil {
			return
		}

		oauthSecret := utils.GetRandomWord(64)

		if err = createOrUpdateKibanaSecret(cluster, oauthSecret); err != nil {
			return
		}

		if err = createOrUpdateOauthClient(cluster, string(oauthSecret)); err != nil {
			return
		}

		if err = createOrUpdateKibanaDeployment(cluster); err != nil {
			return
		}

		kibanaStatus, err := getKibanaStatus(cluster.Namespace)

		if err != nil {
			return fmt.Errorf("Failed to get status for Kibana: %v", err)
		}

		if !reflect.DeepEqual(kibanaStatus, cluster.Status.Visualization.KibanaStatus) {
			logrus.Infof("Updating status of Kibana")
			cluster.Status.Visualization.KibanaStatus = kibanaStatus

			if err = sdk.Update(cluster); err != nil {
				return fmt.Errorf("Failed to update Cluster Logging Kibana status: %v", err)
			}
		}
	}

	return nil
}

func createOrUpdateKibanaServiceAccount(logging *logging.ClusterLogging) error {

	kibanaServiceAccount := utils.ServiceAccount("kibana", logging.Namespace)

	utils.AddOwnerRefToObject(kibanaServiceAccount, utils.AsOwner(logging))

	err := sdk.Create(kibanaServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana service account: %v", err)
	}

	return nil
}

func createOrUpdateKibanaDeployment(cluster *logging.ClusterLogging) (err error) {

	if utils.AllInOne(cluster) {
		kibanaPodSpec := getKibanaPodSpec(cluster, "kibana", "elasticsearch")
		kibanaDeployment := utils.Deployment("kibana", cluster.Namespace, "kibana", "kibana", kibanaPodSpec)

		utils.AddOwnerRefToObject(kibanaDeployment, utils.AsOwner(cluster))

		err = sdk.Create(kibanaDeployment)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana deployment: %v", err)
		}

		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = updateKibanaIfRequired(kibanaDeployment); err != nil {
				return
			}
		}

	} else {
		kibanaPodSpec := getKibanaPodSpec(cluster, "kibana-app", "elasticsearch-app")
		kibanaDeployment := utils.Deployment("kibana-app", cluster.Namespace, "kibana", "kibana", kibanaPodSpec)

		utils.AddOwnerRefToObject(kibanaDeployment, utils.AsOwner(cluster))

		err = sdk.Create(kibanaDeployment)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana App deployment: %v", err)
		}

		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = updateKibanaIfRequired(kibanaDeployment); err != nil {
				return
			}
		}

		kibanaInfraPodSpec := getKibanaPodSpec(cluster, "kibana-infra", "elasticsearch-infra")
		kibanaInfraDeployment := utils.Deployment("kibana-infra", cluster.Namespace, "kibana", "kibana", kibanaInfraPodSpec)

		utils.AddOwnerRefToObject(kibanaInfraDeployment, utils.AsOwner(cluster))

		err = sdk.Create(kibanaInfraDeployment)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana Infra deployment: %v", err)
		}

		if cluster.Spec.ManagementState == logging.ManagementStateManaged {
			if err = updateKibanaIfRequired(kibanaInfraDeployment); err != nil {
				return
			}
		}
	}

	return nil
}

func createOrUpdateOauthClient(logging *logging.ClusterLogging, oauthSecret string) (err error) {

	clusterDomain, err := utils.GetClusterDNSDomain()
	if err != nil {
		return nil
	}

	oauthClient := &oauth.OAuthClient{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OAuthClient",
			APIVersion: oauth.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kibana-proxy",
			Labels: map[string]string{
				"logging-infra": "support",
				"namespace":     logging.Namespace,
			},
		},
		Secret: oauthSecret,
		RedirectURIs: []string{
			fmt.Sprintf("%s.%s", "https://kibana", clusterDomain),
			fmt.Sprintf("%s.%s", "https://kibana-app", clusterDomain),
			fmt.Sprintf("%s.%s", "https://kibana-infra", clusterDomain),
		},
		ScopeRestrictions: []oauth.ScopeRestriction{
			oauth.ScopeRestriction{
				ExactValues: []string{
					"user:info",
					"user:check-access",
					"user:list-projects",
				},
			},
		},
	}

	err = sdk.Create(oauthClient)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating OAuthClient: %v", err)
	}

	return nil
}

func createOrUpdateKibanaRoute(logging *logging.ClusterLogging) error {

	if utils.AllInOne(logging) {
		kibanaRoute := utils.Route(
			"kibana",
			logging.Namespace,
			"kibana",
			"kibana",
			"/tmp/_working_dir/ca.crt",
		)

		utils.AddOwnerRefToObject(kibanaRoute, utils.AsOwner(logging))

		err := sdk.Create(kibanaRoute)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana route: %v", err)
		}

		sharedConfig := createSharedConfig(logging, "https://kibana", "https://kibana")
		utils.AddOwnerRefToObject(sharedConfig, utils.AsOwner(logging))

		err = sdk.Create(sharedConfig)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana route shared config: %v", err)
		}
	} else {
		kibanaRoute := utils.Route(
			"kibana-app",
			logging.Namespace,
			"kibana-app",
			"kibana-app",
			"/tmp/_working_dir/ca.crt",
		)

		utils.AddOwnerRefToObject(kibanaRoute, utils.AsOwner(logging))

		err := sdk.Create(kibanaRoute)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana App route: %v", err)
		}

		kibanaInfraRoute := utils.Route(
			"kibana-infra",
			logging.Namespace,
			"kibana-infra",
			"kibana-infra",
			"/tmp/_working_dir/ca.crt",
		)

		utils.AddOwnerRefToObject(kibanaInfraRoute, utils.AsOwner(logging))

		err = sdk.Create(kibanaInfraRoute)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana Infra route: %v", err)
		}

		sharedConfig := createSharedConfig(logging, "https://kibana-app", "https://kibana-infra")
		utils.AddOwnerRefToObject(sharedConfig, utils.AsOwner(logging))

		err = sdk.Create(sharedConfig)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating Kibana route shared config: %v", err)
		}
	}

	return nil
}

func createOrUpdateKibanaService(logging *logging.ClusterLogging) error {

	if utils.AllInOne(logging) {
		kibanaService := utils.Service(
			"kibana",
			logging.Namespace,
			"kibana",
			[]v1.ServicePort{
				{Port: 443, TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "oaproxy",
				}},
			})

		utils.AddOwnerRefToObject(kibanaService, utils.AsOwner(logging))

		err := sdk.Create(kibanaService)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Kibana service: %v", err)
		}
	} else {
		kibanaService := utils.Service(
			"kibana-app",
			logging.Namespace,
			"kibana",
			[]v1.ServicePort{
				{Port: 443, TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "oaproxy",
				}},
			})

		utils.AddOwnerRefToObject(kibanaService, utils.AsOwner(logging))

		err := sdk.Create(kibanaService)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Kibana App service: %v", err)
		}

		kibanaInfraService := utils.Service(
			"kibana-infra",
			logging.Namespace,
			"kibana",
			[]v1.ServicePort{
				{Port: 443, TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "oaproxy",
				}},
			})

		utils.AddOwnerRefToObject(kibanaInfraService, utils.AsOwner(logging))

		err = sdk.Create(kibanaInfraService)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Kibana Infra service: %v", err)
		}
	}

	return nil
}

func createOrUpdateKibanaSecret(logging *logging.ClusterLogging, oauthSecret []byte) error {

	kibanaSecret := utils.Secret(
		"kibana",
		logging.Namespace,
		map[string][]byte{
			"ca":   utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"key":  utils.GetFileContents("/tmp/_working_dir/system.logging.kibana.key"),
			"cert": utils.GetFileContents("/tmp/_working_dir/system.logging.kibana.crt"),
		})

	utils.AddOwnerRefToObject(kibanaSecret, utils.AsOwner(logging))

	err := sdk.Create(kibanaSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana secret: %v", err)
	}

	proxySecret := utils.Secret(
		"kibana-proxy",
		logging.Namespace,
		map[string][]byte{
			"oauth-secret":   oauthSecret,
			"session-secret": utils.GetRandomWord(32),
			"server-key":     utils.GetFileContents("/tmp/_working_dir/kibana-internal.key"),
			"server-cert":    utils.GetFileContents("/tmp/_working_dir/kibana-internal.crt"),
		})

	utils.AddOwnerRefToObject(proxySecret, utils.AsOwner(logging))

	err = sdk.Create(proxySecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana Proxy secret: %v", err)
	}

	return nil
}

func getKibanaPodSpec(logging *logging.ClusterLogging, kibanaName string, elasticsearchName string) v1.PodSpec {

	kibanaContainer := utils.Container("kibana", v1.PullIfNotPresent, logging.Spec.Visualization.KibanaSpec.Resources)

	var endpoint bytes.Buffer

	endpoint.WriteString("https://")
	endpoint.WriteString(elasticsearchName)
	endpoint.WriteString(":9200")

	kibanaContainer.Env = []v1.EnvVar{
		{Name: "ELASTICSEARCH_URL", Value: endpoint.String()},
		{Name: "KIBANA_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "kibana", Resource: "limits.memory"}}},
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

	kibanaProxyContainer := utils.Container("kibana-proxy", v1.PullIfNotPresent, logging.Spec.Visualization.KibanaSpec.ProxySpec.Resources)

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
		{Name: "OCP_AUTH_PROXY_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "kibana-proxy", Resource: "limits.memory"}}},
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
			{Name: "kibana", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "kibana"}}},
			{Name: "kibana-proxy", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "kibana-proxy"}}},
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
								{Key: "logging-infra", Operator: metav1.LabelSelectorOpIn, Values: []string{"kibana"}},
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
		return fmt.Errorf("Failed to get Kibana deployment: %v", err)
	}

	current, different := isKibanaDifferent(current, desired)

	if different {
		if err = sdk.Update(current); err != nil {
			return fmt.Errorf("Failed to update Kibana deployment: %v", err)
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

func updateCurrentImages(current *apps.Deployment, desired *apps.Deployment) (*apps.Deployment) {

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

func createSharedConfig(logging *logging.ClusterLogging, kibanaAppURL, kibanaInfraURL string) *v1.ConfigMap {
	return utils.ConfigMap(
		"sharing-config",
		logging.Namespace,
		map[string]string{
			"kibanaAppURL":   kibanaAppURL,
			"kibanaInfraURL": kibanaInfraURL,
		},
	)
}
