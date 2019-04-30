package k8shandler

import (
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	fluentdAlertsFile = "/usr/share/logging/fluentd/fluentd_prometheus_alerts.yaml"
)

func removeFluentd(cluster *ClusterLogging) (err error) {
	if cluster.Spec.ManagementState == logging.ManagementStateManaged {

		if err = utils.RemoveService(cluster.Namespace, "fluentd"); err != nil {
			return
		}

		if err = utils.RemoveServiceMonitor(cluster.Namespace, "fluentd"); err != nil {
			return
		}

		if err = utils.RemovePrometheusRule(cluster.Namespace, "fluentd"); err != nil {
			return
		}

		if err = utils.RemoveConfigMap(cluster.Namespace, "fluentd"); err != nil {
			return
		}

		if err = utils.RemoveSecret(cluster.Namespace, "fluentd"); err != nil {
			return
		}

		if err = utils.RemoveDaemonset(cluster.Namespace, "fluentd"); err != nil {
			return
		}
	}

	return nil
}

func createOrUpdateFluentdService(cluster *ClusterLogging) error {
	service := utils.NewService(
		"fluentd",
		cluster.Namespace,
		"fluentd",
		[]v1.ServicePort{
			{
				Port:       metricsPort,
				TargetPort: intstr.FromString(metricsPortName),
				Name:       metricsPortName,
			},
		},
	)

	service.Annotations = map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": metricsVolumeName,
	}

	utils.AddOwnerRefToObject(service, utils.AsOwner(cluster))

	err := sdk.Create(service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd service: %v", err)
	}

	return nil
}

func createOrUpdateFluentdServiceMonitor(cluster *ClusterLogging) error {
	serviceMonitor := utils.NewServiceMonitor("fluentd", cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   metricsPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", "fluentd", cluster.Namespace),
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

	err := sdk.Create(serviceMonitor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd ServiceMonitor: %v", err)
	}

	return nil
}

func createOrUpdateFluentdPrometheusRule(cluster *ClusterLogging) error {
	promRule := utils.NewPrometheusRule("fluentd", cluster.Namespace)

	promRuleSpec, err := utils.NewPrometheusRuleSpecFrom(fluentdAlertsFile)
	if err != nil {
		return fmt.Errorf("Failure creating the fluentd PrometheusRule: %v", err)
	}

	promRule.Spec = *promRuleSpec

	utils.AddOwnerRefToObject(promRule, utils.AsOwner(cluster))

	err = sdk.Create(promRule)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd PrometheusRule: %v", err)
	}

	return nil
}

func createOrUpdateFluentdConfigMap(cluster *ClusterLogging) error {

	fluentdConfigMap := utils.NewConfigMap(
		"fluentd",
		cluster.Namespace,
		map[string]string{
			"fluent.conf":          string(utils.GetFileContents("/usr/share/logging/fluentd/fluent.conf")),
			"throttle-config.yaml": string(utils.GetFileContents("/usr/share/logging/fluentd/fluentd-throttle-config.yaml")),
			"secure-forward.conf":  string(utils.GetFileContents("/usr/share/logging/fluentd/secure-forward.conf")),
		},
	)

	utils.AddOwnerRefToObject(fluentdConfigMap, utils.AsOwner(cluster))

	err := sdk.Create(fluentdConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Fluentd configmap: %v", err)
	}

	return nil
}
func createOrUpdateFluentdSecret(cluster *ClusterLogging) error {

	fluentdSecret := utils.NewSecret(
		"fluentd",
		cluster.Namespace,
		map[string][]byte{
			"app-ca":     utils.GetWorkingDirFileContents("ca.crt"),
			"app-key":    utils.GetWorkingDirFileContents("system.logging.fluentd.key"),
			"app-cert":   utils.GetWorkingDirFileContents("system.logging.fluentd.crt"),
			"infra-ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"infra-key":  utils.GetWorkingDirFileContents("system.logging.fluentd.key"),
			"infra-cert": utils.GetWorkingDirFileContents("system.logging.fluentd.crt"),
		})

	cluster.AddOwnerRefTo(fluentdSecret)

	err := utils.CreateOrUpdateSecret(fluentdSecret)
	if err != nil {
		return err
	}

	return nil
}

