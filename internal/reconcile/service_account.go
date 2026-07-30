package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ServiceAccount(k8Client client.Client, desired *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	sa := runtime.NewServiceAccount(desired.Namespace, desired.Name)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, sa, func() error {
		if err := utils.EnsureCanUpdateOwnedResource(sa, desired.OwnerReferences); err != nil {
			return err
		}
		sa.Annotations = desired.Annotations
		sa.OwnerReferences = desired.OwnerReferences
		sa.Finalizers = desired.Finalizers
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled serviceAccount - operation: %s", op))
	}
	return sa, err
}
