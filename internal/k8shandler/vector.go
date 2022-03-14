package k8shandler

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
)

const (
	vectorConfigValue = "/etc/vector"
	dataDir           = "datadir"
	vectorDataDir     = "/var/lib/vector"

	PreviewVectorCollector = "logging.openshift.io/preview-vector-collector"
)

func (clusterRequest *ClusterLoggingRequest) createOrUpdateVectorDaemonset(fluentdTrustBundle *corev1.ConfigMap, pipelineConfHash string) (err error) {
	cluster := clusterRequest.Cluster

	vectorPodSpec := newVectorPodSpec(cluster, fluentdTrustBundle, clusterRequest.ForwarderSpec)
	vectorDaemonset := NewDaemonSet(constants.CollectorName, cluster.Namespace, constants.CollectorName, constants.CollectorName, vectorPodSpec)
	vectorDaemonset.Spec.Template.Spec.Containers[0].Env = updateEnvVar(corev1.EnvVar{Name: "COLLECTOR_CONF_HASH", Value: pipelineConfHash}, vectorDaemonset.Spec.Template.Spec.Containers[0].Env)
	trustedCAHashValue, err := clusterRequest.getTrustedCABundleHash()
	if err != nil {
		return err
	}
	vectorDaemonset.Spec.Template.Annotations[constants.TrustedCABundleHashName] = trustedCAHashValue
	uid := getServiceAccountLogCollectorUID()
	if len(uid) == 0 {
		// There's no uid for logcollector serviceaccount; setting ClusterLogging for the ownerReference.
		utils.AddOwnerRefToObject(vectorDaemonset, utils.AsOwner(cluster))
	} else {
		// There's a uid for logcollector serviceaccount; setting the ServiceAccount for the ownerReference with blockOwnerDeletion.
		utils.AddOwnerRefToObject(vectorDaemonset, NewLogCollectorServiceAccountRef(uid))
	}
	//With this PR: https://github.com/kubernetes-sigs/controller-runtime/pull/919
	//we have got new behaviour: Reset resource version if fake client Create call failed.
	//So if object already exist version will be reset, going to get before try to create.
	ds := &apps.DaemonSet{}
	err = clusterRequest.Get(vectorDaemonset.Name, ds)
	if errors.IsNotFound(err) {
		err = clusterRequest.Create(vectorDaemonset)
		if err != nil {
			return fmt.Errorf("Failure creating collector Daemonset %v", err)
		}
		return nil
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			vectorDaemonset.ResourceVersion = ds.GetResourceVersion()
			return clusterRequest.updateFluentdDaemonsetIfRequired(vectorDaemonset)
		})
		if retryErr != nil {
			return retryErr
		}
	}
	return nil
}
func newVectorPodSpec(cluster *logging.ClusterLogging, trustedCABundleCM *corev1.ConfigMap, pipelineSpec logging.ClusterLogForwarderSpec) corev1.PodSpec {
	resources := &corev1.ResourceRequirements{}
	vectorContainer := NewContainer(constants.CollectorName, constants.VectorName, corev1.PullIfNotPresent, *resources)
	// deliberately not passing any resources for running the below container process, let it have cpu and memory as the process requires
	exporterresources := &corev1.ResourceRequirements{}

	exporterContainer := NewContainer("logfilesmetricexporter", "logfilesmetricexporter", corev1.PullIfNotPresent, *exporterresources)

	vectorContainer.Ports = []corev1.ContainerPort{
		{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	exporterContainer.Ports = []corev1.ContainerPort{
		{
			Name:          exporterPortName,
			ContainerPort: exporterPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	vectorContainer.Env = []corev1.EnvVar{
		{Name: "LOG", Value: "info"},
		{Name: "VECTOR_SELF_NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
		{Name: "METRICS_CERT", Value: "/etc/vector/metrics/tls.crt"},
		{Name: "METRICS_KEY", Value: "/etc/vector/metrics/tls.key"},
		{Name: "NODE_IPV4", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
	}
	proxyEnv := utils.GetProxyEnvVars()
	vectorContainer.Env = append(vectorContainer.Env, proxyEnv...)

	vectorContainer.VolumeMounts = []corev1.VolumeMount{
		{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
		{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
		{Name: logJournal, ReadOnly: true, MountPath: logJournalTransientValue},
		{Name: logAudit, ReadOnly: true, MountPath: logAuditValue},
		{Name: logOvn, ReadOnly: true, MountPath: logOvnValue},
		{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
		{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
		{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
		{Name: config, ReadOnly: true, MountPath: vectorConfigValue},
		{Name: dataDir, ReadOnly: false, MountPath: vectorDataDir},
		{Name: localtime, ReadOnly: true, MountPath: localtimeValue},
		{Name: tmp, MountPath: tmpValue},
	}
	exporterContainer.VolumeMounts = []corev1.VolumeMount{
		{Name: logVolumeMountName, MountPath: logVolumePath},
		{Name: metricsVolumeName, MountPath: metricsVolumeValue},
	}
	// Setting up CMD for log-file-metric-exporter
	exporterContainer.Command = []string{"/usr/local/bin/log-file-metric-exporter", "  -verbosity=2", " -dir=/var/log/containers", " -http=:2112", " -keyFile=/etc/fluent/metrics/tls.key", " -crtFile=/etc/fluent/metrics/tls.crt"}

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
		vectorContainer.VolumeMounts = append(vectorContainer.VolumeMounts, corev1.VolumeMount{Name: name, ReadOnly: true, MountPath: path})
	}

	addTrustedCAVolume := false
	// If trusted CA bundle ConfigMap exists and its hash value is non-zero, mount the bundle.
	if trustedCABundleCM != nil && hasTrustedCABundle(trustedCABundleCM) {
		addTrustedCAVolume = true
		vectorContainer.VolumeMounts = append(vectorContainer.VolumeMounts,
			corev1.VolumeMount{
				Name:      constants.CollectorTrustedCAName,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})
	}
	vectorContainer.SecurityContext = &corev1.SecurityContext{
		SELinuxOptions: &corev1.SELinuxOptions{
			Type: "spc_t",
		},
		ReadOnlyRootFilesystem:   utils.GetBool(true),
		AllowPrivilegeEscalation: utils.GetBool(false),
	}
	exporterContainer.SecurityContext = &corev1.SecurityContext{
		SELinuxOptions: &corev1.SELinuxOptions{
			Type: "spc_t",
		},
		ReadOnlyRootFilesystem:   utils.GetBool(true),
		AllowPrivilegeEscalation: utils.GetBool(false),
	}
	tolerations := utils.AppendTolerations(nil,
		[]corev1.Toleration{
			{
				Key:      "node-role.kubernetes.io/master",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
			{
				Key:      "node.kubernetes.io/disk-pressure",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
		},
	)
	vectorPodSpec := NewPodSpec(
		constants.CollectorServiceAccountName,
		[]corev1.Container{vectorContainer, exporterContainer},
		[]corev1.Volume{
			{Name: logVolumeMountName, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logVolumePath}}},
			{Name: logContainers, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logContainersValue}}},
			{Name: logPods, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logPodsValue}}},
			{Name: logJournal, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logJournalValue}}},
			{Name: logAudit, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logAuditValue}}},
			{Name: logOvn, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logOvnValue}}},
			{Name: logOauthapiserver, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logOauthapiserverValue}}},
			{Name: logOpenshiftapiserver, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logOpenshiftapiserverValue}}},
			{Name: logKubeapiserver, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: logKubeapiserverValue}}},
			{Name: config, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: constants.CollectorConfigSecretName, Optional: utils.GetBool(true)}}},
			{Name: syslogconfig, VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: syslogName}, Optional: utils.GetBool(true)}}},
			{Name: syslogcerts, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: syslogName, Optional: utils.GetBool(true)}}},
			{Name: entrypoint, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: constants.CollectorConfigSecretName, Optional: utils.GetBool(true)}}},
			{Name: certs, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: constants.CollectorName, Optional: utils.GetBool(true)}}},
			{Name: localtime, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: localtimeValue}}},
			{Name: dataDir, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: vectorDataDir}}},
			{Name: metricsVolumeName, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: constants.CollectorMetricSecretName}}},
			{Name: tmp, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory}}},
		},
		map[string]string{},
		tolerations,
	)
	for _, name := range secretNames {
		vectorPodSpec.Volumes = append(vectorPodSpec.Volumes, corev1.Volume{Name: name, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: name}}})
	}

	if addTrustedCAVolume {
		vectorPodSpec.Volumes = append(vectorPodSpec.Volumes,
			corev1.Volume{
				Name: constants.CollectorTrustedCAName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: constants.CollectorTrustedCAName,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  constants.TrustedCABundleKey,
								Path: constants.TrustedCABundleMountFile,
							},
						},
					},
				},
			})
	}
	vectorPodSpec.PriorityClassName = clusterLoggingPriorityClassName
	// Shorten the termination grace period from the default 30 sec to 10 sec.
	vectorPodSpec.TerminationGracePeriodSeconds = utils.GetInt64(10)
	return vectorPodSpec
}
