package k8shandler

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	collector "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	promtailName        = "promtail"
	promtailMetricsName = "promtail-metrics"
)

func (clusterRequest *ClusterLoggingRequest) createOrUpdatePromTailService() error {
	service := NewService(
		promtailName,
		clusterRequest.cluster.Namespace,
		promtailName,
		[]v1.ServicePort{
			{
				Port:       metricsPort,
				TargetPort: intstr.FromString(metricsPortName),
				Name:       metricsPortName,
			},
		},
	)

	service.Annotations = map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": promtailMetricsName,
	}

	utils.AddOwnerRefToObject(service, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating service %q: %v", service.Name, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdatePromTailServiceMonitor() error {

	cluster := clusterRequest.cluster

	serviceMonitor := NewServiceMonitor(promtailName, cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   metricsPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", promtailName, cluster.Namespace),
		},
	}

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"logging-infra": "support",
		},
	}

	serviceMonitor.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-" + promtailName,
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{cluster.Namespace},
		},
	}

	utils.AddOwnerRefToObject(serviceMonitor, utils.AsOwner(cluster))

	err := clusterRequest.Create(serviceMonitor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the ServiceMonitor %q: %v", promtailName, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdatePromTailConfigMap() error {

	configMap := NewConfigMap(
		promtailName,
		clusterRequest.cluster.Namespace,
		map[string]string{
			"promtail.yaml": string(utils.GetFileContents(utils.GetShareDir() + "/promtail/promtail.yaml")),
		},
	)

	utils.AddOwnerRefToObject(configMap, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(configMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing configmap %q: %v", configMap.Name, err)
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &core.ConfigMap{}
		if err = clusterRequest.Get(configMap.Name, current); err != nil {
			if errors.IsNotFound(err) {
				logrus.Debugf("Returning nil. The configmap %q was not found even though create previously failed.  Was it culled?", configMap.Name)
				return nil
			}
			return fmt.Errorf("Failed to get %v configmap for %q: %v", configMap.Name, clusterRequest.cluster.Name, err)
		}
		if reflect.DeepEqual(configMap.Data, current.Data) {
			return nil
		}
		current.Data = configMap.Data
		return clusterRequest.Update(current)
	})

	return retryErr
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdatePromTailSecret() error {
	return nil
}

func newPromTailPodSpec(collector *collector.CollectorSpec) v1.PodSpec {
	var resources = collector.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultPromTailMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultPromTailMemory,
				v1.ResourceCPU:    defaultPromTailCpuRequest,
			},
		}
	}
	container := NewContainer(promtailName, promtailName, v1.PullIfNotPresent, *resources)

	container.Ports = []v1.ContainerPort{
		v1.ContainerPort{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	hasher := md5.New()
	hasher.Write([]byte(string(utils.GetFileContents(utils.GetShareDir() + "/promtail/promtail.yaml"))))
	md5hash := hex.EncodeToString(hasher.Sum(nil))
	container.Env = []v1.EnvVar{
		v1.EnvVar{Name: "PROMTAIL_YAML_HASH", Value: md5hash},
	}
	container.Args = []string{
		"-config.file=/etc/promtail/promtail.yaml",
		"-client.url=" + collector.PromTailSpec.Endpoint,
	}

	container.VolumeMounts = []v1.VolumeMount{
		{Name: "varlog", ReadOnly: true, MountPath: "/var/log"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/promtail"},
		{Name: "dockerhostname", ReadOnly: true, MountPath: "/etc/docker-hostname"},
		{Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
		{Name: "filebufferstorage", MountPath: "/var/lib/promtail"},
	}

	container.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}

	tolerations := utils.AppendTolerations(
		collector.Tolerations,
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

	podSpec := NewPodSpec(
		"logcollector",
		[]v1.Container{container},
		[]v1.Volume{
			{Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: promtailName}}}},
			{Name: "dockerhostname", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/hostname"}}},
			{Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
			{Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/promtail"}}},
		},
		collector.NodeSelector,
		tolerations,
	)

	podSpec.PriorityClassName = clusterLoggingPriorityClassName
	// Shorten the termination grace period from the default 30 sec to 10 sec.
	podSpec.TerminationGracePeriodSeconds = utils.GetInt64(10)

	return podSpec
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdatePromTailDaemonset() (err error) {

	cluster := clusterRequest.cluster

	podSpec := newPromTailPodSpec(clusterRequest.Collector)

	daemonset := NewDaemonSet(promtailName, cluster.Namespace, promtailName, promtailName, podSpec)

	uid := getServiceAccountLogCollectorUID()
	if len(uid) == 0 {
		// There's no uid for logcollector serviceaccount; setting ClusterLogging for the ownerReference.
		utils.AddOwnerRefToObject(daemonset, utils.AsOwner(cluster))
	} else {
		// There's a uid for logcollector serviceaccount; setting the ServiceAccount for the ownerReference with blockOwnerDeletion.
		utils.AddOwnerRefToObject(daemonset, NewLogCollectorServiceAccountRef(uid))
	}

	err = clusterRequest.Create(daemonset)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Daemonset %q: %v", daemonset.Name, err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updatePromTailDaemonsetIfRequired(daemonset)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) updatePromTailDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
	current := desired.DeepCopy()

	if err = clusterRequest.Get(desired.Name, current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get daemonset %q: %v", desired.Name, err)
	}

	if _, different := isDaemonsetDifferent(current, desired); different {

		if err = clusterRequest.Update(desired); err != nil {
			return err
		}
	}

	return nil
}
