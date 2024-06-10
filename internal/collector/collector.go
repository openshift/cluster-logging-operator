package collector

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/utils/set"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	defaultAudience                 = "openshift"
	saTokenVolumeName               = "sa-token"
	saTokenExpirationSecs           = 3600 //1 hour
	clusterLoggingPriorityClassName = "system-node-critical"
	MetricsPort                     = int32(24231)
	MetricsPortName                 = "metrics"
	logPods                         = "varlogpods"
	logPodsValue                    = "/var/log/pods"
	logJournal                      = "varlogjournal"
	logJournalValue                 = "/var/log/journal"
	logAudit                        = "varlogaudit"
	logAuditValue                   = "/var/log/audit"
	logOvn                          = "varlogovn"
	logOvnValue                     = "/var/log/ovn"
	logOauthapiserver               = "varlogoauthapiserver"
	logOauthapiserverValue          = "/var/log/oauth-apiserver"
	logOauthserver                  = "varlogoauthserver"
	logOauthserverValue             = "/var/log/oauth-server"
	logOpenshiftapiserver           = "varlogopenshiftapiserver"
	logOpenshiftapiserverValue      = "/var/log/openshift-apiserver"
	logKubeapiserver                = "varlogkubeapiserver"
	logKubeapiserverValue           = "/var/log/kube-apiserver"
	metricsVolumePath               = "/etc/collector/metrics"
	tmpVolumeName                   = "tmp"
	tmpPath                         = "/tmp"
)

