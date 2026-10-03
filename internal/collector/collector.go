package collector

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	//DefaultMaxUnavailable is the maxUnavailable collector setting when not defined by spec.collector.maxUnavailable
	DefaultMaxUnavailable = "100%"

	defaultAudience                            = "openshift"
	clusterLoggingPriorityClassName            = "system-node-critical"
	metricsVolumeName                          = "metrics"
	metricsVolumePath                          = "/etc/collector/metrics"
	saTokenVolumeName                          = "sa-token"
	saTokenExpirationSecs                      = 3600 //1 hour
	defaultTerminationGracePeriodSeconds int64 = 10
	sourcePodsName                             = "varlogpods"
	sourcePodsPath                             = "/var/log/pods"
	sourceJournalName                          = "varlogjournal"
	sourceJournalPath                          = "/var/log/journal"
	sourceAuditdName                           = "varlogaudit"
	sourceAuditdPath                           = "/var/log/audit"
	sourceAuditOVNName                         = "varlogovn"
	sourceOVNPath                              = "/var/log/ovn"
	sourceOAuthServerName                      = "varlogoauthserver"
	sourceOAuthServerPath                      = "/var/log/oauth-server"
	sourceOAuthAPIServerName                   = "varlogoauthapiserver"
	sourceOAuthAPIServerPath                   = "/var/log/oauth-apiserver"
	sourceOpenshiftAPIServerName               = "varlogopenshiftapiserver"
	sourceOpenshiftAPIServerPath               = "/var/log/openshift-apiserver"
	sourceKubeAPIServerName                    = "varlogkubeapiserver"
	sourceKubeAPIServerPath                    = "/var/log/kube-apiserver"
	tmpVolumeName                              = "tmp"
	tmpPath                                    = "/tmp"

	// collectorRunAsUser must be 0 (root) because container log files on the host are 0600 root:root.
	// DAC_READ_SEARCH cannot be used as a non-root workaround because Linux only grants effective
	// capabilities to non-root processes when the binary has file capabilities (setcap), and the
	// Vector binary does not. The security improvement comes from replacing spc_t with the
	// MCS-constrained container_logwriter_t SELinux domain and dropping all dangerous capabilities.
	collectorRunAsUser int64 = 0
	// collectorRunAsGroup is the primary GID.
	collectorRunAsGroup int64 = 0
	// selinuxTypeLogWriter (container_logwriter_t) is an MCS-constrained container domain that
	// grants read on /var/log/** plus inotify watch permissions on container_log_t, and write
	// on /var/lib/vector/**. Far more restrictive than spc_t (super-privileged container).
	selinuxTypeLogWriter = "container_logwriter_t"
	// initContainerName is the name of the init container that prepares the data directory.
	initContainerName = "data-dir-init"
)

type Visitor func(collector *v1.Container, podSpec *v1.PodSpec, resNames *factory.ForwarderResourceNames, namespace, logLevel string)
type CommonLabelVisitor func(o runtime.Object)
type PodLabelVisitor func(o runtime.Object)

type Factory struct {
	ConfigHash             string
	CollectorSpec          obs.CollectorSpec
	ClusterID              string
	ImageName              string
	Visit                  Visitor
	Secrets                internalobs.Secrets
	ConfigMaps             internalobs.ConfigMaps
	ForwarderSpec          obs.ClusterLogForwarderSpec
	CommonLabelInitializer CommonLabelVisitor
	PodLabelVisitor        PodLabelVisitor
	ResourceNames          *factory.ForwarderResourceNames
	isDaemonset            bool
	annotations            map[string]string
}

// CollectorResourceRequirements returns the resource requirements for a given collector implementation
// or it's default if none are specified
func (f *Factory) CollectorResourceRequirements() v1.ResourceRequirements {
	if f.CollectorSpec.Resources == nil {
		return v1.ResourceRequirements{}
	}
	return *f.CollectorSpec.Resources
}

