package k8shandler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/daemonsets"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	fluentdAlertsFile      = "fluentd/fluentd_prometheus_alerts.yaml"
	syslogName             = "syslog"
	logVolumeMountName     = "varlog"
	logVolumePath          = "/var/log"
	logContainers          = "varlogcontainers"
	logContainersValue     = "/var/log/containers"
	logPods                = "varlogpods"
	logPodsValue           = "/var/log/pods"
	logJournal             = "varlogjournal"
	logJournalValue        = "/var/log/journal"
	logAudit               = "varlogaudit"
	logAuditValue          = "/var/log/audit"
	logOvn                 = "varlogovn"
	logOvnValue            = "/var/log/ovn"
	logOauthapiserver      = "varlogoauthapiserver"
	logOauthapiserverValue = "/var/log/oauth-apiserver"
	logOpenshiftapiserver  = "varlogopenshiftapiserver"

	logJournalTransientValue = "/run/log/journal"

	logOpenshiftapiserverValue = "/var/log/openshift-apiserver"
	logKubeapiserver           = "varlogkubeapiserver"
	logKubeapiserverValue      = "/var/log/kube-apiserver"
	config                     = "config"
	configValue                = "/etc/fluent/configs.d/user"
	secureforwardconfig        = "secureforwardconfig"
	secureforwardconfigValue   = "/etc/fluent/configs.d/secure-forward"
	secureforwardcerts         = "secureforwardcerts"
	secureforwardcertsValue    = "/etc/ocp-forward"
	syslogconfig               = "syslogconfig"
	syslogconfigValue          = "/etc/fluent/configs.d/syslog"
	syslogcerts                = "syslogcerts"
	syslogcertsValue           = "/etc/ocp-syslog"
	entrypoint                 = "entrypoint"
	entrypointValue            = "/opt/app-root/src/run.sh"
	certs                      = "certs"
	certsValue                 = "/etc/fluent/keys"
	localtime                  = "localtime"
	localtimeValue             = "/etc/localtime"
	filebufferstorage          = "filebufferstorage"
	filebufferstorageValue     = "/var/lib/fluentd"
	metricsVolumeValue         = "/etc/fluent/metrics"
	tmp                        = "tmp"
	tmpValue                   = "/tmp"
)

// useOldRemoteSyslogPlugin checks if old plugin (docebo/fluent-plugin-remote-syslog) is to be used for sending syslog or new plugin (dlackty/fluent-plugin-remote_syslog) is to be used
func (clusterRequest *ClusterLoggingRequest) useOldRemoteSyslogPlugin() bool {
	if clusterRequest.ForwarderRequest == nil {
		return false
	}
	enabled, found := clusterRequest.ForwarderRequest.Annotations[UseOldRemoteSyslogPlugin]
	return found && enabled == "enabled"
}