var (
	saTokenPath = common.ServiceAccountBasePath(saTokenVolumeName)
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
	Secrets                map[string]*v1.Secret
	ForwarderSpec          obs.ClusterLogForwarderSpec
	CommonLabelInitializer CommonLabelVisitor
	PodLabelVisitor        PodLabelVisitor
	ResourceNames          *factory.ForwarderResourceNames
	isDaemonset            bool
	LogLevel               string
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

func New(confHash, clusterID string, collectorSpec *obs.CollectorSpec, secrets map[string]*v1.Secret, forwarderSpec obs.ClusterLogForwarderSpec, resNames *factory.ForwarderResourceNames, isDaemonset bool, logLevel string) *Factory {
	if collectorSpec == nil {
		collectorSpec = &obs.CollectorSpec{}
	}
	factory := &Factory{
		ClusterID:     clusterID,
		ConfigHash:    confHash,
		CollectorSpec: *collectorSpec,
		ImageName:     constants.VectorName,
		Visit:         vector.CollectorVisitor,
		Secrets:       secrets,
		ForwarderSpec: forwarderSpec,
		CommonLabelInitializer: func(o runtime.Object) {
			runtime.SetCommonLabels(o, constants.VectorName, resNames.ForwarderName, constants.CollectorName)
		},
		ResourceNames:   resNames,
		PodLabelVisitor: vector.PodLogExcludeLabel,
		isDaemonset:     isDaemonset,
		LogLevel:        logLevel,
	}
	return factory
}

func (f *Factory) NewDaemonSet(namespace, name string, trustedCABundle *v1.ConfigMap, tlsProfileSpec configv1.TLSProfileSpec, receiverInputs []string) *apps.DaemonSet {
	podSpec := f.NewPodSpec(trustedCABundle, f.ForwarderSpec, f.ClusterID, tlsProfileSpec, receiverInputs, namespace)
	ds := factory.NewDaemonSet(name, namespace, f.ResourceNames.CommonName, constants.CollectorName, constants.VectorName, *podSpec, f.CommonLabelInitializer, f.PodLabelVisitor)
	return ds
}

func (f *Factory) NewDeployment(namespace, name string, trustedCABundle *v1.ConfigMap, tlsProfileSpec configv1.TLSProfileSpec, receiverInputs []string) *apps.Deployment {
	podSpec := f.NewPodSpec(trustedCABundle, f.ForwarderSpec, f.ClusterID, tlsProfileSpec, receiverInputs, namespace)
	dpl := factory.NewDeployment(namespace, name, f.ResourceNames.CommonName, constants.CollectorName, constants.VectorName, *podSpec, f.CommonLabelInitializer, f.PodLabelVisitor)
	return dpl
}

func (f *Factory) NewPodSpec(trustedCABundle *v1.ConfigMap, spec obs.ClusterLogForwarderSpec, clusterID string, tlsProfileSpec configv1.TLSProfileSpec, receiverInputs []string, namespace string) *v1.PodSpec {

	podSpec := &v1.PodSpec{
		NodeSelector:                  utils.EnsureLinuxNodeSelector(f.NodeSelector()),
		PriorityClassName:             clusterLoggingPriorityClassName,
		ServiceAccountName:            f.ResourceNames.ServiceAccount,
		TerminationGracePeriodSeconds: utils.GetPtr[int64](10),
		Tolerations:                   append(constants.DefaultTolerations(), f.Tolerations()...),
		Volumes: []v1.Volume{
			{Name: f.ResourceNames.SecretMetrics, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: f.ResourceNames.SecretMetrics}}},
			{Name: tmpVolumeName, VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory}}},
		},
	}

	if f.isDaemonset {
		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}},
			v1.Volume{Name: logJournal, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logJournalValue}}},
			v1.Volume{Name: logAudit, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logAuditValue}}},
			v1.Volume{Name: logOvn, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOvnValue}}},
			v1.Volume{Name: logOauthapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOauthapiserverValue}}},
			v1.Volume{Name: logOauthserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOauthserverValue}}},
			v1.Volume{Name: logOpenshiftapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOpenshiftapiserverValue}}},
			v1.Volume{Name: logKubeapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logKubeapiserverValue}}},
		)
	}

	for _, receiverInput := range receiverInputs {
		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{Name: receiverInput, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: receiverInput}}},
		)
	}

	secretVolumes := AddSecretVolumes(podSpec, spec.Inputs, spec.Outputs)
	configmapVolumes := AddConfigmapVolumes(podSpec, spec.Inputs, spec.Outputs)
	addServiceAccountVolume := AddServiceAccountProjectedVolume(podSpec, spec.Inputs, spec.Outputs, defaultAudience)

	collector := f.NewCollectorContainer(spec.Inputs, secretVolumes, configmapVolumes, addServiceAccountVolume, clusterID)

	addTrustedCABundle(collector, podSpec, trustedCABundle)

	f.Visit(collector, podSpec, f.ResourceNames, namespace, f.LogLevel)

	addWebIdentityForCloudwatch(collector, podSpec, spec, f.Secrets)

	podSpec.Containers = []v1.Container{
		*collector,
	}
	return podSpec
}

// NewCollectorContainer is a constructor for creating the collector container spec.  Note the secretNames are assumed
// to be a unique list
func (f *Factory) NewCollectorContainer(inputs internalobs.Inputs, secretVolumes, configmapVolumes []string, addServiceAccountVolume bool, clusterID string) *v1.Container {

	collector := runtime.NewContainer(constants.CollectorName, utils.GetComponentImage(f.ImageName), v1.PullIfNotPresent, f.CollectorSpec.Resources)
	collector.Ports = []v1.ContainerPort{
		{
			Name:          MetricsPortName,
			ContainerPort: MetricsPort,
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
	}
	collector.Env = append(collector.Env, utils.GetProxyEnvVars()...)

	collector.VolumeMounts = []v1.VolumeMount{
		{Name: f.ResourceNames.SecretMetrics, ReadOnly: true, MountPath: metricsVolumePath},
		{Name: tmpVolumeName, MountPath: tmpPath},
	}

	if f.isDaemonset {
		if inputs.HasContainerSource() {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: logPods, ReadOnly: true, MountPath: logPodsValue})
		}
		if inputs.HasJournalSource() {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: logJournal, ReadOnly: true, MountPath: logJournalValue})
		}
		if inputs.HasAuditSource(obs.AuditSourceAuditd) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: logAudit, ReadOnly: true, MountPath: logAuditValue})
		}
		if inputs.HasAuditSource(obs.AuditSourceKube) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: logAudit, ReadOnly: true, MountPath: logKubeapiserverValue})
		}
		if inputs.HasAuditSource(obs.AuditSourceOpenShift) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: logAudit, ReadOnly: true, MountPath: logOpenshiftapiserverValue})
		}
		if inputs.HasAuditSource(obs.AuditSourceOVN) {
			collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: logAudit, ReadOnly: true, MountPath: logOvnValue})
		}
		AddSecurityContextTo(collector)
	}

	AddVolumeMounts(collector, secretVolumes, common.SecretBasePath)
	AddVolumeMounts(collector, configmapVolumes, func(name string) string {
		return common.ConfigMapBasePath(strings.TrimPrefix(name, "config-"))
	})
	if addServiceAccountVolume {
		AddVolumeMounts(collector, []string{saTokenVolumeName}, common.ServiceAccountBasePath)
	}

	return collector
}