func (f *Factory) NodeSelector() map[string]string {
	return f.CollectorSpec.NodeSelector
}
func (f *Factory) Tolerations() []v1.Toleration {
	return f.CollectorSpec.Tolerations
}
func (f *Factory) Affinity() *v1.Affinity {
	return f.CollectorSpec.Affinity
}
func (f *Factory) MaxUnavailable() intstr.IntOrString {
	if f.CollectorSpec.MaxUnavailable != nil {
		return *f.CollectorSpec.MaxUnavailable
	}
	if f.annotations != nil {
		if value, found := f.annotations[constants.AnnotationMaxUnavailable]; found {
			return intstr.Parse(value)
		}
	}
	return intstr.Parse(DefaultMaxUnavailable)
}

func New(confHash, clusterID string, collectorSpec *obs.CollectorSpec, secrets internalobs.Secrets, configMaps internalobs.ConfigMaps, forwarderSpec obs.ClusterLogForwarderSpec, resNames *factory.ForwarderResourceNames, isDaemonset bool, annotations map[string]string) *Factory {
	if collectorSpec == nil {
		collectorSpec = &obs.CollectorSpec{}
	}
	factory := &Factory{
		ClusterID:     clusterID,
		ConfigHash:    confHash,
		CollectorSpec: *collectorSpec,
		ImageName:     constants.VectorName,
		Visit:         vector.CollectorVisitor,
		ConfigMaps:    configMaps,
		Secrets:       secrets,
		ForwarderSpec: forwarderSpec,
		CommonLabelInitializer: func(o runtime.Object) {
			runtime.SetCommonLabels(o, constants.VectorName, resNames.ForwarderName, constants.CollectorName)
		},
		ResourceNames:   resNames,
		PodLabelVisitor: vector.PodLogExcludeLabel,
		isDaemonset:     isDaemonset,
		annotations:     annotations,
	}
	return factory
}

func (f *Factory) NewDaemonSet(namespace, name string, trustedCABundle *v1.ConfigMap, tlsProfileSpec configv1.TLSProfileSpec) *apps.DaemonSet {
	podSpec := f.NewPodSpec(trustedCABundle, f.ForwarderSpec, f.ClusterID, tlsProfileSpec, namespace)
	ds := factory.NewDaemonSet(namespace, name, name, constants.CollectorName, constants.VectorName, f.MaxUnavailable(), *podSpec, f.CommonLabelInitializer, f.PodLabelVisitor)
	ds.Spec.Template.Annotations[constants.AnnotationSecretHash] = f.Secrets.Hash64a()
	ds.Spec.Template.Annotations[constants.AnnotationConfigMapHash] = f.ConfigMaps.Hash64a()
	return ds
}

func (f *Factory) NewDeployment(namespace, name string, trustedCABundle *v1.ConfigMap, tlsProfileSpec configv1.TLSProfileSpec) *apps.Deployment {
	podSpec := f.NewPodSpec(trustedCABundle, f.ForwarderSpec, f.ClusterID, tlsProfileSpec, namespace)
	dpl := factory.NewDeployment(namespace, name, constants.CollectorName, constants.VectorName, 2, *podSpec, f.CommonLabelInitializer, f.PodLabelVisitor)
	dpl.Spec.Template.Annotations[constants.AnnotationSecretHash] = f.Secrets.Hash64a()
	dpl.Spec.Template.Annotations[constants.AnnotationConfigMapHash] = f.ConfigMaps.Hash64a()
	return dpl
}

