package k8shandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	"github.com/openshift/cluster-logging-operator/pkg/utils/comparators/servicemonitor"

	"github.com/ViaQ/logerr/log"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/components/fluentd"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/factory"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/pkg/utils/comparators/daemonsets"
	"github.com/openshift/cluster-logging-operator/pkg/utils/comparators/services"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	fluentdAlertsFile        = "fluentd/fluentd_prometheus_alerts.yaml"
	fluentdName              = "fluentd"
	syslogName               = "syslog"
	fluentdRequiredESVersion = "6"
)

func (clusterRequest *ClusterLoggingRequest) removeCollector(name string) (err error) {
	if clusterRequest.isManaged() {

		if err = clusterRequest.RemoveDaemonset(name); err != nil {
			log.V(3).Error(err, "")
			return
		}
		// Wait longer than the terminationGracePeriodSeconds
		time.Sleep(12 * time.Second)

		if err = clusterRequest.RemoveSecret(name); err != nil {
			log.V(3).Error(err, "")
			return
		}
		if err = clusterRequest.RemoveConfigMap(name); err != nil {
			log.V(3).Error(err, "")
		}
		if err = clusterRequest.RemovePrometheusRule(name); err != nil {
			log.V(3).Error(err, "")
		}
		if err = clusterRequest.RemoveServiceMonitor(name); err != nil {
			log.V(3).Error(err, "")
		}
		if err = clusterRequest.RemoveService(name); err != nil {
			log.V(3).Error(err, "")
		}
		if err = clusterRequest.RemoveServiceAccount(name); err != nil {
			log.V(3).Error(err, "")
		}
		caName := fmt.Sprintf("%s-trusted-ca-bundle", name)
		if err = clusterRequest.RemoveConfigMap(caName); err != nil {
			log.V(3).Error(err, "")
			return
		}

	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) reconcileCollectorService() error {
	desired := factory.NewService(
		constants.CollectorName,
		clusterRequest.Cluster.Namespace,
		constants.CollectorName,
		[]v1.ServicePort{
			{
				Port:       metricsPort,
				TargetPort: intstr.FromString(metricsPortName),
				Name:       metricsPortName,
			},
		},
	)

	desired.Annotations = map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": constants.CollectorMetricSecretName,
	}

	utils.AddOwnerRefToObject(desired, utils.AsOwner(clusterRequest.Cluster))
	err := clusterRequest.Create(desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failure creating the #{desired.Name} service: #{err}")
		}
		current := &v1.Service{}
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(desired.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("failed to get %s service for %s: %v", desired.Name, clusterRequest.Cluster.Name, err)
			}
			if services.AreSame(current, desired) {
				log.V(3).Info("Services are the same skipping update")
				return nil
			}
			//Explicitly copying because services are immutable
			current.Labels = desired.Labels
			current.Spec.Selector = desired.Spec.Selector
			current.Spec.Ports = desired.Spec.Ports
			return clusterRequest.Update(current)
		})
		log.V(3).Error(retryErr, "Reconcile Service retry error")
		return retryErr
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) reconcileCollectorServiceMonitor() error {

	cluster := clusterRequest.Cluster

	desired := NewServiceMonitor(constants.CollectorName, cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   metricsPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", constants.CollectorName, cluster.Namespace),
			// ServerName can be e.g. fluentd.openshift-logging.svc
		},
	}

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"logging-infra": "support",
		},
	}

	desired.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  constants.CollectorMonitorJobLabel,
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{cluster.Namespace},
		},
	}

	utils.AddOwnerRefToObject(desired, utils.AsOwner(cluster))

	err := clusterRequest.Create(desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failure creating the %s ServiceMonitor: %v", desired.Name, err)
		}
		current := &monitoringv1.ServiceMonitor{}
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(desired.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %q service for %q: %v", current.Name, clusterRequest.Cluster.Name, err)
			}
			if servicemonitor.AreSame(current, desired) {
				log.V(3).Info("ServiceMonitor are the same skipping update")
				return nil
			}
			current.Labels = desired.Labels
			current.Spec = desired.Spec
			current.Annotations = desired.Annotations

			return clusterRequest.Update(current)
		})
		log.V(3).Error(retryErr, "Reconcile ServiceMonitor retry error")
		return retryErr
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdPrometheusRule() error {
	ctx := context.TODO()
	cluster := clusterRequest.Cluster

	rule := NewPrometheusRule(constants.CollectorName, cluster.Namespace)

	spec, err := NewPrometheusRuleSpecFrom(utils.GetShareDir() + "/" + fluentdAlertsFile)
	if err != nil {
		return fmt.Errorf("failure creating the %s PrometheusRule: %w", rule.Name, err)
	}

	rule.Spec = *spec

	utils.AddOwnerRefToObject(rule, utils.AsOwner(cluster))

	err = clusterRequest.Create(rule)
	if err == nil {
		return nil
	}
	if !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failure creating the %s PrometheusRule: %w", rule.Name, err)
	}

	current := &monitoringv1.PrometheusRule{}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err = clusterRequest.Client.Get(ctx, types.NamespacedName{Name: rule.Name, Namespace: rule.Namespace}, current)
		if err != nil {
			log.V(2).Info("could not get prometheus rule", rule.Name, err)
			return err
		}
		current.Spec = rule.Spec
		if err = clusterRequest.Client.Update(ctx, current); err != nil {
			return err
		}
		log.V(3).Info("updated prometheus rules")
		return nil
	})
}

