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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"

	oauth "github.com/openshift/api/oauth/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cluster *ClusterLogging) CreateOrUpdateVisualization(stack *logging.StackSpec) (err error) {

	if stack.Visualization == nil {
		logrus.Debugf("Visualization is not spec'd for stack '%s'", stack.Name)
		return
	}

	if err = createOrUpdateKibanaServiceAccount(cluster); err != nil {
		return
	}

	if err = createOrUpdateKibanaService(cluster, stack); err != nil {
		return
	}

	if err = createOrUpdateKibanaRoute(cluster, stack); err != nil {
		return
	}

	if err = createOrUpdateKibanaDeployment(cluster, stack); err != nil {
		return
	}

	kibanaStatus, err := getKibanaStatus(cluster.ClusterLogging.Namespace)

	if err != nil {
		return fmt.Errorf("Failed to get status for Kibana: %v", err)
	}

	printUpdateMessage := true
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if exists := utils.DoesClusterLoggingExist(cluster.ClusterLogging); exists {
			if !reflect.DeepEqual(kibanaStatus, cluster.ClusterLogging.Status.Visualization.KibanaStatus) {
				if printUpdateMessage {
					logrus.Info("Updating status of Kibana")
					printUpdateMessage = false
				}
				cluster.ClusterLogging.Status.Visualization.KibanaStatus = kibanaStatus
				return sdk.Update(cluster.ClusterLogging)
			}
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging Kibana status: %v", retryErr)
	}

	return nil
}

func createOrUpdateKibanaServiceAccount(logging *ClusterLogging) error {

	kibanaServiceAccount := utils.NewServiceAccount(Kibana, logging.Namespace)

	logging.addOwnerRefTo(kibanaServiceAccount)

	err := sdk.Create(kibanaServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana service account: %v", err)
	}

	return nil
}

func createOrUpdateKibanaDeployment(cluster *ClusterLogging, stack *logging.StackSpec) (err error) {

	name := cluster.getKibanaName(stack.Name)
	esName := cluster.getElasticsearchName(stack.Name)
	kibanaPodSpec := getKibanaPodSpec(cluster, stack, name, esName)
	kibanaDeployment := utils.NewDeployment(name, cluster.Namespace, Kibana, name, kibanaPodSpec)

	cluster.addOwnerRefTo(kibanaDeployment)

	err = sdk.Create(kibanaDeployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana deployment '%s': %v", name, err)
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

func CreateOrUpdateOauthClient(cluster *ClusterLogging, oauthSecret string) (err error) {

	redirectURIs := sets.NewString()
	for _, stack := range cluster.Spec.Stacks {
		name := cluster.getKibanaName(stack.Name)
		kibanaURL, err := utils.GetRouteURL(name, cluster.Namespace)
		if err != nil {
			return err
		}
		redirectURIs.Insert(kibanaURL)
	}

	oauthClient := &oauth.OAuthClient{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OAuthClient",
			APIVersion: oauth.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: KibanaProxy,
			Labels: map[string]string{
				"logging-infra": "support",
				"namespace":     cluster.Namespace,
			},
		},
		Secret:       oauthSecret,
		RedirectURIs: redirectURIs.List(),
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

	cluster.addOwnerRefTo(oauthClient)

	err = sdk.Create(oauthClient)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating OAuthClient: %v", err)
	}

	return nil
}

func createOrUpdateKibanaRoute(logging *ClusterLogging, stack *logging.StackSpec) error {

	name := logging.getKibanaName(stack.Name)
	kibanaRoute := utils.NewRoute(
		name,
		logging.Namespace,
		name,
		"/tmp/_working_dir/ca.crt",
	)

	logging.addOwnerRefTo(kibanaRoute)

	err := sdk.Create(kibanaRoute)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route '%s': %v", name, err)
	}
	return nil
}

func createOrUpdateKibanaService(logging *ClusterLogging, stack *logging.StackSpec) error {

	name := logging.getKibanaName(stack.Name)
	kibanaService := utils.NewService(
		name,
		logging.Namespace,
		name,
		[]v1.ServicePort{
			{Port: 443, TargetPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "oaproxy",
			}},
		})
	logging.addOwnerRefTo(kibanaService)

	err := sdk.Create(kibanaService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana service: %v", err)
	}

	return nil
}

func CreateOrUpdateKibanaSecret(logging *ClusterLogging, oauthSecret []byte) error {
	kibanaSecret := utils.NewSecret(
		Kibana,
		logging.Namespace,
		map[string][]byte{
			"ca":   utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"key":  utils.GetFileContents("/tmp/_working_dir/system.logging.kibana.key"),
			"cert": utils.GetFileContents("/tmp/_working_dir/system.logging.kibana.crt"),
		})

	logging.addOwnerRefTo(kibanaSecret)

	err := sdk.Create(kibanaSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana secret: %v", err)
	}

	proxySecret := utils.NewSecret(
		KibanaProxy,
		logging.Namespace,
		map[string][]byte{
			"oauth-secret":   oauthSecret,
			"session-secret": utils.GetRandomWord(32),
			"server-key":     utils.GetFileContents("/tmp/_working_dir/kibana-internal.key"),
			"server-cert":    utils.GetFileContents("/tmp/_working_dir/kibana-internal.crt"),
		})

	logging.addOwnerRefTo(proxySecret)

	err = sdk.Create(proxySecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana Proxy secret: %v", err)
	}

	return nil
}

func getKibanaPodSpec(logging *ClusterLogging, stack *logging.StackSpec, kibanaName string, elasticsearchName string) v1.PodSpec {

	kibanaContainer := utils.NewContainer(Kibana, v1.PullIfNotPresent, stack.Visualization.Resources)

	var endpoint bytes.Buffer

	endpoint.WriteString("https://")
	endpoint.WriteString(elasticsearchName)
	endpoint.WriteString(":9200")

	kibanaContainer.Env = []v1.EnvVar{
		{Name: "ELASTICSEARCH_URL", Value: endpoint.String()},
		{Name: "Kibana_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "kibana", Resource: "limits.memory"}}},
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

	kibanaProxyContainer := utils.NewContainer(KibanaProxy, v1.PullIfNotPresent, stack.Visualization.ProxySpec.Resources)

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
		{Name: KibanaProxy, ReadOnly: true, MountPath: "/secret"},
	}

	kibanaPodSpec := utils.NewPodSpec(
		Kibana,
		[]v1.Container{kibanaContainer, kibanaProxyContainer},
		[]v1.Volume{
			{Name: Kibana, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "kibana"}}},
			{Name: KibanaProxy, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "kibana-proxy"}}},
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
		if apierrors.IsNotFound(err) {
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

func CreateSharedConfig(cluster *ClusterLogging) error {
	urls := make(map[string]string)
	names := cluster.getStackNames()
	for _, name := range names.List() {
		url, err := utils.GetRouteURL(cluster.getKibanaName(name), cluster.Namespace)
		if err != nil {
			return err
		}
		if name == APP {
			urls[APP] = url
			urls[INFRA] = url
		}
		if name == INFRA {
			urls[INFRA] = url
		}
	}
	config := utils.NewConfigMap(
		"sharing-config",
		cluster.Namespace,
		map[string]string{
			"kibanaAppURL":   urls[APP],
			"kibanaInfraURL": urls[INFRA],
		},
	)
	cluster.addOwnerRefTo(config)
	err := sdk.Create(config)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Kibana route shared config: %v", err)
	}
	return nil
}
