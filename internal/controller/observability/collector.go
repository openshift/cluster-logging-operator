package observability

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	generatorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

func ReconcileCollector(context internalcontext.ForwarderContext, pollInterval, timeout time.Duration) (err error) {
	// TODO LOG-2620: containers violate PodSecurity ? should we still do this or should this be
	// a pre-req to multi CLF?
	//// LOG-2620: containers violate PodSecurity
	//if err = clusterRequest.addSecurityLabelsToNamespace(); err != nil {
	//	log.Error(err, "Error adding labels to logging Namespace")
	//	return
	//}

	if err = reconcile.SecurityContextConstraints(context.Client, auth.NewSCC()); err != nil {
		log.V(3).Error(err, "reconcile.SecurityContextConstraints")
		return err
	}

	ownerRef := utils.AsOwner(context.Forwarder)
	resourceNames := factory.ResourceNames(*context.Forwarder)

	// Add roles to ServiceAccount to allow the collector to read from the node
	if err = auth.ReconcileRBAC(noOpEventRecorder, context.Client, context.Forwarder.Name, context.Forwarder.Namespace, context.Forwarder.Spec.ServiceAccount.Name, ownerRef); err != nil {
		log.V(3).Error(err, "auth.ReconcileRBAC")
		return
	}

	// TODO: This can be the same per NS but what is the ownerref?  Multiple CLFs will clash
	if err = collector.ReconcileTrustedCABundleConfigMap(noOpEventRecorder, context.Client, context.Forwarder.Namespace, resourceNames.CaTrustBundle, ownerRef); err != nil {
		log.Error(err, "collector.ReconcileTrustedCABundleConfigMap")
		return err
	}
	trustedCABundle := collector.WaitForTrustedCAToBePopulated(context.Client, context.Forwarder.Namespace, resourceNames.CaTrustBundle, pollInterval, timeout)

	var collectorConfig string
	if collectorConfig, err = GenerateConfig(context.Client, *context.Forwarder, *resourceNames, context.Secrets); err != nil {
		log.V(9).Error(err, "collector.GenerateConfig")
		return err
	}
	log.V(3).Info("Generated collector config", "config", collectorConfig)
	var collectorConfHash string
	collectorConfHash, err = utils.CalculateMD5Hash(collectorConfig)
	if err != nil {
		log.Error(err, "unable to calculate MD5 hash")
		log.V(9).Error(err, "Returning from unable to calculate MD5 hash")
		return
	}

	secretReaderScript := common.GenerateSecretReaderScript(context.Secrets)
	isDaemonSet := !internalobs.DeployAsDeployment(*context.Forwarder)
	log.V(3).Info("Deploying as DaemonSet", "isDaemonSet", isDaemonSet)
	factory := collector.New(collectorConfHash, context.ClusterID, context.Forwarder.Spec.Collector, context.Secrets, context.ConfigMaps, context.Forwarder.Spec, resourceNames, isDaemonSet, LogLevel(context.Forwarder.Annotations))
	if err = factory.ReconcileCollectorConfig(noOpEventRecorder, context.Client, context.Reader, context.Forwarder.Namespace, collectorConfig, secretReaderScript, ownerRef); err != nil {
		log.Error(err, "collector.ReconcileCollectorConfig")
		return
	}

	reconcileWorkload := factory.ReconcileDaemonset
	if !isDaemonSet {
		reconcileWorkload = factory.ReconcileDeployment
	}
	if err := reconcileWorkload(noOpEventRecorder, context.Client, context.Forwarder.Namespace, trustedCABundle, ownerRef); err != nil {
		log.Error(err, "Error reconciling the deployment of the collector")
		return err
	}

	if err := factory.ReconcileInputServices(noOpEventRecorder, context.Client, context.Reader, context.Forwarder.Namespace, resourceNames.CommonName, ownerRef, factory.CommonLabelInitializer); err != nil {
		log.Error(err, "collector.ReconcileInputServices")
		return err
	}

	// Reconcile resources to support metrics gathering
	if err := network.ReconcileService(noOpEventRecorder, context.Client, context.Forwarder.Namespace, resourceNames.CommonName, constants.CollectorName, collector.MetricsPortName, resourceNames.SecretMetrics, collector.MetricsPort, ownerRef, factory.CommonLabelInitializer); err != nil {
		log.Error(err, "collector.ReconcileService")
		return err
	}
	if err := metrics.ReconcileServiceMonitor(noOpEventRecorder, context.Client, context.Forwarder.Namespace, resourceNames.CommonName, constants.CollectorName, collector.MetricsPortName, ownerRef); err != nil {
		log.Error(err, "collector.ReconcileServiceMonitor")
		return err
	}

	return nil
}

func GenerateConfig(k8Client client.Client, spec obs.ClusterLogForwarder, resourceNames factory.ForwarderResourceNames, secrets helpers.Secrets) (config string, err error) {
	op := framework.Options{}
	tlsProfile, _ := tls.FetchAPIServerTlsProfile(k8Client)
	op[framework.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(tlsProfile)
	//EvaluateAnnotationsForEnabledCapabilities(clusterRequest.Forwarder, op)
	g := forwardergenerator.New()
	generatedConfig, err := g.GenerateConf(secrets, spec.Spec, spec.Namespace, spec.Name, resourceNames, op)

	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		return "", err
	}

	log.V(3).Info("ClusterLogForwarder generated config", generatedConfig)
	return generatedConfig, err
}

// EvaluateAnnotationsForEnabledCapabilities populates generator options with capabilities enabled by the ClusterLogForwarder
func EvaluateAnnotationsForEnabledCapabilities(annotations map[string]string, options framework.Options) {
	if annotations == nil {
		return
	}
	for key, value := range annotations {
		switch key {
		case constants.AnnotationDebugOutput:
			if strings.ToLower(value) == "true" {
				options[generatorhelpers.EnableDebugOutput] = "true"
			}
		}
	}
}

func LogLevel(annotations map[string]string) string {
	if level, ok := annotations[constants.AnnotationVectorLogLevel]; ok {
		return level
	}
	return "warn"
}
