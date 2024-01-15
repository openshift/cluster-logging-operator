package lokistack

import (
	"context"
	"github.com/ViaQ/logerr/v2/kverrors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	lokiStackFinalizer = "logging.openshift.io/lokistack-rbac"
)

// CheckFinalizer checks if the finalizer used for tracking the cluster-wide RBAC resources
// is attached to the provided object and removes it, if present.
func CheckFinalizer(ctx context.Context, client client.Client, obj client.Object) error {
	if controllerutil.RemoveFinalizer(obj, lokiStackFinalizer) {
		if err := client.Update(ctx, obj); err != nil {
			return kverrors.Wrap(err, "Failed to remove finalizer from ClusterLogging.")
		}
	}

	return nil
}
