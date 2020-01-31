package k8shandler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	configv1 "github.com/openshift/api/config/v1"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	fluentdAlertsFile = "fluentd/fluentd_prometheus_alerts.yaml"
	fluentdName       = "fluentd"
)

func (clusterRequest *ClusterLoggingRequest) removeFluentd() (err error) {
	if clusterRequest.isManaged() {

		if err = clusterRequest.RemoveService(fluentdName); err != nil {
			return
		}

		if err = clusterRequest.RemoveServiceMonitor(fluentdName); err != nil {
			return
		}

		if err = clusterRequest.RemovePrometheusRule(fluentdName); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(fluentdName); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(constants.FluentdTrustedCAName); err != nil {
			return
		}

		if err = clusterRequest.RemoveDaemonset(fluentdName); err != nil {
			return
		}

		// Wait longer than the terminationGracePeriodSeconds
		time.Sleep(12 * time.Second)

		if err = clusterRequest.RemoveSecret(fluentdName); err != nil {
			return
		}

	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdService() error {
	service := NewService(
		fluentdName,
		clusterRequest.cluster.Namespace,
		fluentdName,
		[]v1.ServicePort{
			{
				Port:       metricsPort,
				TargetPort: intstr.FromString(metricsPortName),
				Name:       metricsPortName,
			},
		},
	)

	service.Annotations = map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": "fluentd-metrics",
	}

	utils.AddOwnerRefToObject(service, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd service: %v", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdServiceMonitor() error {

	cluster := clusterRequest.cluster

	serviceMonitor := NewServiceMonitor(fluentdName, cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   metricsPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", fluentdName, cluster.Namespace),
			// ServerName can be e.g. fluentd.openshift-logging.svc
		},
	}

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"logging-infra": "support",
		},
	}

	serviceMonitor.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-fluentd",
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{cluster.Namespace},
		},
	}

	utils.AddOwnerRefToObject(serviceMonitor, utils.AsOwner(cluster))

	err := clusterRequest.Create(serviceMonitor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd ServiceMonitor: %v", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdPrometheusRule() error {

	cluster := clusterRequest.cluster

	promRule := NewPrometheusRule(fluentdName, cluster.Namespace)

	promRuleSpec, err := NewPrometheusRuleSpecFrom(utils.GetShareDir() + "/" + fluentdAlertsFile)
	if err != nil {
		return fmt.Errorf("Failure creating the fluentd PrometheusRule: %v", err)
	}

	promRule.Spec = *promRuleSpec

	utils.AddOwnerRefToObject(promRule, utils.AsOwner(cluster))

	err = clusterRequest.Create(promRule)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd PrometheusRule: %v", err)
	}

	return nil
}

