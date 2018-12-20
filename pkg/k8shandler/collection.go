package k8shandler

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"strings"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	retryInterval = time.Second * 30
	timeout       = time.Second * 1800
)

func CreateOrUpdateCollection(cluster *ClusterLogging) (err error) {

	if err = createOrUpdateCollectionPriorityClass(cluster); err != nil {
		return
	}

	if cluster.Spec.Collection.LogCollection.Type == logging.LogCollectionTypeFluentd {
		if err = createOrUpdateServiceAccount(cluster, Fluentd); err != nil {
			return
		}

		if err = createOrUpdateFluentdConfigMap(cluster); err != nil {
			return
		}

		if err = createOrUpdateFluentdSecret(cluster); err != nil {
			return
		}

		if err = createOrUpdateFluentdDaemonset(cluster); err != nil {
			return
		}

		fluentdStatus, err := getFluentdCollectorStatus(cluster.Namespace)
		if err != nil {
			return fmt.Errorf("Failed to get status of Fluentd: %v", err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if exists := utils.DoesClusterLoggingExist(cluster.ClusterLogging); exists {
				if !reflect.DeepEqual(fluentdStatus, cluster.ClusterLogging.Status.Collection.LogCollection.FluentdStatus) {
					if printUpdateMessage {
						logrus.Info("Updating status of Fluentd")
						printUpdateMessage = false
					}
					cluster.ClusterLogging.Status.Collection.LogCollection.FluentdStatus = fluentdStatus
					return sdk.Update(cluster)
				}
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Fluentd status: %v", retryErr)
		}
	}
	if cluster.Spec.Collection.LogCollection.Type == logging.LogCollectionTypeRsyslog {
		if err = createOrUpdateServiceAccount(cluster, Rsyslog); err != nil {
			return
		}

		if err = createOrUpdateRsyslogConfigMap(cluster.ClusterLogging); err != nil {
			return
		}

		if err = createOrUpdateRsyslogSecret(cluster); err != nil {
			return
		}

		if err = createOrUpdateRsyslogDaemonset(cluster); err != nil {
			return
		}

		rsyslogStatus, err := getRsyslogCollectorStatus(cluster.Namespace)
		if err != nil {
			return fmt.Errorf("Failed to get status of Rsyslog: %v", err)
		}

		printUpdateMessage := true
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if exists := utils.DoesClusterLoggingExist(cluster.ClusterLogging); exists {
				if !reflect.DeepEqual(rsyslogStatus, cluster.ClusterLogging.Status.Collection.LogCollection.RsyslogStatus) {
					if printUpdateMessage {
						logrus.Info("Updating status of Rsyslog")
						printUpdateMessage = false
					}
					cluster.ClusterLogging.Status.Collection.LogCollection.RsyslogStatus = rsyslogStatus
					return sdk.Update(cluster)
				}
			}
			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("Failed to update Cluster Logging Rsyslog status: %v", retryErr)
		}
	}

	return nil
}

func createOrUpdateCollectionPriorityClass(logging *ClusterLogging) error {

	collectionPriorityClass := utils.NewPriorityClass(PriorityClassName, 1000000, false, "This priority class is for the Cluster-Logging Collector")
	logging.addOwnerRefTo(collectionPriorityClass)

	err := sdk.Create(collectionPriorityClass)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Collection priority class: %v", err)
	}

	return nil
}

func createOrUpdateServiceAccount(logging *ClusterLogging, name string) error {

	fluentdServiceAccount := utils.NewServiceAccount(name, logging.Namespace)

	logging.addOwnerRefTo(fluentdServiceAccount)

	err := sdk.Create(fluentdServiceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating service account '%s': %v", name, err)
	}

	return nil
}

