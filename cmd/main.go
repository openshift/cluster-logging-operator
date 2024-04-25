package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/metrics/dashboard"
	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	observabilitycontroller "github.com/openshift/cluster-logging-operator/internal/controller/observability"

	log "github.com/ViaQ/logerr/v2/log/static"

	apis "github.com/openshift/cluster-logging-operator/api"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/version"

	apiruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	oauth "github.com/openshift/api/oauth/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/openshift/cluster-logging-operator/internal/controller/clusterlogging"
	"github.com/openshift/cluster-logging-operator/internal/controller/forwarding"
	"github.com/openshift/cluster-logging-operator/internal/controller/logfilemetricsexporter"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
)

// Change below variables to serve metrics on different host or port.
var (
	scheme = apiruntime.NewScheme()
)

const (
	UnHealthyStatus = "0"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(apis.AddToScheme(scheme))
	utilruntime.Must(elasticsearch.AddToScheme(scheme))
	utilruntime.Must(routev1.AddToScheme(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))
	utilruntime.Must(consolev1alpha1.AddToScheme(scheme))
	utilruntime.Must(oauth.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(securityv1.AddToScheme(scheme))

	utilruntime.Must(loggingv1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(observabilityv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8686", "The address the metric endpoint binds to.")
	//flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe end point binds to.")

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	logger := utils.InitLogger("cluster-logging-operator")
	// LOG-5136 Fixes error caused by updates to controller-runtime
	ctrl.SetLogger(logger)

	log.Info("starting up...",
		"operator_version", version.Version,
		"go_version", runtime.Version(),
		"go_os", runtime.GOOS,
		"go_arch", runtime.GOARCH,
	)

	// https://issues.redhat.com/browse/LOG-3321
	cacheOptions := cache.Options{
		SyncPeriod: utils.GetPtr(time.Minute * 3),
	}
	if watchNS := getWatchNS(); len(watchNS) > 0 {
		cacheOptions.DefaultNamespaces = map[string]cache.Config{}
		for _, ns := range watchNS {
			cacheOptions.DefaultNamespaces[ns] = cache.Config{}
		}
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "b430cc2e.openshift.io",
		Cache:                  cacheOptions,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Clean up
	defer func() {
		if err := cleanUpResources(mgr.GetClient()); err != nil {
			log.V(3).Error(err, "error with resource cleanup")
		}
	}()

	clusterVersion, clusterID, err := version.ClusterVersion(mgr.GetAPIReader())
	if err != nil {
		log.Error(err, "unable to retrieve the cluster version")
		os.Exit(1)
	}
	migrateManifestResources(mgr.GetClient())

	log.Info("Registering Components.")

	if err = (&clusterlogging.ReconcileClusterLogging{
		Client:         mgr.GetClient(),
		Reader:         mgr.GetAPIReader(),
		Scheme:         mgr.GetScheme(),
		Recorder:       mgr.GetEventRecorderFor("clusterlogging-controller"),
		ClusterVersion: clusterVersion,
		ClusterID:      clusterID,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "ClusterLogForwarder")
		telemetry.Data.CLInfo.Set("healthStatus", UnHealthyStatus)
		os.Exit(1)
	}
	if err = (&forwarding.ReconcileForwarder{
		Client:         mgr.GetClient(),
		Reader:         mgr.GetAPIReader(),
		Scheme:         mgr.GetScheme(),
		Recorder:       mgr.GetEventRecorderFor("clusterlogforwarder"),
		ClusterVersion: clusterVersion,
		ClusterID:      clusterID,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "ClusterLogging")
		telemetry.Data.CLFInfo.Set("healthStatus", UnHealthyStatus)
		os.Exit(1)
	}

	// The Log File Metric Exporter Controller
	if err = (&logfilemetricsexporter.ReconcileLogFileMetricExporter{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		Recorder:       mgr.GetEventRecorderFor("logfilemetricexporter"),
		ClusterVersion: clusterVersion,
		ClusterID:      clusterID,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "LogFileMetricExporter")
		telemetry.Data.LFMEInfo.Set(telemetry.HealthStatus, UnHealthyStatus)
		os.Exit(1)
	}

	if err = (&dashboard.ReconcileDashboards{
		Client: mgr.GetClient(),
		Reader: mgr.GetAPIReader(),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "GrafanaDashboard")
		os.Exit(1)
	}

	if err = (&observabilitycontroller.ClusterLogForwarderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "observability.ClusterLogForwarder")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// updating clo Telemetry Data - to be published by prometheus
	cloversion, err := version.OperatorVersion()
	if err != nil {
		cloversion = version.Version
		log.Info("Failed to get clo version from env variable OPERATOR_CONDITION_NAME so falling back to default version")
	}
	telemetry.Data.CLInfo.Set("version", cloversion)

	errr := telemetry.RegisterMetrics()
	if errr != nil {
		log.Error(err, "Error in registering clo metrics for telemetry")
	}

	log.Info("Starting the Cmd.")
	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}

}

// getWatchNS returns the namespaces being watched by the operator.  Empty means all
// - https://sdk.operatorframework.io/docs/building-operators/golang/operator-scope/#configuring-namespace-scoped-operators
func getWatchNS() []string {
	OpenshiftNSEnvVar := "WATCH_NAMESPACE"
	ns, found := os.LookupEnv(OpenshiftNSEnvVar)
	if !found {
		log.Error(fmt.Errorf("Exiting. %s must be set", OpenshiftNSEnvVar), "Failed to get watch namespace")
		os.Exit(1)
	}
	return strings.Split(ns, ",")
}

func migrateManifestResources(k8sClient client.Client) {
	log.Info("migrating resources provided by the manifest")
	if err := k8sClient.Delete(context.TODO(), loggingruntime.NewPriorityClass("cluster-logging", 0, false, "")); err != nil && !errors.IsNotFound(err) {
		log.V(1).Error(err, "There was an error trying to remove the old collector PriorityClass named 'cluster-logging'")
	}
}

func cleanUpResources(k8sClient client.Client) error {
	// Remove the dashboard config map
	if err := dashboard.RemoveDashboardConfigMap(k8sClient); err != nil {
		return err
	}
	return nil
}
