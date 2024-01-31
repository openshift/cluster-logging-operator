package collector

import (
	"path"

	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	clusterLoggingPriorityClassName = "system-node-critical"
	MetricsPort                     = int32(24231)
	MetricsPortName                 = "metrics"
	logContainers                   = "varlogcontainers"
	logContainersValue              = "/var/log/containers"
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
	httpInputVolumePath             = "/etc/collector/"
	tmpVolumeName                   = "tmp"
	tmpPath                         = "/tmp"
)

type Visitor func(collector *v1.Container, podSpec *v1.PodSpec, resNames *factory.ForwarderResourceNames, namespace string)
type CommonLabelVisitor func(o runtime.Object)
type PodLabelVisitor func(o runtime.Object)

type Factory struct {
	ConfigHash             string
	CollectorSpec          logging.CollectionSpec
	CollectorType          logging.LogCollectionType
	ClusterID              string
	ImageName              string
	TrustedCAHash          string
	Visit                  Visitor
	Secrets                map[string]*v1.Secret
	ForwarderSpec          logging.ClusterLogForwarderSpec
	CommonLabelInitializer CommonLabelVisitor
	PodLabelVisitor        PodLabelVisitor
	ResourceNames          *factory.ForwarderResourceNames
}

// CollectorResourceRequirements returns the resource requirements for a given collector implementation
// or it's default if none are specified
func (f *Factory) CollectorResourceRequirements() v1.ResourceRequirements {
	if f.CollectorSpec.Resources == nil {
		if f.CollectorType == logging.LogCollectionTypeVector {
			return v1.ResourceRequirements{}
		}
		return v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: fluentd.DefaultMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: fluentd.DefaultMemory,
				v1.ResourceCPU:    fluentd.DefaultCpuRequest,
			},
		}
	}
	return *f.CollectorSpec.Resources
}

func (f *Factory) NodeSelector() map[string]string {
	return f.CollectorSpec.CollectorSpec.NodeSelector
}
func (f *Factory) Tolerations() []v1.Toleration {
	return f.CollectorSpec.CollectorSpec.Tolerations
}

func New(confHash, clusterID string, collectorSpec logging.CollectionSpec, secrets map[string]*v1.Secret, forwarderSpec logging.ClusterLogForwarderSpec, instanceName string, resNames *factory.ForwarderResourceNames) *Factory {
	factory := &Factory{
		ClusterID:     clusterID,
		ConfigHash:    confHash,
		CollectorSpec: collectorSpec,
		CollectorType: collectorSpec.Type,
		ImageName:     constants.FluentdName,
		Visit:         fluentd.CollectorVisitor,
		Secrets:       secrets,
		ForwarderSpec: forwarderSpec,
		CommonLabelInitializer: func(o runtime.Object) {
			runtime.SetCommonLabels(o, utils.GetCollectorName(collectorSpec.Type), instanceName, constants.CollectorName)
		},
		ResourceNames:   resNames,
		PodLabelVisitor: func(o runtime.Object) {}, //do noting for fluentd
	}
	if collectorSpec.Type == logging.LogCollectionTypeVector {
		factory.ImageName = constants.VectorName
		factory.Visit = vector.CollectorVisitor
		factory.PodLabelVisitor = vector.PodLogExcludeLabel
	}
	return factory
}

func (f *Factory) NewDaemonSet(namespace, name string, trustedCABundle *v1.ConfigMap, tlsProfileSpec configv1.TLSProfileSpec, httpInputs []string) *apps.DaemonSet {
	podSpec := f.NewPodSpec(trustedCABundle, f.ForwarderSpec, f.ClusterID, f.TrustedCAHash, tlsProfileSpec, httpInputs, namespace)
	ds := factory.NewDaemonSet(name, namespace, f.ResourceNames.CommonName, constants.CollectorName, string(f.CollectorSpec.Type), *podSpec, f.CommonLabelInitializer, f.PodLabelVisitor)
	return ds
}

func (f *Factory) NewPodSpec(trustedCABundle *v1.ConfigMap, forwarderSpec logging.ClusterLogForwarderSpec, clusterID, trustedCAHash string, tlsProfileSpec configv1.TLSProfileSpec, httpInputs []string, namespace string) *v1.PodSpec {

	podSpec := &v1.PodSpec{
		NodeSelector:                  utils.EnsureLinuxNodeSelector(f.NodeSelector()),
		PriorityClassName:             clusterLoggingPriorityClassName,
		ServiceAccountName:            f.ResourceNames.ServiceAccount,
		TerminationGracePeriodSeconds: utils.GetPtr[int64](10),
		Tolerations:                   constants.DefaultTolerations(),
		Volumes: []v1.Volume{
			{Name: logContainers, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logContainersValue}}},
			{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}},
			{Name: logJournal, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logJournalValue}}},
			{Name: logAudit, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logAuditValue}}},
			{Name: logOvn, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOvnValue}}},
			{Name: logOauthapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOauthapiserverValue}}},
			{Name: logOauthserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOauthserverValue}}},
			{Name: logOpenshiftapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOpenshiftapiserverValue}}},
			{Name: logKubeapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logKubeapiserverValue}}},
			{Name: f.ResourceNames.SecretMetrics, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: f.ResourceNames.SecretMetrics}}},
			{Name: tmpVolumeName, VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory}}},
		},
	}
	for _, httpInput := range httpInputs {
		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{Name: httpInput, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: httpInput}}},
		)
	}

	podSpec.Tolerations = append(podSpec.Tolerations, f.Tolerations()...)

	secretNames := AddSecretVolumes(podSpec, forwarderSpec)

	collector := f.NewCollectorContainer(secretNames, clusterID, httpInputs)

	addTrustedCABundle(collector, podSpec, trustedCABundle, f.ResourceNames.CaTrustBundle)

	f.Visit(collector, podSpec, f.ResourceNames, namespace)

	addWebIdentityForCloudwatch(collector, podSpec, forwarderSpec, f.Secrets, f.CollectorType)

	podSpec.Containers = []v1.Container{
		*collector,
	}

	return podSpec
}