func (f *Factory) NewPodSpec(trustedCABundle *v1.ConfigMap, spec obs.ClusterLogForwarderSpec, clusterID string, tlsProfileSpec configv1.TLSProfileSpec, namespace string) *v1.PodSpec {

	var gracePeriod *int64
	if f.CollectorSpec.TerminationGracePeriodSeconds != nil {
		gracePeriod = f.CollectorSpec.TerminationGracePeriodSeconds
	} else {
		gracePeriod = utils.GetPtr(defaultTerminationGracePeriodSeconds)
	}

	podSpec := &v1.PodSpec{
		NodeSelector:                  utils.EnsureLinuxNodeSelector(f.NodeSelector()),
		PriorityClassName:             clusterLoggingPriorityClassName,
		ServiceAccountName:            f.ResourceNames.ServiceAccount,
		TerminationGracePeriodSeconds: gracePeriod,
		Tolerations:                   append(constants.DefaultTolerations(), f.Tolerations()...),
		Affinity:                      f.Affinity(),

		Volumes: []v1.Volume{
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: f.ResourceNames.SecretMetrics}}},
			{Name: tmpVolumeName, VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory}}},
		},
	}

	if f.isDaemonset {
		inputs := internalobs.Inputs(spec.Inputs)
		if inputs.HasContainerSource() {
			podSpec.Volumes = append(podSpec.Volumes,
				v1.Volume{Name: sourcePodsName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourcePodsPath}}},
			)
		}
		if inputs.HasJournalSource() {
			podSpec.Volumes = append(podSpec.Volumes,
				v1.Volume{Name: sourceJournalName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourceJournalPath}}},
			)
		}
		if inputs.HasAuditSource(obs.AuditSourceAuditd) {
			podSpec.Volumes = append(podSpec.Volumes,
				v1.Volume{Name: sourceAuditdName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourceAuditdPath}}},
			)
		}
		if inputs.HasAuditSource(obs.AuditSourceKube) {
			podSpec.Volumes = append(podSpec.Volumes,
				v1.Volume{Name: sourceKubeAPIServerName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourceKubeAPIServerPath}}},
			)
		}
		if inputs.HasAuditSource(obs.AuditSourceOpenShift) {
			podSpec.Volumes = append(podSpec.Volumes,
				v1.Volume{Name: sourceOpenshiftAPIServerName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourceOpenshiftAPIServerPath}}},
				v1.Volume{Name: sourceOAuthServerName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourceOAuthServerPath}}},
				v1.Volume{Name: sourceOAuthAPIServerName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourceOAuthAPIServerPath}}},
			)
		}
		if inputs.HasAuditSource(obs.AuditSourceOVN) {
			podSpec.Volumes = append(podSpec.Volumes,
				v1.Volume{Name: sourceAuditOVNName, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: sourceOVNPath}}},
			)
		}
	}

	secretVolumes := AddSecretVolumes(podSpec, f.Secrets)
	configmapVolumes := AddConfigmapVolumes(podSpec, f.ConfigMaps)
	if internalobs.Outputs(spec.Outputs).NeedServiceAccountToken() {
		AddServiceAccountProjectedVolume(podSpec, defaultAudience)
	}

	collector := f.NewCollectorContainer(spec.Inputs, spec.Outputs, secretVolumes, configmapVolumes, clusterID)

	addTrustedCABundle(collector, podSpec, trustedCABundle)

	f.Visit(collector, podSpec, f.ResourceNames, namespace, LogLevel(f.annotations))

	// Add init container for daemonsets to prepare the data directory with proper SELinux labeling
	if f.isDaemonset {
		dataPath := vector.GetDataPath(namespace, f.ResourceNames.ForwarderName)
		podSpec.InitContainers = []v1.Container{
			newDataDirInitContainer(dataPath),
		}
	}

	podSpec.Containers = []v1.Container{
		*collector,
	}
	return podSpec
}

