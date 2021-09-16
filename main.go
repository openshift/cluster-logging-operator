package main

import (
	"flag"
	"fmt"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"os"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/apis"
	"github.com/openshift/cluster-logging-operator/version"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	oauth "github.com/openshift/api/oauth/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/controllers/clusterlogging"
	"github.com/openshift/cluster-logging-operator/controllers/forwarding"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis"
)

// Change below variables to serve metrics on different host or port.
var (
	scheme   = apiruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(apis.AddToScheme(scheme))
	utilruntime.Must(elasticsearch.AddToScheme(scheme))
	utilruntime.Must(routev1.AddToScheme(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))
	utilruntime.Must(oauth.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(securityv1.AddToScheme(scheme))

	utilruntime.Must(loggingv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()
	logLevel, present := os.LookupEnv("LOG_LEVEL")
	if present {
		verbosity, err := strconv.Atoi(logLevel)
		if err != nil {
			log.Error(err, "LOG_LEVEL must be an integer", "value", logLevel)
			os.Exit(1)
		}
		log.MustInit("cluster-logging-operator")
		log.SetLogLevel(verbosity)
	} else {
		log.MustInit("cluster-logging-operator")
	}
	log.Info("starting up...",
		"operator_version", version.Version,
		"go_version", runtime.Version(),
		"go_os", runtime.GOOS,
		"go_arch", runtime.GOARCH,
	)

	namespace, err := getWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "b430cc2e.openshift.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	if err = (&clusterlogging.ReconcileClusterLogging{
		Client: mgr.GetClient(),
		//Log:    ctrl.Log.WithName("controllers").WithName("ClusterLogForwarder"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("clusterlogging-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterLogForwarder")
		os.Exit(1)
	}
	if err = (&forwarding.ReconcileForwarder{
		Client: mgr.GetClient(),
		//Log:    ctrl.Log.WithName("controllers").WithName("ClusterLogging"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("clusterlogforwarder"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterLogging")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

// getWatchNamespace get the namespace name of the scoped operator
// - https://sdk.operatorframework.io/docs/building-operators/golang/operator-scope/#configuring-namespace-scoped-operators
func getWatchNamespace() (string, error) {
	watchNamespaceEnvVar := "WATCH_NAMESPACE"
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