func createOrUpdateFluentdConfigMap(logging *ClusterLogging) error {

	fluentdConfigMap := utils.NewConfigMap(
		Fluentd,
		logging.Namespace,
		map[string]string{
			"fluent.conf":          string(utils.GetFileContents("files/fluent.conf")),
			"throttle-config.yaml": string(utils.GetFileContents("files/fluentd-throttle-config.yaml")),
			"secure-forward.conf":  string(utils.GetFileContents("files/secure-forward.conf")),
		},
	)

	logging.addOwnerRefTo(fluentdConfigMap)

	err := sdk.Create(fluentdConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Fluentd configmap: %v", err)
	}

	return nil
}

func createOrUpdateRsyslogConfigMap(logging *logging.ClusterLogging) error {

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
		Rsyslog,
		logging.Namespace,
		rsyslogConfigMapFiles,
	)
	rsyslogConfigMaps[Rsyslog] = rsyslogConfigMap
	for name, cm := range rsyslogConfigMaps {
		utils.AddOwnerRefToObject(cm, utils.AsOwner(logging))

		err = sdk.Create(cm)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing Rsyslog configmap %v: %v", name, err)
		}
	}

	return nil
}

func createOrUpdateFluentdSecret(logging *ClusterLogging) error {

	fluentdSecret := utils.NewSecret(
		Fluentd,
		logging.Namespace,
		map[string][]byte{
			"app-ca":     utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"app-key":    utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.key"),
			"app-cert":   utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.crt"),
			"infra-ca":   utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"infra-key":  utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.key"),
			"infra-cert": utils.GetFileContents("/tmp/_working_dir/system.logging.fluentd.crt"),
		})

	logging.addOwnerRefTo(fluentdSecret)

	err := sdk.Create(fluentdSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Fluentd secret: %v", err)
	}

	return nil
}

func createOrUpdateRsyslogSecret(logging *ClusterLogging) error {

	rsyslogSecret := utils.NewSecret(
		Rsyslog,
		logging.Namespace,
		map[string][]byte{
			"app-ca":     utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"app-key":    utils.GetFileContents("/tmp/_working_dir/system.logging.rsyslog.key"),
			"app-cert":   utils.GetFileContents("/tmp/_working_dir/system.logging.rsyslog.crt"),
			"infra-ca":   utils.GetFileContents("/tmp/_working_dir/ca.crt"),
			"infra-key":  utils.GetFileContents("/tmp/_working_dir/system.logging.rsyslog.key"),
			"infra-cert": utils.GetFileContents("/tmp/_working_dir/system.logging.rsyslog.crt"),
		})

	logging.addOwnerRefTo(rsyslogSecret)

	err := sdk.Create(rsyslogSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Rsyslog secret: %v", err)
	}

	return nil
}

