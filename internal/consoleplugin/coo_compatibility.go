package consoleplugin

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	uiPluginAPIVersion = "observability.openshift.io/v1alpha1"
	uiPluginKind       = "UIPlugin"
)

func (r *Reconciler) checkObservabilityOperator(ctx context.Context) (bool, error) {
	cooManaged, err := r.isManagedByObservabilityOperator(ctx)
	if err != nil {
		return false, err
	}

	if !cooManaged {
		// ConsolePlugin exists but is not managed by Cluster Observability Operator
		return false, nil
	}

	// COO is installed in parallel and is managing the logging-view-plugin.
	// The cluster-scoped resources are already "adopted" by COO at this point, so we only need to clean up
	// our namespaced resources in openshift-logging: Deployment, Service and ConfigMap
	log.V(3).Info("Found ConsolePlugin managed by Cluster Observability Operator")

	err = r.each(func(m mutable) error {
		if m.o == &r.consolePlugin {
			// Do not operate on the ConsolePlugin
			return nil
		}

		err := r.c.Delete(ctx, m.o)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		return nil
	})

	if err != nil {
		return false, fmt.Errorf("error removing console resources due to COO presence: %w", err)
	}

	return true, nil
}

func (r *Reconciler) isManagedByObservabilityOperator(ctx context.Context) (bool, error) {
	key := client.ObjectKey{
		Name: Name,
	}

	plugin := &consolev1alpha1.ConsolePlugin{}
	err := r.c.Get(ctx, key, plugin)
	switch {
	case apierrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("error getting ConsolePlugin: %w", err)
	default:
		return hasUIPluginOwner(plugin), nil
	}
}

func hasUIPluginOwner(plugin *consolev1alpha1.ConsolePlugin) bool {
	ownerRef := metav1.GetControllerOf(plugin)
	if ownerRef == nil {
		return false
	}

	return ownerRef.APIVersion == uiPluginAPIVersion && ownerRef.Kind == uiPluginKind
}
