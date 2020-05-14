package kibana

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/types"

	configv1 "github.com/openshift/api/config/v1"
	kibana "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/openshift/elasticsearch-operator/pkg/constants"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kibanaServiceAccountName     = "kibana"
	kibanaOAuthRedirectReference = "{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"kibana\"}}"
)

var (
	kibanaServiceAccountAnnotations = map[string]string{
		"serviceaccounts.openshift.io/oauth-redirectreference.first": kibanaOAuthRedirectReference,
	}
)

func ReconcileKibana(requestCluster *kibana.Kibana, requestClient client.Client, proxyConfig *configv1.Proxy) error {
	clusterKibanaRequest := KibanaRequest{
		client:  requestClient,
		cluster: requestCluster,
	}

	if clusterKibanaRequest.cluster == nil {
		return nil
	}

	// ensure that we have the certs pulled in from the secret first... required for route generation
	if err := clusterKibanaRequest.readSecrets(); err != nil {
		return err
	}

	if err := clusterKibanaRequest.CreateOrUpdateServiceAccount(kibanaServiceAccountName, &kibanaServiceAccountAnnotations); err != nil {
		return err
	}

	if err := clusterKibanaRequest.createOrUpdateKibanaService(); err != nil {
		return err
	}

	if err := clusterKibanaRequest.createOrUpdateKibanaRoute(); err != nil {
		return err
	}

	if err := clusterKibanaRequest.createOrUpdateKibanaConsoleExternalLogLink(); err != nil {
		return err
	}

	if err := clusterKibanaRequest.createOrUpdateKibanaConsoleLinks(); err != nil {
		return err
	}

	if err := clusterKibanaRequest.createOrUpdateKibanaDeployment(proxyConfig); err != nil {
		return err
	}

	if err := clusterKibanaRequest.removeSharedConfigMapPre45x(); err != nil {
		return err
	}

	kibanaStatus, err := clusterKibanaRequest.getKibanaStatus()
	cluster := clusterKibanaRequest.cluster

	if err != nil {
		return fmt.Errorf("Failed to get Kibana status for %q: %v", cluster.Name, err)
	}

	printUpdateMessage := true
	retryErr := retry.RetryOnConflict(retry.DefaultRetry,
		func() error {
			if !compareKibanaStatus(kibanaStatus,
				cluster.Status) {
				if printUpdateMessage {
					logrus.Infof("Updating status of Kibana")
					printUpdateMessage = false
				}
				cluster.Status = kibanaStatus
				return clusterKibanaRequest.UpdateStatus(cluster)
			}
			return nil
		})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Kibana status for %q: %v", cluster.Name, retryErr)
	}
	logrus.Infof("Kibana status successfully updated")

	return nil
}

func ReconcileKibanaInstance(request reconcile.Request, rClient client.Client) error {
	kibanaInstance := &kibana.Kibana{}
	key := types.NamespacedName{
		Namespace: request.Namespace,
		Name:      constants.KibanaInstanceName,
	}

	err := rClient.Get(context.TODO(), key, kibanaInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if kibanaInstance.Spec.ManagementState == kibana.ManagementStateUnmanaged {
		return nil
	}

	proxyCfg, err := getProxyConfig(rClient)
	if err != nil {
		return err
	}

	if err := ReconcileKibana(kibanaInstance, rClient, proxyCfg); err != nil {
		return err
	}

	return nil
}

func getProxyConfig(r client.Client) (*configv1.Proxy, error) {
	proxyNamespacedName := types.NamespacedName{Name: constants.ProxyName}
	proxyConfig := &configv1.Proxy{}
	if err := r.Get(context.TODO(), proxyNamespacedName, proxyConfig); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("Encountered unexpected error getting %v. Error: %s\r\n", proxyNamespacedName,
				err.Error())
		}
	}
	return proxyConfig, nil
}

