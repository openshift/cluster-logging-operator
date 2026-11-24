package observability

import (
	"context"
	"fmt"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	otelruntime "github.com/openshift/cluster-logging-operator/internal/runtime/opentelemetry"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/yaml"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileOpenTelemetryCollector creates or updates the OpenTelemetryCollector
// CR carrying the generated collector configuration, letting the
// OpenTelemetry operator deploy and manage the collector workload.
// The CR runs as a daemonset, or a deployment when the forwarder spec requires
// it. It uses the forwarder's service account so the reconciled RBAC applies,
// mounts the forwarder's configmaps and secrets at the paths the generated
// config references, and receives the node name and cluster ID through the
// environment variables the config expands.
func ReconcileOpenTelemetryCollector(fctx internalcontext.ForwarderContext, resourceNames *factory.ForwarderResourceNames, config string, ownerRef metav1.OwnerReference) error {
	mode := otelv1beta1.ModeDaemonSet
	if internalobs.DeployAsDeployment(*fctx.Forwarder) {
		mode = otelv1beta1.ModeDeployment
	}

	// the generated config is a YAML string; the CR carries it as the
	// structured collector configuration
	var collectorConfig otelv1beta1.Config
	if err := yaml.Unmarshal([]byte(config), &collectorConfig); err != nil {
		log.Error(err, "unable to unmarshal the generated otelc config")
		return err
	}

	// mount every configmap the forwarder references through the operator's
	// spec.configmaps, at the OtelConfigMapPath locations the generated
	// config references (the shared path logic in internal/collector/common)
	configMaps := make([]otelv1beta1.ConfigMapsSpec, 0, len(fctx.ConfigMaps))
	for name := range fctx.ConfigMaps {
		configMaps = append(configMaps, otelv1beta1.ConfigMapsSpec{
			Name:      name,
			MountPath: common.OtelConfigMapMountPath,
		})
	}
	// secrets have no operator-native mount: mount them with raw volumes at
	// the well-known collector secrets directory the generated config expands
	volumes := make([]corev1.Volume, 0, len(fctx.Secrets))
	volumeMounts := make([]corev1.VolumeMount, 0, len(fctx.Secrets))
	for name := range fctx.Secrets {
		volumes = append(volumes, corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: name,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name,
			MountPath: common.SecretBasePath(name),
		})
	}

	// mount the node filesystem directories holding the log sources the
	// forwarder's inputs read, mirroring the volumes the CLO-managed vector
	// daemonset mounts. The forwarder's service account is already granted the
	// CLO collector SCC, which permits hostPath volumes
	sourceVolumes, sourceMounts := logSourceVolumes(fctx.Forwarder.Spec.Inputs)
	volumes = append(volumes, sourceVolumes...)
	volumeMounts = append(volumeMounts, sourceMounts...)

	desired := otelruntime.NewOpenTelemetryCollector(fctx.Forwarder.Namespace, resourceNames.CommonName, runtime.Initialize, func(otelcol *otelv1beta1.OpenTelemetryCollector) {
		runtime.SetCommonLabels(otelcol, constants.ComponentNameOtelc, resourceNames.ForwarderName, constants.CollectorName)
		utils.AddOwnerRefToObject(otelcol, ownerRef)
		otelcol.Spec = otelv1beta1.OpenTelemetryCollectorSpec{
			Mode:       mode,
			Config:     collectorConfig,
			ConfigMaps: configMaps,
			OpenTelemetryCommonFields: otelv1beta1.OpenTelemetryCommonFields{
				ServiceAccount: fctx.Forwarder.Spec.ServiceAccount.Name,
				Volumes:        volumes,
				VolumeMounts:   volumeMounts,
				// mirror the vector collector's security context: the node log
				// directories (e.g. /var/log/pods) are root-owned and mode
				// 0750, so the collector runs as root with the spc_t SELinux
				// type; the otel operator's default non-root pod context must
				// be overridden for the runAsUser to be admitted
				PodSecurityContext: &corev1.PodSecurityContext{
					RunAsNonRoot: utils.GetPtr(false),
				},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: utils.GetPtr[int64](0),
					Capabilities: &corev1.Capabilities{
						Drop: auth.RequiredDropCapabilities,
					},
					SELinuxOptions: &corev1.SELinuxOptions{
						Type: "spc_t",
					},
					ReadOnlyRootFilesystem:   utils.GetPtr(true),
					AllowPrivilegeEscalation: utils.GetPtr(false),
					SeccompProfile: &corev1.SeccompProfile{
						Type: corev1.SeccompProfileTypeRuntimeDefault,
					},
				},
				Env: []corev1.EnvVar{
					{
						Name: "NODE_NAME",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
						},
					},
					{
						Name:  "OPENSHIFT_CLUSTER_ID",
						Value: fctx.ClusterID,
					},
				},
			},
		}
	})

	return reconcile.OpenTelemetryCollector(fctx.Client, desired)
}

// logSourceVolumes returns the hostPath volumes and read-only mounts for the
// node filesystem directories holding the log sources the inputs read. It
// mirrors the source volumes the CLO-managed vector daemonset mounts
func logSourceVolumes(inputs []obs.InputSpec) ([]corev1.Volume, []corev1.VolumeMount) {
	inputTypes := internalobs.Inputs(inputs)
	dirs := []string{}
	if inputTypes.HasContainerSource() {
		dirs = append(dirs, "/var/log/pods")
	}
	if inputTypes.HasJournalSource() {
		dirs = append(dirs, "/var/log/journal")
	}
	if inputTypes.HasAuditSource(obs.AuditSourceAuditd) {
		dirs = append(dirs, "/var/log/audit")
	}
	if inputTypes.HasAuditSource(obs.AuditSourceKube) {
		dirs = append(dirs, "/var/log/kube-apiserver")
	}
	if inputTypes.HasAuditSource(obs.AuditSourceOpenShift) {
		dirs = append(dirs, "/var/log/openshift-apiserver", "/var/log/oauth-server", "/var/log/oauth-apiserver")
	}
	if inputTypes.HasAuditSource(obs.AuditSourceOVN) {
		dirs = append(dirs, "/var/log/ovn")
	}

	volumes := make([]corev1.Volume, 0, len(dirs))
	mounts := make([]corev1.VolumeMount, 0, len(dirs))
	for _, dir := range dirs {
		// volume names must be DNS-1123; use the path with slashes stripped
		name := strings.ReplaceAll(strings.Trim(dir, "/"), "/", "-")
		volumes = append(volumes, corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: dir},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      name,
			MountPath: dir,
			ReadOnly:  true,
		})
	}
	return volumes, mounts
}

// otelcCRDAvailable returns a non-nil error when the OpenTelemetryCollector
// custom resource definition is not served by the cluster, which means the
// OpenTelemetry operator (a deployment-order prerequisite of the otelc
// collector implementation) is not installed
func otelcCRDAvailable(k8sClient client.Client) error {
	list := &otelv1beta1.OpenTelemetryCollectorList{}
	if err := k8sClient.List(context.Background(), list, client.Limit(1)); err != nil {
		if meta.IsNoMatchError(err) || apiruntime.IsNotRegisteredError(err) {
			return fmt.Errorf("the opentelemetrycollectors.opentelemetry.io custom resource definition is not served: %w", err)
		}
		return err
	}
	return nil
}
