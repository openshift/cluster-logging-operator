package collector

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/metrics/alerts"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcilePrometheusRule(r record.EventRecorder, k8sClient client.Client, collectorType logging.LogCollectionType, namespace, name string, owner metav1.OwnerReference) error {

	rule := runtime.NewPrometheusRule(namespace, name)
	alertRulesFile := alerts.FluentdPrometheusAlert
	if collectorType == logging.LogCollectionTypeVector {
		alertRulesFile = alerts.VectorPrometheusAlerts
	}

	spec, err := runtime.NewPrometheusRuleSpecFrom(alertRulesFile)
	if err != nil {
		return fmt.Errorf("failure creating the collector PrometheusRule: %w", err)
	}

	rule.Spec = *spec

	utils.AddOwnerRefToObject(rule, owner)
	return reconcile.PrometheusRule(r, k8sClient, rule)
}