// NewCollectorContainer is a constructor for creating the collector container spec.  Note the secretNames are assumed
// to be a unique list
func (f *Factory) NewCollectorContainer(inputs internalobs.Inputs, outputs internalobs.Outputs, secretVolumes, configmapVolumes []string, clusterID string) *v1.Container {
	collector := runtime.NewContainer(constants.CollectorName, utils.GetComponentImage(f.ImageName), v1.PullIfNotPresent, f.CollectorSpec.Resources)
	collector.TerminationMessagePolicy = v1.TerminationMessageFallbackToLogsOnError
	collector.Ports = []v1.ContainerPort{
		{
			Name:          constants.MetricsPortName,
			ContainerPort: constants.MetricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	collector.Env = []v1.EnvVar{
		{Name: "COLLECTOR_CONF_HASH", Value: f.ConfigHash},
		{Name: "K8S_NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
		{Name: "OPENSHIFT_CLUSTER_ID", Value: clusterID},
		{Name: "POD_IP", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
		{Name: "POD_IPS", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIPs"}}},
		{Name: "VECTOR_RAISE_FD_LIMIT", Value: "true"},
	}
	collector.Env = append(collector.Env, utils.GetProxyEnvVars()...)

	collector.VolumeMounts = []v1.VolumeMount{
		{Name: metricsVolumeName, ReadOnly: true, MountPath: metricsVolumePath},
		{Name: tmpVolumeName, MountPath: tmpPath},
	}

	if f.isDaemonset {
		if inputs.HasContainerSource() {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourcePodsName, ReadOnly: true, MountPath: sourcePodsPath})
		}
		if inputs.HasJournalSource() {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourceJournalName, ReadOnly: true, MountPath: sourceJournalPath})
		}
		if inputs.HasAuditSource(obs.AuditSourceAuditd) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourceAuditdName, ReadOnly: true, MountPath: sourceAuditdPath})
		}
		if inputs.HasAuditSource(obs.AuditSourceKube) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourceKubeAPIServerName, ReadOnly: true, MountPath: sourceKubeAPIServerPath})
		}
		if inputs.HasAuditSource(obs.AuditSourceOpenShift) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourceOpenshiftAPIServerName, ReadOnly: true, MountPath: sourceOpenshiftAPIServerPath})
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourceOAuthServerName, ReadOnly: true, MountPath: sourceOAuthServerPath})
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourceOAuthAPIServerName, ReadOnly: true, MountPath: sourceOAuthAPIServerPath})
		}
		if inputs.HasAuditSource(obs.AuditSourceOVN) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: sourceAuditOVNName, ReadOnly: true, MountPath: sourceOVNPath})
		}
		AddSecurityContextTo(collector)
	}

	AddVolumeMounts(collector, secretVolumes, common.SecretBasePath)
	AddVolumeMounts(collector, configmapVolumes, func(name string) string {
		return common.ConfigMapBasePath(strings.TrimPrefix(name, "config-"))
	})

	if outputs.NeedServiceAccountToken() {
		AddVolumeMounts(collector, []string{saTokenVolumeName}, func(name string) string {
			// projected sa tokens are created in their own 'token' directory at this path
			return constants.ServiceAccountSecretPath
		})
	}

	return collector
}

func sanitizeVolumeName(input string) string {
	return strings.ReplaceAll(input, ".", "")
}

// AddVolumeMounts to the collector container
func AddVolumeMounts(collector *v1.Container, names []string, path func(string) string) {
	log.WithName("AddVolumeMounts").V(4).Info("volumeMounts", "names", names)
	for _, name := range names {
		log.WithName("volumeMount").V(4).Info("mount", "name", name)
		collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{
			Name:      sanitizeVolumeName(name),
			ReadOnly:  true,
			MountPath: path(name),
		})
	}
}

// AddSecretVolumes adds secret volumes to the pod spec for the unique set of output secrets and returns the list of
// the names
func AddSecretVolumes(podSpec *v1.PodSpec, secrets internalobs.Secrets) []string {
	names := secrets.Names()
	log.WithName("AddSecretVolumes").V(4).Info("volumes", "names", secrets.Names())
	for _, name := range names {
		log.WithName("AddSecretVolumes").V(4).Info("secret", "name", name)
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
			Name: sanitizeVolumeName(name),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: name,
				},
			},
		})
	}
	return names
}

// AddConfigmapVolumes adds configmap volumes to the pod spec for the unique set of configmaps and returns the list of
// the named volumes where the names are of the format 'config-<ConfigMap.Name>'
func AddConfigmapVolumes(podSpec *v1.PodSpec, configMaps internalobs.ConfigMaps) (results []string) {
	names := configMaps.Names()
	log.WithName("AddConfigmapVolumes").V(4).Info("volumes", "names", names)
	for _, name := range names {
		vName := fmt.Sprintf("config-%s", name)
		log.WithName("AddConfigmapVolumes").V(4).Info("configmap", "name", vName)
		results = append(results, vName)
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
			Name: sanitizeVolumeName(vName),
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
				},
			},
		})
	}
	return results
}