//createOrUpdateFluentdLegacyForwardConfigMap to address Bug 1782566. To be removed for LogForwarding GA
func (clusterRequest *ClusterLoggingRequest) includeLegacyForwardConfig() bool {
	config := &v1.ConfigMap{
		Data: map[string]string{},
	}
	if err := clusterRequest.Get("secure-forward", config); err != nil {
		if errors.IsNotFound(err) {
			return false
		}
		logger.Warnf("There was a non-critical error trying to fetch the secure-forward configmap: %v", err)
	}
	_, found := config.Data["secure-forward.conf"]
	return found
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdConfigMap(fluentConf string) error {
	logrus.Debug("createOrUpdateFluentdConfigMap...")
	fluentdConfigMap := NewConfigMap(
		fluentdName,
		clusterRequest.cluster.Namespace,
		map[string]string{
			"fluent.conf":          fluentConf,
			"throttle-config.yaml": string(utils.GetFileContents(utils.GetShareDir() + "/fluentd/fluentd-throttle-config.yaml")),
			"run.sh":               string(utils.GetFileContents(utils.GetShareDir() + "/fluentd/run.sh")),
		},
	)

	utils.AddOwnerRefToObject(fluentdConfigMap, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(fluentdConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Fluentd configmap: %v", err)
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &v1.ConfigMap{}
		if err = clusterRequest.Get(fluentdConfigMap.Name, current); err != nil {
			if errors.IsNotFound(err) {
				logrus.Debugf("Returning nil. The configmap %q was not found even though create previously failed.  Was it culled?", fluentdConfigMap.Name)
				return nil
			}
			return fmt.Errorf("Failed to get %v configmap for %q: %v", fluentdConfigMap.Name, clusterRequest.cluster.Name, err)
		}
		if reflect.DeepEqual(fluentdConfigMap.Data, current.Data) {
			return nil
		}
		current.Data = fluentdConfigMap.Data
		return clusterRequest.Update(current)
	})

	return retryErr
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdSecret() error {

	fluentdSecret := NewSecret(
		fluentdName,
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"ca-bundle.crt": utils.GetWorkingDirFileContents("ca.crt"),
			"tls.key":       utils.GetWorkingDirFileContents("system.logging.fluentd.key"),
			"tls.crt":       utils.GetWorkingDirFileContents("system.logging.fluentd.crt"),
		})

	utils.AddOwnerRefToObject(fluentdSecret, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.CreateOrUpdateSecret(fluentdSecret)
	if err != nil {
		return err
	}

	return nil
}

func newFluentdPodSpec(cluster *logging.ClusterLogging, elasticsearchAppName string, elasticsearchInfraName string, proxyConfig *configv1.Proxy, trustedCABundleCM *v1.ConfigMap, pipelineSpec logforward.ForwardingSpec) v1.PodSpec {
	collectionSpec := logging.CollectionSpec{}
	if cluster.Spec.Collection != nil {
		collectionSpec = *cluster.Spec.Collection
	}
	var resources = collectionSpec.Logs.FluentdSpec.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultFluentdMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultFluentdMemory,
				v1.ResourceCPU:    defaultFluentdCpuRequest,
			},
		}
	}
	fluentdContainer := NewContainer("fluentd", "fluentd", v1.PullIfNotPresent, *resources)

	fluentdContainer.Ports = []v1.ContainerPort{
		v1.ContainerPort{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	fluentdContainer.Env = []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
		{Name: "METRICS_CERT", Value: "/etc/fluent/metrics/tls.crt"},
		{Name: "METRICS_KEY", Value: "/etc/fluent/metrics/tls.key"},
		{Name: "BUFFER_QUEUE_LIMIT", Value: "32"},
		{Name: "BUFFER_SIZE_LIMIT", Value: "8m"},
		{Name: "FILE_BUFFER_LIMIT", Value: "256Mi"},
		{Name: "FLUENTD_CPU_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.cpu"}}},
		{Name: "FLUENTD_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.memory"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
		{Name: "CDM_KEEP_EMPTY_FIELDS", Value: "message"}, // by default, keep empty messages
	}

	proxyEnv := utils.SetProxyEnvVars(proxyConfig)
	fluentdContainer.Env = append(fluentdContainer.Env, proxyEnv...)

	fluentdContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "runlogjournal", MountPath: "/run/log/journal"},
		{Name: "varlog", MountPath: "/var/log"},
		{Name: "varlibdockercontainers", ReadOnly: true, MountPath: "/var/lib/docker"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/fluent/configs.d/user"},
		{Name: "secureforwardconfig", ReadOnly: true, MountPath: "/etc/fluent/configs.d/secure-forward"},
		{Name: "secureforwardcerts", ReadOnly: true, MountPath: "/etc/ocp-forward"},
		{Name: "entrypoint", ReadOnly: true, MountPath: "/opt/app-root/src/run.sh", SubPath: "run.sh"},
		{Name: "certs", ReadOnly: true, MountPath: "/etc/fluent/keys"},
		{Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
		{Name: "dockercfg", ReadOnly: true, MountPath: "/etc/sysconfig/docker"},
		{Name: "dockerdaemoncfg", ReadOnly: true, MountPath: "/etc/docker"},
		{Name: "filebufferstorage", MountPath: "/var/lib/fluentd"},
		{Name: metricsVolumeName, MountPath: "/etc/fluent/metrics"},
	}
	for _, target := range pipelineSpec.Outputs {
		if target.Secret != nil {
			path := fmt.Sprintf("/var/run/ocp-collector/secrets/%s", target.Secret.Name)
			fluentdContainer.VolumeMounts = append(fluentdContainer.VolumeMounts, v1.VolumeMount{Name: target.Name, MountPath: path})
		}
	}

	addTrustedCAVolume := false
	// If trusted CA bundle ConfigMap exists and its hash value is non-zero, mount the bundle.
	if trustedCABundleCM != nil && hasTrustedCABundle(trustedCABundleCM) {
		addTrustedCAVolume = true
		fluentdContainer.VolumeMounts = append(fluentdContainer.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.FluentdTrustedCAName,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})
	}

	fluentdContainer.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}

	tolerations := utils.AppendTolerations(
		collectionSpec.Logs.FluentdSpec.Tolerations,
		[]v1.Toleration{
			v1.Toleration{
				Key:      "node-role.kubernetes.io/master",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
			v1.Toleration{
				Key:      "node.kubernetes.io/disk-pressure",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	)

	fluentdPodSpec := NewPodSpec(
		"logcollector",
		[]v1.Container{fluentdContainer},
		[]v1.Volume{
			{Name: "runlogjournal", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/run/log/journal"}}},
			{Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
			{Name: "varlibdockercontainers", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker"}}},
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "fluentd"}}}},
			{Name: "secureforwardconfig", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "secure-forward"}, Optional: utils.GetBool(true)}}},
			{Name: "secureforwardcerts", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "secure-forward", Optional: utils.GetBool(true)}}},
			{Name: "entrypoint", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "fluentd"}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "fluentd"}}},
			{Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
			{Name: "dockercfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/sysconfig/docker"}}},
			{Name: "dockerdaemoncfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/docker"}}},
			{Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/fluentd"}}},
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "fluentd-metrics"}}},
		},
		collectionSpec.Logs.FluentdSpec.NodeSelector,
		tolerations,
	)
	for _, target := range pipelineSpec.Outputs {
		if target.Secret != nil {
			fluentdPodSpec.Volumes = append(fluentdPodSpec.Volumes, v1.Volume{Name: target.Name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: target.Secret.Name}}})
		}
	}

	if addTrustedCAVolume {
		optional := true
		fluentdPodSpec.Volumes = append(fluentdPodSpec.Volumes,
			v1.Volume{
				Name: constants.FluentdTrustedCAName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: constants.FluentdTrustedCAName,
						},
						Optional: &optional,
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

	fluentdPodSpec.PriorityClassName = clusterLoggingPriorityClassName
	// Shorten the termination grace period from the default 30 sec to 10 sec.
	fluentdPodSpec.TerminationGracePeriodSeconds = utils.GetInt64(10)

	return fluentdPodSpec
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdDaemonset(pipelineConfHash string, proxyConfig *configv1.Proxy) (err error) {

	cluster := clusterRequest.cluster

	fluentdTrustBundle := &v1.ConfigMap{}
	if proxyConfig != nil {
		// Create or update cluster proxy trusted CA bundle.
		err = clusterRequest.createOrUpdateTrustedCABundleConfigMap(constants.FluentdTrustedCAName)
		if err != nil {
			return
		}

		// fluentd-trusted-ca-bundle
		fluentdTrustBundleName := types.NamespacedName{Name: constants.FluentdTrustedCAName, Namespace: constants.OpenshiftNS}
		if err := clusterRequest.client.Get(context.TODO(), fluentdTrustBundleName, fluentdTrustBundle); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}
	}

	fluentdPodSpec := newFluentdPodSpec(cluster, "elasticsearch", "elasticsearch", proxyConfig, fluentdTrustBundle, clusterRequest.ForwardingSpec)

	fluentdDaemonset := NewDaemonSet("fluentd", cluster.Namespace, "fluentd", "fluentd", fluentdPodSpec)
	fluentdDaemonset.Spec.Template.Spec.Containers[0].Env = updateEnvVar(v1.EnvVar{Name: "FLUENT_CONF_HASH", Value: pipelineConfHash}, fluentdDaemonset.Spec.Template.Spec.Containers[0].Env)

	uid := getServiceAccountLogCollectorUID()
	if len(uid) == 0 {
		// There's no uid for logcollector serviceaccount; setting ClusterLogging for the ownerReference.
		utils.AddOwnerRefToObject(fluentdDaemonset, utils.AsOwner(cluster))
	} else {
		// There's a uid for logcollector serviceaccount; setting the ServiceAccount for the ownerReference with blockOwnerDeletion.
		utils.AddOwnerRefToObject(fluentdDaemonset, NewLogCollectorServiceAccountRef(uid))
	}

	err = clusterRequest.Create(fluentdDaemonset)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Fluentd Daemonset %v", err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateFluentdDaemonsetIfRequired(fluentdDaemonset, fluentdTrustBundle)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) updateFluentdDaemonsetIfRequired(desired *apps.DaemonSet, trustedCABundleCM *v1.ConfigMap) (err error) {
	logger.DebugObject("desired fluent update: %v", desired)
	current := &apps.DaemonSet{}

	if err = clusterRequest.Get(desired.Name, current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Fluentd daemonset: %v", err)
	}

	flushBuffer := isBufferFlushRequired(current, desired)
	desired, different := isDaemonsetDifferent(current, desired)

	// Check trustedCA certs have been updated or not by comparing the hash values in annotation.
	newTrustedCAHashedValue, err := calcTrustedCAHashValue(trustedCABundleCM)
	if err != nil {
		return fmt.Errorf("unable to calculate trusted CA hash value. E: %s", err.Error())
	}

	trustedCAHashedValue, _ := current.Spec.Template.ObjectMeta.Annotations[constants.TrustedCABundleHashName]
	if trustedCAHashedValue != newTrustedCAHashedValue {
		different = true
		if desired.Spec.Template.ObjectMeta.Annotations == nil {
			desired.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
		}
		desired.Spec.Template.ObjectMeta.Annotations[constants.TrustedCABundleHashName] = newTrustedCAHashedValue
	}

	if different {
		current.Spec = desired.Spec
		if flushBuffer {
			current.Spec.Template.Spec.Containers[0].Env = updateEnvVar(v1.EnvVar{Name: "FLUSH_AT_SHUTDOWN", Value: "True"}, current.Spec.Template.Spec.Containers[0].Env)
			if err = clusterRequest.Update(current); err != nil {
				logrus.Debugf("Failed to prepare Fluentd daemonset to flush its buffers: %v", err)
				return err
			}

			// wait for pods to all restart then continue
			if err = clusterRequest.waitForDaemonSetReady(current); err != nil {
				return fmt.Errorf("Timed out waiting for Fluentd to be ready")
			}
		}
		logger.DebugObject("updating fluentd to: %v", current)
		if err = clusterRequest.Update(desired); err != nil {
			return err
		}
	}

	return nil
}

//updateEnvar adds the value to the list or replaces it if it already existing
func updateEnvVar(value v1.EnvVar, values []v1.EnvVar) []v1.EnvVar {
	found := false
	for i, envvar := range values {
		if envvar.Name == value.Name {
			values[i] = value
			found = true
		}
	}
	if !found {
		values = append(values, value)
	}
	return values
}