// includeLegacyForwardConfig to address Bug 1782566.
// To be removed when legacy forwarding is unsupported
func (clusterRequest *ClusterLoggingRequest) includeLegacyForwardConfig() bool {
	if clusterRequest.FnIncludeLegacyForward != nil {
		return clusterRequest.FnIncludeLegacySyslog()
	}
	config := &v1.ConfigMap{
		Data: map[string]string{},
	}
	if err := clusterRequest.Get("secure-forward", config); err != nil {
		if errors.IsNotFound(err) {
			return false
		}
		log.Info("There was a non-critical error trying to fetch the secure-forward configmap", "error", err.Error())
	}
	_, found := config.Data["secure-forward.conf"]
	return found
}

// includeLegacySyslogConfig to address Bug 1799024.
// To be removed when legacy syslog is no longer supported.
func (clusterRequest *ClusterLoggingRequest) includeLegacySyslogConfig() bool {
	if clusterRequest.FnIncludeLegacySyslog != nil {
		return clusterRequest.FnIncludeLegacySyslog()
	}
	config := &v1.ConfigMap{
		Data: map[string]string{},
	}
	if err := clusterRequest.Get(syslogName, config); err != nil {
		if errors.IsNotFound(err) {
			return false
		}
		log.Info("There was a non-critical error trying to fetch the configmap", "error", err.Error())
	}
	_, found := config.Data["syslog.conf"]
	return found
}

// useOldRemoteSyslogPlugin checks if old plugin (docebo/fluent-plugin-remote-syslog) is to be used for sending syslog or new plugin (dlackty/fluent-plugin-remote_syslog) is to be used
func (clusterRequest *ClusterLoggingRequest) useOldRemoteSyslogPlugin() bool {
	if clusterRequest.ForwarderRequest == nil {
		return false
	}
	enabled, found := clusterRequest.ForwarderRequest.Annotations[UseOldRemoteSyslogPlugin]
	return found && enabled == "enabled"
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorConfigMap(config string) error {
	configMap := NewConfigMap(
		constants.CollectorName,
		clusterRequest.Cluster.Namespace,
		map[string]string{
			"fluent.conf": config,
			"run.sh":      fluentd.RunScript,
		},
	)

	utils.AddOwnerRefToObject(configMap, utils.AsOwner(clusterRequest.Cluster))

	err := clusterRequest.Create(configMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Fluentd configmap: %v", err)
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &v1.ConfigMap{}
		if err = clusterRequest.Get(configMap.Name, current); err != nil {
			if errors.IsNotFound(err) {
				log.V(2).Info("Returning nil. The configmap was not found even though create previously failed.  Was it culled?", "configmap name", configMap.Name)
				return nil
			}
			return fmt.Errorf("Failed to get %v configmap for %q: %v", configMap.Name, clusterRequest.Cluster.Name, err)
		}
		if reflect.DeepEqual(configMap.Data, current.Data) {
			return nil
		}
		current.Data = configMap.Data
		return clusterRequest.Update(current)
	})

	return retryErr
}