func newFluentdPodSpec(cluster *logging.ClusterLogging, trustedCABundleCM *v1.ConfigMap, clfspec logging.ClusterLogForwarderSpec, secrets map[string]*v1.Secret) v1.PodSpec {
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
	fluentdContainer := NewContainer(constants.CollectorName, constants.FluentdName, v1.PullIfNotPresent, *resources)
	// deliberately not passing any resources for running the below container process, let it have cpu and memory as the process requires
	exporterresources := &v1.ResourceRequirements{}

	exporterContainer := NewContainer("logfilesmetricexporter", "logfilesmetricexporter", v1.PullIfNotPresent, *exporterresources)

	fluentdContainer.Ports = []v1.ContainerPort{
		{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	exporterContainer.Ports = []v1.ContainerPort{
		{
			Name:          exporterPortName,
			ContainerPort: exporterPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	fluentdContainer.Env = []v1.EnvVar{
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
					fluentdContainer.Env = append(fluentdContainer.Env, v1.EnvVar{Name: "BUFFER_SIZE_LIMIT", Value: strconv.FormatInt(chunkLimitSize.Value(), 10)})
				}
			}
			if cluster.Spec.Forwarder.Fluentd.Buffer.TotalLimitSize != "" {
				if totalLimitSize, err := utils.ParseQuantity(string(cluster.Spec.Forwarder.Fluentd.Buffer.TotalLimitSize)); err == nil {
					fluentdContainer.Env = append(fluentdContainer.Env, v1.EnvVar{Name: "TOTAL_LIMIT_SIZE_PER_BUFFER", Value: strconv.FormatInt(totalLimitSize.Value(), 10)})
				}
			}

		}
	}

	proxyEnv := utils.GetProxyEnvVars()
	fluentdContainer.Env = append(fluentdContainer.Env, proxyEnv...)

	fluentdContainer.VolumeMounts = []v1.VolumeMount{
		{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
		{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
		{Name: logJournal, ReadOnly: true, MountPath: logJournalValue},
		{Name: logAudit, ReadOnly: true, MountPath: logAuditValue},
		{Name: logOvn, ReadOnly: true, MountPath: logOvnValue},
		{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
		{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
		{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
		{Name: config, ReadOnly: true, MountPath: configValue},
		{Name: secureforwardconfig, ReadOnly: true, MountPath: secureforwardconfigValue},
		{Name: secureforwardcerts, ReadOnly: true, MountPath: secureforwardcertsValue},
		{Name: syslogconfig, ReadOnly: true, MountPath: syslogconfigValue},
		{Name: syslogcerts, ReadOnly: true, MountPath: syslogcertsValue},
		{Name: entrypoint, ReadOnly: true, MountPath: entrypointValue, SubPath: "run.sh"},
		{Name: certs, ReadOnly: true, MountPath: certsValue},
		{Name: localtime, ReadOnly: true, MountPath: localtimeValue},
		{Name: filebufferstorage, MountPath: filebufferstorageValue},
		{Name: metricsVolumeName, ReadOnly: true, MountPath: metricsVolumeValue},
		{Name: tmp, MountPath: tmpValue},
	}

	exporterContainer.VolumeMounts = []v1.VolumeMount{
		{Name: logVolumeMountName, MountPath: logVolumePath},
		{Name: metricsVolumeName, MountPath: metricsVolumeValue},
	}
	// Setting up CMD for log-file-metric-exporter
	exporterContainer.Command = []string{"/usr/local/bin/log-file-metric-exporter", "  -verbosity=2", " -dir=/var/log/containers", " -http=:2112", " -keyFile=/etc/fluent/metrics/tls.key", " -crtFile=/etc/fluent/metrics/tls.crt"}

	// List of _unique_ output secret names, several outputs may use the same secret.
	unique := sets.NewString()
	for _, o := range clfspec.Outputs {
		if o.Secret != nil && o.Secret.Name != "" {
			unique.Insert(o.Secret.Name)
		}
	}
	secretNames := unique.List()

	for _, name := range secretNames {
		path := fmt.Sprintf("/var/run/ocp-collector/secrets/%s", name)
		fluentdContainer.VolumeMounts = append(fluentdContainer.VolumeMounts, v1.VolumeMount{Name: name, ReadOnly: true, MountPath: path})
	}

	addTrustedCAVolume := false
	// If trusted CA bundle ConfigMap exists and its hash value is non-zero, mount the bundle.
	if trustedCABundleCM != nil && hasTrustedCABundle(trustedCABundleCM) {
		addTrustedCAVolume = true
		fluentdContainer.VolumeMounts = append(fluentdContainer.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.CollectorTrustedCAName,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})
	}

	// Append any additional volumes based on attributes of the secret and forwarder spec
	addProjectServiceAccountVolume := false
	if CloudwatchSecretWithRoleArnKey(secrets, &clfspec) {
		addProjectServiceAccountVolume = true
		fluentdContainer.VolumeMounts = append(fluentdContainer.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.AWSWebIdentityTokenName,
				ReadOnly:  true,
				MountPath: constants.AWSWebIdentityTokenMount,
			})
	}

	fluentdContainer.SecurityContext = &v1.SecurityContext{
		SELinuxOptions: &v1.SELinuxOptions{
			Type: "spc_t",
		},
		ReadOnlyRootFilesystem:   utils.GetBool(true),
		AllowPrivilegeEscalation: utils.GetBool(false),
	}
	exporterContainer.SecurityContext = &v1.SecurityContext{
		SELinuxOptions: &v1.SELinuxOptions{
			Type: "spc_t",
		},
		ReadOnlyRootFilesystem:   utils.GetBool(true),
		AllowPrivilegeEscalation: utils.GetBool(false),
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

	fluentdPodSpec := NewPodSpec(
		constants.CollectorServiceAccountName,
		[]v1.Container{fluentdContainer, exporterContainer},
		[]v1.Volume{
			{Name: logVolumeMountName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logVolumePath}}},
			{Name: logContainers, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logContainersValue}}},
			{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}},
			{Name: logJournal, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logJournalValue}}},
			{Name: logAudit, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logAuditValue}}},
			{Name: logOvn, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOvnValue}}},
			{Name: logOauthapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOauthapiserverValue}}},
			{Name: logOpenshiftapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOpenshiftapiserverValue}}},
			{Name: logKubeapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logKubeapiserverValue}}},
			{Name: config, VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: constants.CollectorName}}}},
			{Name: secureforwardconfig, VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "secure-forward"}, Optional: utils.GetBool(true)}}},
			{Name: secureforwardcerts, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "secure-forward", Optional: utils.GetBool(true)}}},
			{Name: syslogconfig, VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: syslogName}, Optional: utils.GetBool(true)}}},
			{Name: syslogcerts, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: syslogName, Optional: utils.GetBool(true)}}},
			{Name: entrypoint, VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: constants.CollectorName}}}},
			{Name: certs, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: constants.CollectorName, Optional: utils.GetBool(true)}}},
			{Name: localtime, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: localtimeValue}}},
			{Name: filebufferstorage, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: filebufferstorageValue}}},
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: constants.CollectorMetricSecretName}}},
			{Name: tmp, VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory}}},
		},
		collectionSpec.Logs.FluentdSpec.NodeSelector,
		tolerations,
	)
	for _, name := range secretNames {
		fluentdPodSpec.Volumes = append(fluentdPodSpec.Volumes, v1.Volume{Name: name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: name}}})
	}

	if addTrustedCAVolume {
		fluentdPodSpec.Volumes = append(fluentdPodSpec.Volumes,
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

	if addProjectServiceAccountVolume {
		fluentdPodSpec.Volumes = append(fluentdPodSpec.Volumes,
			v1.Volume{
				Name: constants.AWSWebIdentityTokenName,
				VolumeSource: v1.VolumeSource{
					Projected: &v1.ProjectedVolumeSource{
						Sources: []v1.VolumeProjection{
							{
								ServiceAccountToken: &v1.ServiceAccountTokenProjection{
									Audience: "openshift",
									Path:     constants.AWSWebIdentityTokenFilePath,
								},
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

func (clusterRequest *ClusterLoggingRequest) createOrUpdateFluentdDaemonset(fluentdTrustBundle *v1.ConfigMap, pipelineConfHash string) (err error) {

	cluster := clusterRequest.Cluster

	fluentdPodSpec := newFluentdPodSpec(cluster, fluentdTrustBundle, clusterRequest.ForwarderSpec, clusterRequest.OutputSecrets)

	fluentdDaemonset := NewDaemonSet(constants.CollectorName, cluster.Namespace, constants.CollectorName, constants.CollectorName, fluentdPodSpec)
	fluentdDaemonset.Spec.Template.Spec.Containers[0].Env = updateEnvVar(v1.EnvVar{Name: "COLLECTOR_CONF_HASH", Value: pipelineConfHash}, fluentdDaemonset.Spec.Template.Spec.Containers[0].Env)

	trustedCAHashValue, err := clusterRequest.getTrustedCABundleHash()
	if err != nil {
		return err
	}
	fluentdDaemonset.Spec.Template.Annotations[constants.TrustedCABundleHashName] = trustedCAHashValue

	uid := getServiceAccountLogCollectorUID()
	if len(uid) == 0 {
		// There's no uid for logcollector serviceaccount; setting ClusterLogging for the ownerReference.
		utils.AddOwnerRefToObject(fluentdDaemonset, utils.AsOwner(cluster))
	} else {
		// There's a uid for logcollector serviceaccount; setting the ServiceAccount for the ownerReference with blockOwnerDeletion.
		utils.AddOwnerRefToObject(fluentdDaemonset, NewLogCollectorServiceAccountRef(uid))
	}

	//With this PR: https://github.com/kubernetes-sigs/controller-runtime/pull/919
	//we have got new behaviour: Reset resource version if fake client Create call failed.
	//So if object already exist version will be reset, going to get before try to create.
	ds := &apps.DaemonSet{}
	err = clusterRequest.Get(fluentdDaemonset.Name, ds)
	if errors.IsNotFound(err) {
		err = clusterRequest.Create(fluentdDaemonset)
		if err != nil {
			return fmt.Errorf("Failure creating collector Daemonset %v", err)
		}
		return nil
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			fluentdDaemonset.ResourceVersion = ds.GetResourceVersion()
			return clusterRequest.updateFluentdDaemonsetIfRequired(fluentdDaemonset)
		})
		if retryErr != nil {
			return retryErr
		}
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) updateFluentdDaemonsetIfRequired(desired *apps.DaemonSet) (err error) {
	current := &apps.DaemonSet{}

	if err = clusterRequest.Get(desired.Name, current); err != nil {
		if errors.IsNotFound(err) {
			// the object doesn't exist -- it was likely culled
			// recreate it on the next time through if necessary
			return nil
		}
		return fmt.Errorf("Failed to get collector daemonset: %v", err)
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
				log.V(2).Error(err, "Failed to prepare collector daemonset to flush its buffers")
				return err
			}

			// wait for pods to all restart then continue
			if err = clusterRequest.waitForDaemonSetReady(current); err != nil {
				return fmt.Errorf("Timed out waiting for collector to be ready")
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

	fluentdTrustBundle := &v1.ConfigMap{}
	fluentdTrustBundleName := types.NamespacedName{Name: constants.CollectorTrustedCAName, Namespace: constants.OpenshiftNS}
	if err := clusterRequest.Client.Get(context.TODO(), fluentdTrustBundleName, fluentdTrustBundle); err != nil {
		if !errors.IsNotFound(err) {
			return "", err
		}
	}

	if _, ok := fluentdTrustBundle.Data[constants.TrustedCABundleKey]; !ok {
		log.V(1).Info("Cluster wide proxy may not be configured. ConfigMap does not contain expected key", "configmapName", fluentdTrustBundle.Name, "key", constants.TrustedCABundleKey)
		return "", nil
	}

	trustedCAHashValue, err := calcTrustedCAHashValue(fluentdTrustBundle)
	if err != nil {
		return "", fmt.Errorf("unable to calculate trusted CA value. E: %s", err.Error())
	}

	if trustedCAHashValue == "" {
		log.V(1).Info("Cluster wide proxy may not be configured. ConfigMap does not contain a ca bundle", "configmapName", fluentdTrustBundle.Name)
		return "", nil
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
	var collectorType logging.LogCollectionType = clusterRequest.Cluster.Spec.Collection.Logs.Type

	if err = clusterRequest.createOrUpdateCollectorDaemonset(collectorType, collectorConfHash); err != nil {
		return
	}

	return clusterRequest.UpdateCollectorStatus(collectorType)
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

// CloudwatchSecretWithRoleArnKey return true if secret has 'role_arn' key and output type is cloudwatch
func CloudwatchSecretWithRoleArnKey(secrets map[string]*v1.Secret, clfspec *logging.ClusterLogForwarderSpec) bool {
	if secrets == nil {
		return false
	}
	for _, o := range clfspec.Outputs {
		secret := secrets[o.Name]
		if security.HasAwsRoleArnKey(secret) && o.Type == logging.OutputTypeCloudwatch {
			return true
		}
	}
	return false
}
