package k8shandler

import (
	"fmt"
	"io/ioutil"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	rsyslogAlertsFile = "rsyslog/rsyslog_prometheus_alerts.yaml"
)

func (clusterRequest *ClusterLoggingRequest) removeRsyslog() (err error) {
	if clusterRequest.isManaged() {

		if err = clusterRequest.RemoveService("rsyslog"); err != nil {
			return
		}

		if err = clusterRequest.RemoveServiceMonitor("rsyslog"); err != nil {
			return
		}

		if err = clusterRequest.RemovePrometheusRule("rsyslog"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("rsyslog-bin"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("rsyslog-main"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("rsyslog"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("logrotate-bin"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("logrotate-crontab"); err != nil {
			return
		}

		if err = clusterRequest.RemoveSecret("rsyslog"); err != nil {
			return
		}

		if err = clusterRequest.RemoveDaemonset("rsyslog"); err != nil {
			return
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateRsyslogService() error {
	service := NewService(
		"rsyslog",
		clusterRequest.cluster.Namespace,
		"rsyslog",
		[]v1.ServicePort{
			{
				Port:       metricsPort,
				TargetPort: intstr.FromString(metricsPortName),
				Name:       metricsPortName,
			},
		},
	)

	service.Annotations = map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": "rsyslog-metrics",
	}

	utils.AddOwnerRefToObject(service, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the rsyslog service: %v", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateRsyslogServiceMonitor() error {

	cluster := clusterRequest.cluster

	serviceMonitor := NewServiceMonitor("rsyslog", cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   metricsPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", "rsyslog", cluster.Namespace),
			// ServerName can be e.g. rsyslog.openshift-logging.svc
		},
	}

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"logging-infra": "support",
		},
	}

	serviceMonitor.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-rsyslog",
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{cluster.Namespace},
		},
	}

	utils.AddOwnerRefToObject(serviceMonitor, utils.AsOwner(cluster))

	err := clusterRequest.Create(serviceMonitor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the rsyslog ServiceMonitor: %v", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateRsyslogPrometheusRule() error {

	cluster := clusterRequest.cluster

	promRule := NewPrometheusRule("rsyslog", cluster.Namespace)

	promRuleSpec, err := NewPrometheusRuleSpecFrom(utils.GetShareDir() + "/" + rsyslogAlertsFile)
	if err != nil {
		return fmt.Errorf("Failure creating the rsyslog PrometheusRule: %v", err)
	}

	promRule.Spec = *promRuleSpec

	utils.AddOwnerRefToObject(promRule, utils.AsOwner(cluster))

	err = clusterRequest.Create(promRule)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the rsyslog PrometheusRule: %v", err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateRsyslogConfigMap() error {
	// need three configmaps
	// - one for rsyslog run.sh script - rsyslog-bin
	// - one for main rsyslog.conf file - rsyslog-main
	// - one for the actual config files - rsyslog
	rsyslogConfigMaps := make(map[string]*v1.ConfigMap)
	rsyslogBinConfigMap := NewConfigMap(
		"rsyslog-bin",
		clusterRequest.cluster.Namespace,
		map[string]string{
			"rsyslog.sh": string(utils.GetFileContents(utils.GetShareDir() + "/rsyslog/rsyslog.sh")),
		},
	)
	rsyslogConfigMaps["rsyslog-bin"] = rsyslogBinConfigMap

	rsyslogMainConfigMap := NewConfigMap(
		"rsyslog-main",
		clusterRequest.cluster.Namespace,
		map[string]string{
			"rsyslog.conf": string(utils.GetFileContents(utils.GetShareDir() + "/rsyslog/rsyslog.conf")),
		},
	)
	rsyslogConfigMaps["rsyslog-main"] = rsyslogMainConfigMap

	rsyslogConfigMapFiles := make(map[string]string)
	readerDir, err := ioutil.ReadDir(utils.GetShareDir() + "/rsyslog")
	if err != nil {
		return fmt.Errorf("Failure %v to read files from directory '%v/rsyslog' for Rsyslog configmap", err, utils.GetShareDir())
	}
	for _, fileInfo := range readerDir {
		// exclude files provided by other configmaps
		if fileInfo.Name() == "rsyslog.conf" {
			continue
		}
		if fileInfo.Name() == "rsyslog.sh" {
			continue
		}
		// include all other files
		fullname := utils.GetShareDir() + "/rsyslog/" + fileInfo.Name()
		rsyslogConfigMapFiles[fileInfo.Name()] = string(utils.GetFileContents(fullname))
	}
	rsyslogConfigMap := NewConfigMap(
		"rsyslog",
		clusterRequest.cluster.Namespace,
		rsyslogConfigMapFiles,
	)
	rsyslogConfigMaps["rsyslog"] = rsyslogConfigMap
	for name, cm := range rsyslogConfigMaps {
		utils.AddOwnerRefToObject(cm, utils.AsOwner(clusterRequest.cluster))

		err = clusterRequest.Create(cm)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Rsyslog configmap %v: %v", name, err)
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateLogrotateConfigMap() error {

	// need two configmaps
	// - one for logrotate cron.sh, logrotate.sh and logrotate_pod.sh script - bin
	// - one for logrotate crontab - crontab
	logrotateConfigMaps := make(map[string]*v1.ConfigMap)
	logrotateBinConfigMap := NewConfigMap(
		"logrotate-bin",
		clusterRequest.cluster.Namespace,
		map[string]string{
			"cron.sh":          string(utils.GetFileContents(utils.GetShareDir() + "/logrotate/cron.sh")),
			"logrotate.sh":     string(utils.GetFileContents(utils.GetShareDir() + "/logrotate/logrotate.sh")),
			"logrotate_pod.sh": string(utils.GetFileContents(utils.GetShareDir() + "/logrotate/logrotate_pod.sh")),
		},
	)
	logrotateConfigMaps["logrotate-bin"] = logrotateBinConfigMap

	logrotateCrontabConfigMap := NewConfigMap(
		"logrotate-crontab",
		clusterRequest.cluster.Namespace,
		map[string]string{
			"logrotate": string(utils.GetFileContents(utils.GetShareDir() + "/logrotate/logrotate")),
		},
	)
	logrotateConfigMaps["logrotate-crontab"] = logrotateCrontabConfigMap

	for name, cm := range logrotateConfigMaps {
		utils.AddOwnerRefToObject(cm, utils.AsOwner(clusterRequest.cluster))

		err := clusterRequest.Create(cm)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Logrotate configmap %v: %v", name, err)
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateRsyslogSecret() error {

	rsyslogSecret := NewSecret(
		"rsyslog",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"app-ca":     utils.GetWorkingDirFileContents("ca.crt"),
			"app-key":    utils.GetWorkingDirFileContents("system.logging.rsyslog.key"),
			"app-cert":   utils.GetWorkingDirFileContents("system.logging.rsyslog.crt"),
			"infra-ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"infra-key":  utils.GetWorkingDirFileContents("system.logging.rsyslog.key"),
			"infra-cert": utils.GetWorkingDirFileContents("system.logging.rsyslog.crt"),
		})

	utils.AddOwnerRefToObject(rsyslogSecret, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.CreateOrUpdateSecret(rsyslogSecret)
	if err != nil {
		return err
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateRsyslogDaemonset() (err error) {

	cluster := clusterRequest.cluster

	rsyslogPodSpec := newRsyslogPodSpec(clusterRequest.cluster, "elasticsearch", "elasticsearch")

	rsyslogDaemonset := NewDaemonSet("rsyslog", cluster.Namespace, "rsyslog", "rsyslog", rsyslogPodSpec)

	utils.AddOwnerRefToObject(rsyslogDaemonset, utils.AsOwner(cluster))

	err = clusterRequest.Create(rsyslogDaemonset)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Rsyslog Daemonset %v", err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateRsyslogDaemonsetIfRequired(rsyslogDaemonset)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func newRsyslogPodSpec(logging *logging.ClusterLogging, elasticsearchAppName string, elasticsearchInfraName string) v1.PodSpec {
	var resources = logging.Spec.Collection.Logs.RsyslogSpec.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultRsyslogMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultRsyslogMemory,
				v1.ResourceCPU:    defaultRsyslogCpuRequest,
			}}
	}
	rsyslogContainer := NewContainer("rsyslog", "rsyslog", v1.PullIfNotPresent, *resources)

	rsyslogContainer.Ports = []v1.ContainerPort{
		v1.ContainerPort{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	rsyslogContainer.Env = []v1.EnvVar{
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
		{Name: "ES_HOST", Value: elasticsearchAppName},
		{Name: "ES_PORT", Value: "9200"},
		{Name: "ES_CLIENT_CERT", Value: "/etc/rsyslog/keys/app-cert"},
		{Name: "ES_CLIENT_KEY", Value: "/etc/rsyslog/keys/app-key"},
		{Name: "ES_CA", Value: "/etc/rsyslog/keys/app-ca"},
		{Name: "OPS_HOST", Value: elasticsearchInfraName},
		{Name: "OPS_PORT", Value: "9200"},
		{Name: "OPS_CLIENT_CERT", Value: "/etc/rsyslog/keys/infra-cert"},
		{Name: "OPS_CLIENT_KEY", Value: "/etc/rsyslog/keys/infra-key"},
		{Name: "OPS_CA", Value: "/etc/rsyslog/keys/infra-ca"},
		{Name: "BUFFER_QUEUE_LIMIT", Value: "32"},
		{Name: "BUFFER_SIZE_LIMIT", Value: "8m"},
		{Name: "FILE_BUFFER_LIMIT", Value: "256Mi"},
		{Name: "RSYSLOG_CPU_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "rsyslog", Resource: "limits.cpu"}}},
		{Name: "RSYSLOG_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "rsyslog", Resource: "limits.memory"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "status.hostIP"}}},
		{Name: "RSYSLOG_WORKDIRECTORY", Value: "/var/lib/rsyslog.pod"},
		{Name: "CDM_KEEP_EMPTY_FIELDS", Value: "message"}, // by default, keep empty messages
	}

	rsyslogContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "runlogjournal", MountPath: "/run/log/journal"},
		{Name: "varlog", MountPath: "/var/log"},
		{Name: "varrun", MountPath: "/var/run"},
		{Name: "varlibdockercontainers", ReadOnly: true, MountPath: "/var/lib/docker"},
		{Name: "bin", ReadOnly: true, MountPath: "/opt/app-root/bin"},
		{Name: "main", ReadOnly: true, MountPath: "/etc/rsyslog/conf"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/rsyslog.d"},
		{Name: "certs", ReadOnly: true, MountPath: "/etc/rsyslog/keys"},
		{Name: "dockerhostname", ReadOnly: true, MountPath: "/etc/docker-hostname"},
		{Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
		{Name: "machineid", ReadOnly: true, MountPath: "/etc/machine-id"},
		{Name: "filebufferstorage", MountPath: "/var/lib/rsyslog.pod"},
		{Name: metricsVolumeName, MountPath: "/etc/rsyslog/metrics"},
	}

	rsyslogContainer.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}

	rsyslogContainer.Command = []string{
		"/bin/sh",
	}

	rsyslogContainer.Args = []string{
		"/opt/app-root/bin/rsyslog.sh",
	}

	logrotateContainer := NewContainer("logrotate", "rsyslog", v1.PullIfNotPresent, *resources)

	logrotateContainer.Env = []v1.EnvVar{
		{Name: "LOGGING_FILE_PATH", Value: "/var/log/rsyslog/rsyslog.log"},
		{Name: "LOGGING_FILE_SIZE", Value: "1024000"},
		{Name: "LOGGING_FILE_AGE", Value: "10"},
		{Name: "CROND_OPTIONS", Value: ""},
		{Name: "RSYSLOG_WORKDIRECTORY", Value: "/var/lib/rsyslog.pod"},
	}

	logrotateContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "logrotate-bin", ReadOnly: true, MountPath: "/opt/app-root/bin"},
		{Name: "logrotate-crontab", ReadOnly: true, MountPath: "/etc/cron.d"},
		{Name: "varlog", MountPath: "/var/log"},
		{Name: "varrun", MountPath: "/var/run"},
		{Name: "filebufferstorage", MountPath: "/var/lib/rsyslog.pod"},
	}

	logrotateContainer.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}

	logrotateContainer.Command = []string{
		"/bin/sh",
	}

	logrotateContainer.Args = []string{
		"/opt/app-root/bin/cron.sh",
	}

	tolerations := utils.AppendTolerations(
		logging.Spec.Collection.Logs.RsyslogSpec.Tolerations,
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

	rsyslogPodSpec := NewPodSpec(
		"logcollector",
		[]v1.Container{rsyslogContainer, logrotateContainer},
		[]v1.Volume{
			{Name: "runlogjournal", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/run/log/journal"}}},
			{Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
			{Name: "varrun", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/run"}}},
			{Name: "varlibdockercontainers", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker"}}},
			{Name: "bin", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "rsyslog-bin"}}}},
			{Name: "main", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "rsyslog-main"}}}},
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "rsyslog"}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "rsyslog"}}},
			{Name: "dockerhostname", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/hostname"}}},
			{Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
			{Name: "machineid", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/machine-id"}}},
			{Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/rsyslog.pod"}}},
			{Name: "logrotate-bin", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "logrotate-bin"}}}},
			{Name: "logrotate-crontab", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "logrotate-crontab"}}}},
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "rsyslog-metrics"}}},
		},
		logging.Spec.Collection.Logs.RsyslogSpec.NodeSelector,
		tolerations,
	)

	rsyslogPodSpec.PriorityClassName = clusterLoggingPriorityClassName

	b := bool(true)
	rsyslogPodSpec.ShareProcessNamespace = &b

	return rsyslogPodSpec
}

func (clusterRequest *ClusterLoggingRequest) updateRsyslogDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
	current := desired.DeepCopy()

	if err = clusterRequest.Get(desired.Name, current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Rsyslog daemonset: %v", err)
	}

	current, different := isDaemonsetDifferent(current, desired)

	if different {
		if err = clusterRequest.Update(current); err != nil {
			return err
		}
	}

	return nil
}