func compareKibanaStatus(lhs, rhs []kibana.KibanaStatus) bool {
	// there should only ever be a single kibana status object
	if len(lhs) != len(rhs) {
		return false
	}

	if len(lhs) > 0 {
		for index := range lhs {
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

func (clusterRequest *KibanaRequest) createOrUpdateKibanaDeployment(proxyConfig *configv1.Proxy) (err error) {
	kibanaTrustBundle := &v1.ConfigMap{}

	// Create cluster proxy trusted CA bundle.
	if proxyConfig != nil {
		err = clusterRequest.createOrUpdateTrustedCABundleConfigMap(constants.KibanaTrustedCAName)
		if err != nil {
			return
		}
	}

	kibanaPodSpec := newKibanaPodSpec(clusterRequest,
		fmt.Sprintf("elasticsearch.%s.svc.cluster.local", clusterRequest.cluster.Namespace),
		proxyConfig,
		kibanaTrustBundle)

	kibanaDeployment := NewDeployment(
		"kibana",
		clusterRequest.cluster.Namespace,
		"kibana",
		"kibana",
		kibanaPodSpec,
	)
	kibanaDeployment.Spec.Replicas = &clusterRequest.cluster.Spec.Replicas

	// if we don't have the hash values we shouldn't start/create
	annotations, err := clusterRequest.getKibanaAnnotations(kibanaDeployment)
	if err != nil {
		return err
	}

	kibanaDeployment.Spec.Template.ObjectMeta.Annotations = annotations

	utils.AddOwnerRefToObject(kibanaDeployment, getOwnerRef(clusterRequest.cluster))

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

			if current.Spec.Template.ObjectMeta.Annotations[constants.TrustedCABundleHashName] != kibanaDeployment.Spec.Template.ObjectMeta.Annotations[constants.TrustedCABundleHashName] {
				different = true
			}

			for _, secretName := range []string{"kibana", "kibana-proxy"} {
				hashKey := fmt.Sprintf("%s%s", constants.SecretHashPrefix, secretName)
				if current.Spec.Template.ObjectMeta.Annotations[hashKey] != kibanaDeployment.Spec.Template.ObjectMeta.Annotations[hashKey] {
					different = true
				}
			}

			if different {
				return clusterRequest.Update(current)
			}
			return nil
		})
	}

	return nil
}

func (clusterRequest *KibanaRequest) getKibanaAnnotations(deployment *apps.Deployment) (map[string]string, error) {

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	annotations := deployment.Spec.Template.ObjectMeta.Annotations

	kibanaTrustBundle := &v1.ConfigMap{}
	kibanaTrustBundleName := types.NamespacedName{Name: constants.KibanaTrustedCAName, Namespace: clusterRequest.cluster.Namespace}
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

	//generate secret hash
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

	if updateCurrentDeploymentEnvIfDifferent(current, desired) {
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

func updateCurrentDeploymentEnvIfDifferent(current *apps.Deployment, desired *apps.Deployment) bool {

	different := false

	containers := current.Spec.Template.Spec.Containers

	for index, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the env of containers with the same name
			if curr.Name == des.Name {
				if !utils.EnvValueEqual(curr.Env, des.Env) {
					containers[index].Env = des.Env
					different = true
				}
			}
		}
	}

	return different
}

func (clusterRequest *KibanaRequest) createOrUpdateKibanaService() error {

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

	utils.AddOwnerRefToObject(kibanaService, getOwnerRef(clusterRequest.cluster))

	err := clusterRequest.Create(kibanaService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Kibana service for %q: %v", clusterRequest.cluster.Name, err)
	}

	return nil
}

func getImage() string {
	return utils.LookupEnvWithDefault("KIBANA_IMAGE", kibanaDefaultImage)
}

func getProxyImage() string {
	return utils.LookupEnvWithDefault("PROXY_IMAGE", kibanaProxyDefaultImage)
}

func newKibanaPodSpec(cluster *KibanaRequest, elasticsearchName string, proxyConfig *configv1.Proxy,
	trustedCABundleCM *v1.ConfigMap) v1.PodSpec {
	visSpec := kibana.KibanaSpec{}
	if cluster.cluster != nil {
		visSpec = cluster.cluster.Spec
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

	kibanaImage := getImage()
	kibanaContainer := NewContainer(
		"kibana",
		kibanaImage,
		v1.PullIfNotPresent,
		*kibanaResources,
	)

	endpoints := fmt.Sprintf(`["https://%s:9200"]`, elasticsearchName)

	kibanaContainer.Env = []v1.EnvVar{
		{
			Name:  "ELASTICSEARCH_HOSTS",
			Value: endpoints,
		},
		{
			Name: "KIBANA_MEMORY_LIMIT",
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

	proxyImage := getProxyImage()
	kibanaProxyContainer := NewContainer(
		"kibana-proxy",
		proxyImage,
		v1.PullIfNotPresent,
		*kibanaProxyResources,
	)

	kibanaProxyContainer.Args = []string{
		"--upstream-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		"--https-address=:3000",
		"-provider=openshift",
		fmt.Sprintf("-client-id=system:serviceaccount:%s:kibana", cluster.cluster.Namespace),
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

func getOwnerRef(v *kibana.Kibana) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: kibana.SchemeGroupVersion.String(),
		Kind:       "Kibana",
		Name:       v.Name,
		UID:        v.UID,
		Controller: &trueVar,
	}
}
