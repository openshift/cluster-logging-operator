package stub

import (
  "context"
  "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
  "github.com/openshift/cluster-logging-operator/pkg/k8shandler"
  "github.com/operator-framework/operator-sdk/pkg/sdk"
  "github.com/sirupsen/logrus"
)

func NewHandler() sdk.Handler {
  return &Handler{}
}

type Handler struct {
  // Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
  switch o := event.Object.(type) {
    case *v1alpha1.ClusterLogging:
      return Reconcile(o)
  }
  return nil
}

func Reconcile(logging *v1alpha1.ClusterLogging)(err error) {
  logrus.Info("Started reconciliation")

  // Reconcile certs
  err = k8shandler.CreateOrUpdateCertificates(logging)
  if err != nil {
    logrus.Fatalf("Unable to create or update certificates: %v", err)
  }

  // Reconcile Log Store
  err = k8shandler.CreateOrUpdateLogStore(logging)
  if err != nil {
    logrus.Fatalf("Unable to create or update logstore: %v", err)
  }

  // Reconcile Visualization
  err = k8shandler.CreateOrUpdateVisualization(logging)
  if err != nil {
    logrus.Fatalf("Unable to create or update visualization: %v", err)
  }

  // Reconcile Curation

  // Reconcile Collection

  return nil
}