func newFluentdPodSpec(logging *logging.ClusterLogging, elasticsearchAppName string, elasticsearchInfraName string) v1.PodSpec {
	var resources = logging.Spec.Collection.Logs.FluentdSpec.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultFluentdMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultFluentdMemory,
				v1.ResourceCPU:    defaultFluentdCpuRequest,
			},
		}
	}
	fluentdContainer := utils.NewContainer("fluentd", v1.PullIfNotPresent, *resources)

	fluentdContainer.Ports = []v1.ContainerPort{
		v1.ContainerPort{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	fluentdContainer.Env = []v1.EnvVar{
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
		{Name: "ES_HOST", Value: elasticsearchAppName},
		{Name: "ES_PORT", Value: "9200"},
		{Name: "ES_CLIENT_CERT", Value: "/etc/fluent/keys/app-cert"},
		{Name: "ES_CLIENT_KEY", Value: "/etc/fluent/keys/app-key"},
		{Name: "ES_CA", Value: "/etc/fluent/keys/app-ca"},
		{Name: "METRICS_CERT", Value: "/etc/fluent/metrics/tls.crt"},
		{Name: "METRICS_KEY", Value: "/etc/fluent/metrics/tls.key"},
		{Name: "OPS_HOST", Value: elasticsearchInfraName},
		{Name: "OPS_PORT", Value: "9200"},
		{Name: "OPS_CLIENT_CERT", Value: "/etc/fluent/keys/infra-cert"},
		{Name: "OPS_CLIENT_KEY", Value: "/etc/fluent/keys/infra-key"},
		{Name: "OPS_CA", Value: "/etc/fluent/keys/infra-ca"},
		{Name: "JOURNAL_SOURCE", Value: ""},
		{Name: "JOURNAL_READ_FROM_HEAD", Value: ""},
		{Name: "BUFFER_QUEUE_LIMIT", Value: "32"},
		{Name: "BUFFER_SIZE_LIMIT", Value: "8m"},
		{Name: "FILE_BUFFER_LIMIT", Value: "256Mi"},
		{Name: "FLUENTD_CPU_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.cpu"}}},
		{Name: "FLUENTD_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.memory"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "status.hostIP"}}},
	}

	fluentdContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "runlogjournal", MountPath: "/run/log/journal"},
		{Name: "varlog", MountPath: "/var/log"},
		{Name: "varlibdockercontainers", ReadOnly: true, MountPath: "/var/lib/docker"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/fluent/configs.d/user"},
		{Name: "certs", ReadOnly: true, MountPath: "/etc/fluent/keys"},
		{Name: "dockerhostname", ReadOnly: true, MountPath: "/etc/docker-hostname"},
		{Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
		{Name: "dockercfg", ReadOnly: true, MountPath: "/etc/sysconfig/docker"},
		{Name: "dockerdaemoncfg", ReadOnly: true, MountPath: "/etc/docker"},
		{Name: "filebufferstorage", MountPath: "/var/lib/fluentd"},
		{Name: metricsVolumeName, MountPath: "/etc/fluent/metrics"},
	}

	fluentdContainer.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}

	fluentdPodSpec := utils.NewPodSpec(
		"logcollector",
		[]v1.Container{fluentdContainer},
		[]v1.Volume{
			{Name: "runlogjournal", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/run/log/journal"}}},
			{Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
			{Name: "varlibdockercontainers", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker"}}},
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "fluentd"}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "fluentd"}}},
			{Name: "dockerhostname", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/hostname"}}},
			{Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
			{Name: "dockercfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/sysconfig/docker"}}},
			{Name: "dockerdaemoncfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/docker"}}},
			{Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/fluentd"}}},
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: metricsVolumeName}}},
		},
		logging.Spec.Collection.Logs.FluentdSpec.NodeSelector,
	)

	fluentdPodSpec.PriorityClassName = clusterLoggingPriorityClassName

	fluentdPodSpec.NodeSelector = logging.Spec.Collection.Logs.FluentdSpec.NodeSelector

	fluentdPodSpec.Tolerations = []v1.Toleration{
		v1.Toleration{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	return fluentdPodSpec
}

func createOrUpdateFluentdDaemonset(cluster *ClusterLogging) (err error) {

	fluentdPodSpec := newFluentdPodSpec(cluster.ClusterLogging, "elasticsearch", "elasticsearch")

	fluentdDaemonset := utils.NewDaemonSet("fluentd", cluster.Namespace, "fluentd", "fluentd", fluentdPodSpec)
	cluster.AddOwnerRefTo(fluentdDaemonset)

	err = sdk.Create(fluentdDaemonset)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Fluentd Daemonset %v", err)
	}

	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return updateFluentdDaemonsetIfRequired(fluentdDaemonset)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func updateFluentdDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
	current := desired.DeepCopy()

	current.Spec = apps.DaemonSetSpec{}

	if err = sdk.Get(current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Fluentd daemonset: %v", err)
	}

	flushBuffer := isBufferFlushRequired(current, desired)
	desired, different := isDaemonsetDifferent(current, desired)

	if different {

		if flushBuffer {
			current.Spec.Template.Spec.Containers[0].Env = append(current.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: "FLUSH_AT_SHUTDOWN", Value: "True"})
			if err = sdk.Update(current); err != nil {
				logrus.Debugf("Failed to prepare Fluentd daemonset to flush its buffers: %v", err)
				return err
			}

			// wait for pods to all restart then continue
			if err = waitForDaemonSetReady(current); err != nil {
				return fmt.Errorf("Timed out waiting for Fluentd to be ready")
			}
		}

		if err = sdk.Update(desired); err != nil {
			return err
		}
	}

	return nil
}
