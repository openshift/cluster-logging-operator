package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// OpenTelemetryCollector creates or updates an OpenTelemetryCollector owned by
// the desired ownerReferences. If a collector with the same name already
// exists and is not owned by the desired owners, it is left unchanged and an
// error is returned.
func OpenTelemetryCollector(k8Client client.Client, desired *otelv1beta1.OpenTelemetryCollector) error {
	otelcol := &otelv1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      desired.Name,
			Namespace: desired.Namespace,
		},
	}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, otelcol, func() error {
		if err := utils.EnsureCanUpdateOwnedResource(otelcol, desired.OwnerReferences...); err != nil {
			return err
		}
		otelcol.Labels = desired.Labels
		otelcol.OwnerReferences = desired.OwnerReferences
		otelcol.Spec = desired.Spec
		return nil
	})
	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled OpenTelemetryCollector - operation: %s", op))
	}
	return err
}
