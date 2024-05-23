package observability

import (
	"context"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	generatorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func ReconcileCollector(k8Client client.Client, k8Reader client.Reader, clf obs.ClusterLogForwarder, clusterID string) (err error) {
	// TODO LOG-2620: containers violate PodSecurity ? should we still do this or should this be
	// a pre-req to multi CLF?
	//// LOG-2620: containers violate PodSecurity
	//if err = clusterRequest.addSecurityLabelsToNamespace(); err != nil {
	//	log.Error(err, "Error adding labels to logging Namespace")
	//	return
	//}

	if err = reconcile.SecurityContextConstraints(k8Client, auth.NewSCC()); err != nil {
		log.V(3).Error(err, "reconcile.SecurityContextConstraints")
		return err
	}

	ownerRef := utils.AsOwner(&clf)
	resourceNames := factory.ResourceNames(clf)

	// Add roles to ServiceAccount to allow the collector to read from the node
	if err = auth.ReconcileRBAC(noOpEventRecorder, k8Client, clf.Namespace, resourceNames, ownerRef); err != nil {
		log.V(3).Error(err, "auth.ReconcileRBAC")
		return
	}

	// TODO: This can be the same per NS but what is the ownerref?  Multiple CLFs will clash
	if err = collector.ReconcileTrustedCABundleConfigMap(noOpEventRecorder, k8Client, clf.Namespace, resourceNames.CaTrustBundle, ownerRef); err != nil {
		log.Error(err, "collector.ReconcileTrustedCABundleConfigMap")
		return err
	}
	trustedCABundle := collector.WaitForTrustedCAToBePopulated(k8Client, clf.Namespace, resourceNames.CaTrustBundle)

	secrets, err := LoadSecrets(k8Client, clf.Namespace, clf.Spec.Inputs, clf.Spec.Outputs)
	if err != nil {
		log.V(3).Error(err, "auth.LoadSecrets")
		return err
	}

	var collectorConfig string
	if collectorConfig, err = GenerateConfig(k8Client, clf, *resourceNames, secrets); err != nil {
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
	isDaemonSet := ShouldDeployAsDaemonSet(clf.Annotations, clf.Spec.Inputs)
	log.V(3).Info("Deploying as DaemonSet", "isDaemonSet", isDaemonSet)
	factory := collector.New(collectorConfHash, clusterID, clf.Spec.Collector, secrets, clf.Spec, resourceNames, isDaemonSet, LogLevel(clf.Annotations))
	if err = factory.ReconcileCollectorConfig(noOpEventRecorder, k8Client, k8Reader, clf.Namespace, collectorConfig, ownerRef); err != nil {
		log.Error(err, "collector.ReconcileCollectorConfig")
		return
	}

	reconcileDeployment := factory.ReconcileDaemonset
	if !isDaemonSet {
		reconcileDeployment = factory.ReconcileDeployment
	}
	if err := reconcileDeployment(noOpEventRecorder, k8Client, clf.Namespace, trustedCABundle, ownerRef); err != nil {
		log.Error(err, "Error reconciling the deployment of the collector")
		return err
	}

	if err := factory.ReconcileInputServices(noOpEventRecorder, k8Client, clf.Namespace, resourceNames.CommonName, ownerRef, factory.CommonLabelInitializer); err != nil {
		log.Error(err, "collector.ReconcileInputServices")
		return err
	}

	// Reconcile resources to support metrics gathering
	if err := network.ReconcileService(noOpEventRecorder, k8Client, clf.Namespace, resourceNames.CommonName, constants.CollectorName, collector.MetricsPortName, resourceNames.SecretMetrics, collector.MetricsPort, ownerRef, factory.CommonLabelInitializer); err != nil {
		log.Error(err, "collector.ReconcileService")
		return err
	}
	if err := metrics.ReconcileServiceMonitor(noOpEventRecorder, k8Client, clf.Namespace, resourceNames.CommonName, constants.CollectorName, collector.MetricsPortName, ownerRef); err != nil {
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

func LoadSecrets(k8Client client.Client, namespace string, inputs internalobs.Inputs, outputs internalobs.Outputs) (secrets helpers.Secrets, err error) {
	secrets = helpers.Secrets{}
	for _, name := range inputs.SecretNames() {
		secret := &corev1.Secret{}
		if err = k8Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, secret); err == nil {
			secrets[name] = secret
		} else {
			return secrets, err
		}
	}
	return secrets, nil
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

// ShouldDeployAsDaemonSet evaluates the forwarder spec to determine if it should deply as a deployment or daemonset
// Collector is can be deployed as a deployment if the only input source is an HTTP receiver and it has the necessary
// annotation
func ShouldDeployAsDaemonSet(annoations map[string]string, inputs internalobs.Inputs) bool {
	if _, ok := annoations[constants.AnnotationEnableCollectorAsDeployment]; ok {
		asDeployment := inputs.HasReceiverSource() && !inputs.HasContainerSource() && !inputs.HasJournalSource() && !inputs.HasAnyAuditSource()
		return !asDeployment
	}
	return true
}