// NewCollectorContainer is a constructor for creating the collector container spec.  Note the secretNames are assumed
// to be a unique list
func (f *Factory) NewCollectorContainer(secretNames []string, clusterID string, httpInputs []string) *v1.Container {

	collector := factory.NewContainer(constants.CollectorName, f.ImageName, v1.PullIfNotPresent, f.CollectorResourceRequirements())
	collector.Ports = []v1.ContainerPort{
		{
			Name:          MetricsPortName,
			ContainerPort: MetricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	collector.Env = []v1.EnvVar{
		{Name: "COLLECTOR_CONF_HASH", Value: f.ConfigHash},
		{Name: common.TrustedCABundleHashName, Value: f.TrustedCAHash},
		{Name: "K8S_NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
		{Name: "OPENSHIFT_CLUSTER_ID", Value: clusterID},
		{Name: "POD_IP", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
		{Name: "POD_IPS", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIPs"}}},
	}
	collector.Env = append(collector.Env, utils.GetProxyEnvVars()...)

	collector.VolumeMounts = []v1.VolumeMount{
		{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
		{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
		{Name: logJournal, ReadOnly: true, MountPath: logJournalValue},
		{Name: logAudit, ReadOnly: true, MountPath: logAuditValue},
		{Name: logOvn, ReadOnly: true, MountPath: logOvnValue},
		{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
		{Name: logOauthserver, ReadOnly: true, MountPath: logOauthserverValue},
		{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
		{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
		{Name: f.ResourceNames.SecretMetrics, ReadOnly: true, MountPath: metricsVolumePath},
		{Name: tmpVolumeName, MountPath: tmpPath},
	}
	for _, httpInput := range httpInputs {
		collector.VolumeMounts = append(collector.VolumeMounts,
			v1.VolumeMount{Name: httpInput, ReadOnly: true, MountPath: httpInputVolumePath + httpInput},
		)
	}

	// List of _unique_ output secret names, several outputs may use the same secret.
	AddSecretVolumeMounts(&collector, secretNames)

	AddSecurityContextTo(&collector)
	return &collector
}

func (f *Factory) ReconcileInputServices(er record.EventRecorder, k8sClient client.Client, namespace, selectorComponent string, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {
	if f.CollectorType != logging.LogCollectionTypeVector {
		return nil
	}

	for _, input := range f.ForwarderSpec.Inputs {
		if input.Receiver != nil && input.Receiver.HTTP != nil {
			listenPort := input.Receiver.HTTP.GetPort()
			if err := network.ReconcileInputService(er, k8sClient, namespace, input.Name, selectorComponent, input.Name, listenPort, listenPort, owner, visitors); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddSecretVolumeMounts to the collector container
func AddSecretVolumeMounts(collector *v1.Container, secretNames []string) {
	// List of _unique_ output secret names, several outputs may use the same secret.
	for _, name := range secretNames {
		path := OutputSecretPath(name)
		collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: name, ReadOnly: true, MountPath: path})
	}
}

func OutputSecretPath(secretName string) string {
	return path.Join(constants.CollectorSecretsDir, secretName)
}

// AddSecretVolumes adds secret volumes to the pod spec for the unique set of pipeline secrets and returns the list of
// the secret names
func AddSecretVolumes(podSpec *v1.PodSpec, pipelineSpec logging.ClusterLogForwarderSpec) []string {
	// List of _unique_ output secret names, several outputs may use the same secret.
	unique := sets.NewString()
	for _, o := range pipelineSpec.Outputs {
		if o.Secret != nil && o.Secret.Name != "" {
			unique.Insert(o.Secret.Name)
		}
	}
	secretNames := unique.List()
	for _, name := range secretNames {
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{Name: name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: name}}})
	}
	return secretNames
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

func addTrustedCABundle(collector *v1.Container, podSpec *v1.PodSpec, trustedCABundleCM *v1.ConfigMap, name string) {
	if trustedCABundleCM != nil && hasTrustedCABundle(trustedCABundleCM) {
		collector.VolumeMounts = append(collector.VolumeMounts,
			v1.VolumeMount{
				Name:      name,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})

		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{
				Name: name,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: name,
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
		//TODO add trusted ca hash to env vars
	}
}

func hasTrustedCABundle(configMap *v1.ConfigMap) bool {
	if configMap == nil {
		return false
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	return ok && caBundle != ""
}