// AddServiceAccountProjectedVolume adds ServiceAccountTokenProjection to the podspec and returns the named sa volume
func AddServiceAccountProjectedVolume(podSpec *v1.PodSpec, audience string) {
	podSpec.Volumes = append(podSpec.Volumes,
		v1.Volume{
			Name: saTokenVolumeName,
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					Sources: []v1.VolumeProjection{
						{
							ServiceAccountToken: &v1.ServiceAccountTokenProjection{
								Audience:          audience,
								ExpirationSeconds: utils.GetPtr[int64](saTokenExpirationSecs),
								Path:              constants.TokenKey,
							},
						},
					},
				},
			},
		})
}

func AddSecurityContextTo(container *v1.Container) *v1.Container {
	container.SecurityContext = &v1.SecurityContext{
		Capabilities: &v1.Capabilities{
			Drop: auth.RequiredDropCapabilities,
		},
		SELinuxOptions: &v1.SELinuxOptions{
			Type: selinuxTypeLogWriter,
		},
		RunAsUser:                utils.GetPtr(collectorRunAsUser),
		RunAsGroup:               utils.GetPtr(collectorRunAsGroup),
		RunAsNonRoot:             utils.GetPtr(false),
		ReadOnlyRootFilesystem:   utils.GetPtr(true),
		AllowPrivilegeEscalation: utils.GetPtr(false),
		SeccompProfile: &v1.SeccompProfile{
			Type: v1.SeccompProfileTypeRuntimeDefault,
		},
	}
	return container
}

// newDataDirInitContainer creates an init container that prepares the collector's data directory.
// The init container runs as root with spc_t to create the directory and relabel it to
// container_file_t (which container_logwriter_t can write to). Ownership is set to 1000:0
// to prepare for a future non-root collector migration.
func newDataDirInitContainer(dataPath string) v1.Container {
	return v1.Container{
		Name:    initContainerName,
		Image:   utils.GetComponentImage(constants.VectorName),
		Command: []string{"/bin/sh", "-c"},
		Args: []string{
			"set -e; " +
				"mkdir -p \"$1\" && " +
				"chown -R 1000:0 \"$1\" && " +
				"chmod -R 2770 \"$1\" && " +
				"chcon -R -t container_file_t \"$1\"",
			"--",
			dataPath,
		},
		VolumeMounts: []v1.VolumeMount{
			{Name: common.DataDir, MountPath: dataPath},
		},
		SecurityContext: &v1.SecurityContext{
			Privileged:               utils.GetPtr(true), // Init container needs privilege for chown and chcon
			AllowPrivilegeEscalation: utils.GetPtr(true),
			SELinuxOptions: &v1.SELinuxOptions{
				Type: "spc_t", // Init container needs spc_t for chcon
			},
			RunAsUser: utils.GetPtr[int64](0), // Must run as root for chcon and chown
		},
	}
}

func addTrustedCABundle(collector *v1.Container, podSpec *v1.PodSpec, trustedCABundleCM *v1.ConfigMap) {
	if trustedCABundleCM != nil {
		if bundle, found := hasTrustedCABundle(trustedCABundleCM); found {
			collector.VolumeMounts = append(collector.VolumeMounts,
				v1.VolumeMount{
					Name:      constants.VolumeNameTrustedCA,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				})

			podSpec.Volumes = append(podSpec.Volumes,
				v1.Volume{
					Name: constants.VolumeNameTrustedCA,
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: trustedCABundleCM.Name,
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
			if bundleHash, err := utils.CalculateMD5Hash(bundle); err == nil {
				collector.Env = append(collector.Env, v1.EnvVar{
					Name:  common.TrustedCABundleHashName,
					Value: bundleHash,
				})
			} else {
				log.V(0).Error(err, "There was an error trying to calculate the hash of the trusted CA", "bundle")
			}
		}
	}
}

func hasTrustedCABundle(configMap *v1.ConfigMap) (string, bool) {
	if configMap == nil {
		return "", false
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	return caBundle, ok && caBundle != ""
}

func LogLevel(annotations map[string]string) string {
	if level, ok := annotations[constants.AnnotationVectorLogLevel]; ok {
		return level
	}
	return "warn"
}