func (f *Factory) ReconcileInputServices(er record.EventRecorder, k8sClient client.Client, namespace, selectorComponent string, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {
	for _, input := range f.ForwarderSpec.Inputs {
		var listenPort int32
		serviceName := f.ResourceNames.GenerateInputServiceName(input.Name)
		if input.Receiver != nil {
			listenPort = input.Receiver.Port
			if err := network.ReconcileInputService(er, k8sClient, namespace, serviceName, selectorComponent, serviceName, listenPort, listenPort, input.Receiver.Type, f.isDaemonset, owner, visitors); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddVolumeMounts to the collector container
func AddVolumeMounts(collector *v1.Container, names []string, path func(string) string) {
	for _, name := range names {
		collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: name, ReadOnly: true, MountPath: path(name)})
	}
}

// AddSecretVolumes adds secret volumes to the pod spec for the unique set of output secrets and returns the list of
// the names
func AddSecretVolumes(podSpec *v1.PodSpec, inputs internalobs.Inputs, outputs internalobs.Outputs) []string {
	// List of _unique_ output secret names, several outputs may use the same secret.
	secretNames := set.New(outputs.SecretNames()...).Insert(inputs.SecretNames()...).UnsortedList()
	for _, name := range secretNames {
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{Name: name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: name}}})
	}
	return secretNames
}

// AddConfigmapVolumes adds configmap volumes to the pod spec for the unique set of configmaps and returns the list of
// the named volumes where the names are of the format 'config-<ConfigMap.Name>'
func AddConfigmapVolumes(podSpec *v1.PodSpec, inputs internalobs.Inputs, outputs internalobs.Outputs) (results []string) {
	// List of _unique_ output secret names, several outputs may use the same secret.
	names := set.New(outputs.ConfigmapNames()...).Insert(inputs.ConfigmapNames()...).UnsortedList()
	for _, name := range names {
		vName := fmt.Sprintf("config-%s", name)
		results = append(results, vName)
		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{
				Name: vName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						}}}})
	}
	return results
}

// AddServiceAccountProjectedVolume adds ServiceAccountTokenProjection to the podspec and returns the named sa volume
func AddServiceAccountProjectedVolume(podSpec *v1.PodSpec, inputs internalobs.Inputs, outputs internalobs.Outputs, audience string) bool {
	if outputs.NeedServiceAccountToken() {
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
		return true
	}
	return false
}

func AddSecurityContextTo(container *v1.Container) *v1.Container {
	container.SecurityContext = &v1.SecurityContext{
		Capabilities: &v1.Capabilities{
			Drop: auth.RequiredDropCapabilities,
		},
		SELinuxOptions: &v1.SELinuxOptions{
			Type: "spc_t",
		},
		ReadOnlyRootFilesystem:   utils.GetPtr(true),
		AllowPrivilegeEscalation: utils.GetPtr(false),
		SeccompProfile: &v1.SeccompProfile{
			Type: v1.SeccompProfileTypeRuntimeDefault,
		},
	}
	return container
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
