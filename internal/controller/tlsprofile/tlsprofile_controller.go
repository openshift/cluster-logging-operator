package tlsprofile

import (
	"context"
	"os"
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TLSProfileReconciler watches for APIServer TLS profile changes and restarts the operator pod
type TLSProfileReconciler struct {
	client.Client
	InitialProfile *configv1.TLSSecurityProfile
}

// Reconcile handles APIServer TLS profile changes
func (r *TLSProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Name != tls.APIServerName {
		return ctrl.Result{}, nil
	}

	apiServer := &configv1.APIServer{}
	if err := r.Get(ctx, req.NamespacedName, apiServer); err != nil {
		log.Error(err, "Failed to get APIServer")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if TLS profile changed from initial startup
	if !reflect.DeepEqual(r.InitialProfile, apiServer.Spec.TLSSecurityProfile) {
		log.Info("Cluster TLS profile has changed, operator will restart to apply new configuration",
			"oldProfile", r.InitialProfile,
			"newProfile", apiServer.Spec.TLSSecurityProfile)

		// Exit gracefully to allow Kubernetes to restart the pod
		// Exit code 0 indicates normal termination
		os.Exit(0)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *TLSProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1.APIServer{}).
		WithEventFilter(tls.APIServerTLSProfileChangedPredicate(false)).
		Complete(r)
}
