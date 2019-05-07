package k8shandler

import (
	"fmt"
	"io/ioutil"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
)

func removeRsyslog(cluster *ClusterLogging) (err error) {
	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		if err = utils.RemoveConfigMap(cluster.Namespace, "rsyslog-bin"); err != nil {
			return
		}

		if err = utils.RemoveConfigMap(cluster.Namespace, "rsyslog-main"); err != nil {
			return
		}

		if err = utils.RemoveConfigMap(cluster.Namespace, "rsyslog"); err != nil {
			return
		}

		if err = utils.RemoveSecret(cluster.Namespace, "rsyslog"); err != nil {
			return
		}

		if err = utils.RemoveDaemonset(cluster.Namespace, "rsyslog"); err != nil {
			return
		}
	}

	return nil
}

func createOrUpdateRsyslogConfigMap(logging *ClusterLogging) error {

	// need three configmaps
	// - one for rsyslog run.sh script - rsyslog-bin
	// - one for main rsyslog.conf file - rsyslog-main
	// - one for the actual config files - rsyslog
	rsyslogConfigMaps := make(map[string]*v1.ConfigMap)
	rsyslogBinConfigMap := utils.NewConfigMap(
		"rsyslog-bin",
		logging.Namespace,
		map[string]string{
			"rsyslog.sh": string(utils.GetFileContents("files/rsyslog/rsyslog.sh")),
		},
	)
	rsyslogConfigMaps["rsyslog-bin"] = rsyslogBinConfigMap

	rsyslogMainConfigMap := utils.NewConfigMap(
		"rsyslog-main",
		logging.Namespace,
		map[string]string{
			"rsyslog.conf": string(utils.GetFileContents("files/rsyslog/rsyslog.conf")),
		},
	)
	rsyslogConfigMaps["rsyslog-main"] = rsyslogMainConfigMap

	rsyslogConfigMapFiles := make(map[string]string)
	readerDir, err := ioutil.ReadDir("files/rsyslog")
	if err != nil {
		return fmt.Errorf("Failure %v to read files from directory 'files/rsyslog' for Rsyslog configmap", err)
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
		fullname := "files/rsyslog/" + fileInfo.Name()
		rsyslogConfigMapFiles[fileInfo.Name()] = string(utils.GetFileContents(fullname))
	}
	rsyslogConfigMap := utils.NewConfigMap(
		"rsyslog",
		logging.Namespace,
		rsyslogConfigMapFiles,
	)
	rsyslogConfigMaps["rsyslog"] = rsyslogConfigMap
	for name, cm := range rsyslogConfigMaps {
		logging.AddOwnerRefTo(cm)

		err = sdk.Create(cm)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Rsyslog configmap %v: %v", name, err)
		}
	}

	return nil
}

func createOrUpdateRsyslogSecret(logging *ClusterLogging) error {

	rsyslogSecret := utils.NewSecret(
		"rsyslog",
		logging.Namespace,
		map[string][]byte{
			"app-ca":     utils.GetWorkingDirFileContents("ca.crt"),
			"app-key":    utils.GetWorkingDirFileContents("system.logging.rsyslog.key"),
			"app-cert":   utils.GetWorkingDirFileContents("system.logging.rsyslog.crt"),
			"infra-ca":   utils.GetWorkingDirFileContents("ca.crt"),
			"infra-key":  utils.GetWorkingDirFileContents("system.logging.rsyslog.key"),
			"infra-cert": utils.GetWorkingDirFileContents("system.logging.rsyslog.crt"),
		})

	logging.AddOwnerRefTo(rsyslogSecret)

	err := utils.CreateOrUpdateSecret(rsyslogSecret)
	if err != nil {
		return err
	}

	return nil
}

func createOrUpdateRsyslogDaemonset(cluster *ClusterLogging) (err error) {

	var rsyslogPodSpec v1.PodSpec

	rsyslogPodSpec = newRsyslogPodSpec(cluster.ClusterLogging, "elasticsearch", "elasticsearch")

	rsyslogDaemonset := utils.NewDaemonSet("rsyslog", cluster.Namespace, "rsyslog", "rsyslog", rsyslogPodSpec)

	utils.AddOwnerRefToObject(rsyslogDaemonset, utils.AsOwner(cluster))

	err = sdk.Create(rsyslogDaemonset)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Rsyslog Daemonset %v", err)
	}

	if cluster.Spec.ManagementState == logging.ManagementStateManaged {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return updateRsyslogDaemonsetIfRequired(rsyslogDaemonset)
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
	rsyslogContainer := utils.NewContainer("rsyslog", v1.PullIfNotPresent, *resources)

	rsyslogContainer.Env = []v1.EnvVar{
		{Name: "MERGE_JSON_LOG", Value: "false"},
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
		{Name: "JOURNAL_READ_FROM_HEAD", Value: ""},
		{Name: "BUFFER_QUEUE_LIMIT", Value: "32"},
		{Name: "BUFFER_SIZE_LIMIT", Value: "8m"},
		{Name: "FILE_BUFFER_LIMIT", Value: "256Mi"},
		{Name: "RSYSLOG_CPU_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "rsyslog", Resource: "limits.cpu"}}},
		{Name: "RSYSLOG_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "rsyslog", Resource: "limits.memory"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "status.hostIP"}}},
	}

	rsyslogContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "runlogjournal", MountPath: "/run/log/journal"},
		{Name: "varlog", MountPath: "/var/log"},
		{Name: "varlibdockercontainers", ReadOnly: true, MountPath: "/var/lib/docker"},
		{Name: "bin", ReadOnly: true, MountPath: "/opt/app-root/bin"},
		{Name: "main", ReadOnly: true, MountPath: "/etc/rsyslog/conf"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/rsyslog.d"},
		{Name: "certs", ReadOnly: true, MountPath: "/etc/rsyslog/keys"},
		{Name: "dockerhostname", ReadOnly: true, MountPath: "/etc/docker-hostname"},
		{Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
		{Name: "machineid", ReadOnly: true, MountPath: "/etc/machine-id"},
		{Name: "filebufferstorage", MountPath: "/var/lib/rsyslog.pod"},
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

	rsyslogPodSpec := utils.NewPodSpec(
		"logcollector",
		[]v1.Container{rsyslogContainer},
		[]v1.Volume{
			{Name: "runlogjournal", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/run/log/journal"}}},
			{Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
			{Name: "varlibdockercontainers", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker"}}},
			{Name: "bin", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "rsyslog-bin"}}}},
			{Name: "main", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "rsyslog-main"}}}},
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "rsyslog"}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "rsyslog"}}},
			{Name: "dockerhostname", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/hostname"}}},
			{Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
			{Name: "machineid", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/machine-id"}}},
			{Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/rsyslog.pod"}}},
		},
		logging.Spec.Collection.Logs.RsyslogSpec.NodeSelector,
	)

	rsyslogPodSpec.PriorityClassName = clusterLoggingPriorityClassName

	rsyslogPodSpec.NodeSelector = logging.Spec.Collection.Logs.RsyslogSpec.NodeSelector

	rsyslogPodSpec.Tolerations = []v1.Toleration{
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
	}

	return rsyslogPodSpec
}

func updateRsyslogDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
	current := desired.DeepCopy()

	current.Spec = apps.DaemonSetSpec{}

	if err = sdk.Get(current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get Rsyslog daemonset: %v", err)
	}

	current, different := isDaemonsetDifferent(current, desired)

	if different {
		if err = sdk.Update(current); err != nil {
			return err
		}
	}

	return nil
}
