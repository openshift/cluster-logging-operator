package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ServiceMonitor reconciles a ServiceMonitor to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func ServiceMonitor(k8Client client.Client, desired *monitoringv1.ServiceMonitor) error {
	sm := runtime.NewServiceMonitor(desired.Namespace, desired.Name)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, sm, func() error {
		if err := utils.EnsureCanUpdateOwnedResource(sm, desired.OwnerReferences); err != nil {
			return err
		}
		sm.Labels = desired.Labels
		sm.Spec = desired.Spec
		sm.Annotations = desired.Annotations
		sm.OwnerReferences = desired.OwnerReferences
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled serviceMonitor - operation: %s", op))
	}
	return err
}