//createOrUpdateCollectorSecret reconciles the secrets needed to communicate with the default
//log storage (e.g. Elasticsearch).
//TODO: Remove me once default storage is able to provide certs when requested by this operator
func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorSecret() error {
	var secrets = map[string][]byte{}
	_ = Syncronize(func() error {
		secrets = map[string][]byte{}
		// Ignore errors, these files are optional depending on security settings.
		secrets[constants.TrustedCABundleKey], _ = ioutil.ReadFile(utils.GetWorkingDirFilePath(constants.GeneratedCAFileName))
		secrets[constants.ClientPrivateKey], _ = ioutil.ReadFile(utils.GetWorkingDirFilePath(constants.GeneratedKeyFileName))
		secrets[constants.ClientCertKey], _ = ioutil.ReadFile(utils.GetWorkingDirFilePath(constants.GeneratedCertFileName))
		return nil
	})
	secret := NewSecret(
		constants.CollectorName,
		clusterRequest.Cluster.Namespace,
		secrets)

	utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.Cluster))

	err := clusterRequest.CreateOrUpdateSecret(secret)
	if err != nil {
		return err
	}

	return nil
}

func newFluentdPodSpec(cluster *logging.ClusterLogging, trustedCABundleCM *v1.ConfigMap, pipelineSpec logging.ClusterLogForwarderSpec) v1.PodSpec {
	collectionSpec := logging.CollectionSpec{}
	if cluster.Spec.Collection != nil {
		collectionSpec = *cluster.Spec.Collection
	}
	resources := collectionSpec.Logs.FluentdSpec.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultFluentdMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultFluentdMemory,
				v1.ResourceCPU:    defaultFluentdCpuRequest,
			},
		}
	}
	collectorContainer := NewContainer(constants.CollectorName, constants.CollectorName, v1.PullIfNotPresent, *resources)

	collectorContainer.Ports = []v1.ContainerPort{
		{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	collectorContainer.Env = []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
		{Name: "METRICS_CERT", Value: "/etc/fluent/metrics/tls.crt"},
		{Name: "METRICS_KEY", Value: "/etc/fluent/metrics/tls.key"},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
		{Name: "POD_IP", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
	}

	if cluster.Spec.Forwarder != nil {
		if cluster.Spec.Forwarder.Fluentd.Buffer != nil {
			if cluster.Spec.Forwarder.Fluentd.Buffer.ChunkLimitSize != "" {
				if chunkLimitSize, err := utils.ParseQuantity(string(cluster.Spec.Forwarder.Fluentd.Buffer.ChunkLimitSize)); err == nil {
					collectorContainer.Env = append(collectorContainer.Env, v1.EnvVar{Name: "BUFFER_SIZE_LIMIT", Value: strconv.FormatInt(chunkLimitSize.Value(), 10)})
				}
			}
		}
	}

	proxyEnv := utils.GetProxyEnvVars()
	collectorContainer.Env = append(collectorContainer.Env, proxyEnv...)

	collectorContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "varlog", MountPath: "/var/log"},
		{Name: "varlibdockercontainers", ReadOnly: true, MountPath: "/var/lib/docker"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/fluent/configs.d/user"},
		{Name: "secureforwardconfig", ReadOnly: true, MountPath: "/etc/fluent/configs.d/secure-forward"},
		{Name: "secureforwardcerts", ReadOnly: true, MountPath: "/etc/ocp-forward"},
		{Name: "syslogconfig", ReadOnly: true, MountPath: "/etc/fluent/configs.d/syslog"},
		{Name: "syslogcerts", ReadOnly: true, MountPath: "/etc/ocp-syslog"},
		{Name: "entrypoint", ReadOnly: true, MountPath: "/opt/app-root/src/run.sh", SubPath: "run.sh"},
		{Name: "certs", ReadOnly: true, MountPath: "/etc/fluent/keys"},
		{Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
		{Name: "dockercfg", ReadOnly: true, MountPath: "/etc/sysconfig/docker"},
		{Name: "dockerdaemoncfg", ReadOnly: true, MountPath: "/etc/docker"},
		{Name: "filebufferstorage", MountPath: "/var/lib/fluentd"},
		{Name: metricsVolumeName, MountPath: "/etc/fluent/metrics"},
	}

	// List of _unique_ output secret names, several outputs may use the same secret.
	unique := sets.NewString()
	for _, o := range pipelineSpec.Outputs {
		if o.Secret != nil && o.Secret.Name != "" {
			unique.Insert(o.Secret.Name)
		}
	}
	secretNames := unique.List()

	for _, name := range secretNames {
		path := fmt.Sprintf("/var/run/ocp-collector/secrets/%s", name)
		collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts, v1.VolumeMount{Name: name, MountPath: path})
	}

	addTrustedCAVolume := false
	// If trusted CA bundle ConfigMap exists and its hash value is non-zero, mount the bundle.
	if trustedCABundleCM != nil && hasTrustedCABundle(trustedCABundleCM) {
		addTrustedCAVolume = true
		collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.CollectorTrustedCAName,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})
	}

	collectorContainer.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}

	tolerations := utils.AppendTolerations(
		collectionSpec.Logs.FluentdSpec.Tolerations,
		[]v1.Toleration{
			{
				Key:      "node-role.kubernetes.io/master",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
			{
				Key:      "node.kubernetes.io/disk-pressure",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	)

	podSpec := NewPodSpec(
		"logcollector",
		[]v1.Container{collectorContainer},
		[]v1.Volume{
			{Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
			{Name: "varlibdockercontainers", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker"}}},
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: constants.CollectorName}}}},
			{Name: "secureforwardconfig", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "secure-forward"}, Optional: utils.GetBool(true)}}},
			{Name: "secureforwardcerts", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "secure-forward", Optional: utils.GetBool(true)}}},
			{Name: "syslogconfig", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: syslogName}, Optional: utils.GetBool(true)}}},
			{Name: "syslogcerts", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: syslogName, Optional: utils.GetBool(true)}}},
			{Name: "entrypoint", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: constants.CollectorName}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: constants.CollectorName, Optional: utils.GetBool(true)}}},
			{Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
			{Name: "dockercfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/sysconfig/docker"}}},
			{Name: "dockerdaemoncfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/docker"}}},
			{Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/fluentd"}}},
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: constants.CollectorMetricSecretName}}},
		},
		collectionSpec.Logs.FluentdSpec.NodeSelector,
		tolerations,
	)
	for _, name := range secretNames {
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{Name: name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: name}}})
	}

	if addTrustedCAVolume {
		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{
				Name: constants.CollectorTrustedCAName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: constants.CollectorTrustedCAName,
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

	podSpec.PriorityClassName = clusterLoggingPriorityClassName
	// Shorten the termination grace period from the default 30 sec to 10 sec.
	podSpec.TerminationGracePeriodSeconds = utils.GetInt64(10)

	// FIXME: The following conditional branch is a refactoring candidate
	// To address here https://issues.redhat.com/browse/LOG-833
	//
	// Why do we need this branch at all?
	//
	// The ClusterLogging CR supports three additional use cases without providing
	// a logStore stanza namely:
	// Case 1: A collection only stanza to use with secure forward
	// Case 2: A collection only stanza to use with syslog
	// Case 3: A collection only stanza to use with ClusterLogFowarder API
	//
	// Supporting all three cases implies:
	// 1. We don't want to add the init container if `cluster.Spec.LogStore == nil`.
	// 2. We don't want to add the init container if `cluster.Spec.LogStore == nil`
	//    and a ClusterLogFowarder CR and does not provide a default output (only
	//    relevant for case 3)
	//
	// What is the init container bound to a log store stanza?
	//
	// The init container implementation is only relevant for the default ES log store.
	// It checks for the version of an ES cluster being >=6. In upgrade scenarios
	// an ES cluster can eventually provide nodes with multiple version (5.x and 6.x)
	// until all nodes are upgraded. The data model used in 6.x is not backward
	// compatible to 5.x, thus preventing fluentd from starting avoids data model
	// inconsistencies during ES upgrades from 5.x to 6.x.
	//
	// Why does this work w/o a ClusterLogForwarder CR provided?
	//
	// ClusterLogging and ClusterLogFowarder CR reconciliation share the same logic
	// for collection config generation. If first is provided without the latter a
	// ClusterLogFowarder with default fields is assumed.
	// (See ClusterLoggingRequest#getLogForwarder)
	if pipelineSpec.HasDefaultOutput() && cluster.Spec.LogStore != nil && cluster.Spec.LogStore.ElasticsearchSpec.NodeCount > 0 {
		podSpec.InitContainers = []v1.Container{
			newFluentdInitContainer(cluster),
		}
	} else {
		podSpec.InitContainers = []v1.Container{}
	}

	return podSpec
}

func newFluentdInitContainer(cluster *logging.ClusterLogging) v1.Container {
	collectionSpec := logging.CollectionSpec{}
	resources := collectionSpec.Logs.FluentdSpec.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultFluentdMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultFluentdMemory,
				v1.ResourceCPU:    defaultFluentdCpuRequest,
			},
		}
	}
	initContainer := NewContainer("fluentd-init", "fluentd", v1.PullIfNotPresent, *resources)

	initContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "certs", ReadOnly: true, MountPath: "/etc/fluent/keys"},
	}

	initContainer.Command = []string{
		"./wait_for_es_version.sh",
		fluentdRequiredESVersion,
		fmt.Sprintf("https://%s.%s.svc:9200", elasticsearchResourceName, cluster.Namespace),
	}

	return initContainer
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorDaemonset(pipelineConfHash string) (err error) {

	cluster := clusterRequest.Cluster

	var trustBundle *v1.ConfigMap
	// Create or update cluster proxy trusted CA bundle.
	trustBundle, err = clusterRequest.createOrGetTrustedCABundleConfigMap(constants.CollectorTrustedCAName)
	if err != nil {
		return
	}

	podSpec := newFluentdPodSpec(cluster, trustBundle, clusterRequest.ForwarderSpec)

	daemonset := NewDaemonSet(constants.CollectorName, cluster.Namespace, constants.CollectorName, constants.CollectorName, podSpec)
	daemonset.Spec.Template.Spec.Containers[0].Env = updateEnvVar(v1.EnvVar{Name: "CONFIG_HASH", Value: pipelineConfHash}, daemonset.Spec.Template.Spec.Containers[0].Env)

	trustedCAHashValue, err := clusterRequest.getTrustedCABundleHash()
	if err != nil {
		return err
	}
	daemonset.Spec.Template.Annotations[constants.TrustedCABundleHashName] = trustedCAHashValue

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
		return fmt.Errorf("failure creating %s Daemonset: %v", daemonset.Name, err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateCollectorDaemonsetIfRequired(daemonset)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) updateCollectorDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
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
	if flushBuffer {
		current.Spec.Template.Spec.Containers[0].Env = updateEnvVar(v1.EnvVar{Name: "FLUSH_AT_SHUTDOWN", Value: "True"}, current.Spec.Template.Spec.Containers[0].Env)
	}
	trustedCABundleHashAreSame := current.Spec.Template.Annotations[constants.TrustedCABundleHashName] == desired.Spec.Template.Annotations[constants.TrustedCABundleHashName]
	if !daemonsets.AreSame(current, desired) || !trustedCABundleHashAreSame {
		log.V(3).Info("Current and desired collectors are different, updating DaemonSet", "DaemonSet", current.Name)
		if flushBuffer {
			log.Info("Updating and restarting collector pods to flush its buffers...")
			if err = clusterRequest.Update(current); err != nil {
				log.V(2).Error(err, "Failed to prepare Fluentd daemonset to flush its buffers")
				return err
			}

			// wait for pods to all restart then continue
			if err = clusterRequest.waitForDaemonSetReady(current); err != nil {
				return fmt.Errorf("Timed out waiting for Fluentd to be ready")
			}
		}
		current.Spec = desired.Spec
		if err = clusterRequest.Update(desired); err != nil {
			return err
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) getTrustedCABundleHash() (string, error) {

	trustBundle := &v1.ConfigMap{}
	trustBundleName := types.NamespacedName{Name: constants.CollectorTrustedCAName, Namespace: constants.OpenshiftNS}
	if err := clusterRequest.Client.Get(context.TODO(), trustBundleName, trustBundle); err != nil {
		if !errors.IsNotFound(err) {
			return "", err
		}
	}

	if _, ok := trustBundle.Data[constants.TrustedCABundleKey]; !ok {
		return "", fmt.Errorf("%v does not yet contain expected key %v", trustBundle.Name, constants.TrustedCABundleKey)
	}

	trustedCAHashValue, err := calcTrustedCAHashValue(trustBundle)
	if err != nil {
		return "", fmt.Errorf("unable to calculate trusted CA value. E: %s", err.Error())
	}

	if trustedCAHashValue == "" {
		return "", fmt.Errorf("Did not receive hashvalue for trusted CA value")
	}

	return trustedCAHashValue, nil
}

func (clusterRequest *ClusterLoggingRequest) RestartCollector() (err error) {

	collectorConfig, err := clusterRequest.generateCollectorConfig()
	if err != nil {
		return err
	}

	log.V(3).Info("Generated collector config", "config", collectorConfig)
	collectorConfHash, err := utils.CalculateMD5Hash(collectorConfig)
	if err != nil {
		log.Error(err, "unable to calculate MD5 hash.")
		return
	}

	if err = clusterRequest.createOrUpdateCollectorDaemonset(collectorConfHash); err != nil {
		return
	}

	return clusterRequest.UpdateCollectorStatus()
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