func createOrUpdateFluentdDaemonset(cluster *ClusterLogging) (err error) {

	// TODO iterate the stacks format names based on convention
	// ie. env var APP_HOST=elasticsearch-app; INFRA_HOST=elasticsearch-infra
	elasticSearchAppName := cluster.getElasticsearchName(APP)
	elasticSearchInfraName := cluster.getElasticsearchName(INFRA)
	fluentdPodSpec := newFluentdPodSpec(cluster, elasticSearchAppName, elasticSearchInfraName)

	fluentdDaemonset := utils.NewDaemonSet(Fluentd, cluster.Namespace, Fluentd, Fluentd, fluentdPodSpec)
	cluster.addOwnerRefTo(fluentdDaemonset)

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

func createOrUpdateRsyslogDaemonset(cluster *ClusterLogging) (err error) {

	// TODO iterate the stacks format names based on convention
	// ie. env var APP_HOST=elasticsearch-app; INFRA_HOST=elasticsearch-infra
	elasticSearchAppName := cluster.getElasticsearchName(APP)
	elasticSearchInfraName := cluster.getElasticsearchName(INFRA)
	rsyslogPodSpec := newRsyslogPodSpec(cluster, elasticSearchAppName, elasticSearchInfraName)

	rsyslogDaemonset := utils.NewDaemonSet(Rsyslog, cluster.Namespace, Rsyslog, Rsyslog, rsyslogPodSpec)

	cluster.addOwnerRefTo(rsyslogDaemonset)

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

func newFluentdPodSpec(logging *ClusterLogging, elasticsearchAppName string, elasticsearchInfraName string) v1.PodSpec {

	fluentdContainer := utils.NewContainer("fluentd", v1.PullIfNotPresent, logging.Spec.Collection.LogCollection.FluentdSpec.Resources)

	fluentdContainer.Env = []v1.EnvVar{
		{Name: "MERGE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
		{Name: "ES_HOST", Value: elasticsearchAppName},
		{Name: "ES_PORT", Value: "9200"},
		{Name: "ES_CLIENT_CERT", Value: "/etc/fluent/keys/app-cert"},
		{Name: "ES_CLIENT_KEY", Value: "/etc/fluent/keys/app-key"},
		{Name: "ES_CA", Value: "/etc/fluent/keys/app-ca"},
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
	}

	fluentdContainer.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}

	fluentdPodSpec := utils.NewPodSpec(
		"fluentd",
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
		},
	)

	fluentdPodSpec.PriorityClassName = PriorityClassName

	fluentdPodSpec.NodeSelector = logging.Spec.Collection.LogCollection.FluentdSpec.NodeSelector

	return fluentdPodSpec
}

func newRsyslogPodSpec(logging *ClusterLogging, elasticsearchAppName string, elasticsearchInfraName string) v1.PodSpec {

	rsyslogContainer := utils.NewContainer(Rsyslog, v1.PullIfNotPresent, logging.Spec.Collection.LogCollection.RsyslogSpec.Resources)

	rsyslogContainer.Env = []v1.EnvVar{
		{Name: "MERGE_JSON_LOG", Value: "true"},
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
		"rsyslog",
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
	)

	rsyslogPodSpec.PriorityClassName = PriorityClassName

	rsyslogPodSpec.NodeSelector = logging.Spec.Collection.LogCollection.RsyslogSpec.NodeSelector

	return rsyslogPodSpec
}

func updateFluentdDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
	current := desired.DeepCopy()

	if err = sdk.Get(current); err != nil {
		if apierrors.IsNotFound(err) {
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

func updateRsyslogDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
	current := desired.DeepCopy()

	if err = sdk.Get(current); err != nil {
		if apierrors.IsNotFound(err) {
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

func isDaemonsetDifferent(current *apps.DaemonSet, desired *apps.DaemonSet) (*apps.DaemonSet, bool) {

	different := false

	if isDaemonsetImageDifference(current, desired) {
		logrus.Infof("DaemonSet Image change found, updating %q", current.Name)
		current = updateCurrentDaemonsetImages(current, desired)
		different = true
	}

	return current, different
}

func waitForDaemonSetReady(ds *apps.DaemonSet) error {

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = sdk.Get(ds)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, fmt.Errorf("Failed to get Fluentd daemonset: %v", err)
			}
			return false, err
		}

		if int(ds.Status.DesiredNumberScheduled) == int(ds.Status.NumberReady) {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}

func isDaemonsetImageDifference(current *apps.DaemonSet, desired *apps.DaemonSet) bool {

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

func updateCurrentDaemonsetImages(current *apps.DaemonSet, desired *apps.DaemonSet) *apps.DaemonSet {

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

func isBufferFlushRequired(current *apps.DaemonSet, desired *apps.DaemonSet) bool {

	currImage := strings.Split(current.Spec.Template.Spec.Containers[0].Image, ":")
	desImage := strings.Split(desired.Spec.Template.Spec.Containers[0].Image, ":")

	if len(currImage) != 2 || len(desImage) != 2 {
		// we don't have versions here -- not sure how we would compare versions to determine
		// need to flush buffers
		return false
	}

	currVersion := currImage[1]
	desVersion := desImage[1]

	if strings.HasPrefix(currVersion, "v") {
		currVersion = strings.Split(currVersion, "v")[1]
	}

	if strings.HasPrefix(desVersion, "v") {
		desVersion = strings.Split(desVersion, "v")[1]
	}

	return (currVersion == "3.11" && desVersion == "4.0.0")
}
