package collector

import (
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/metrics/alerts"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcilePrometheusRule(r record.EventRecorder, k8sClient client.Client, collectorType logging.LogCollectionType, namespace, name string, owner metav1.OwnerReference) error {
	if namespace != constants.OpenshiftNS {
		log.V(3).Info("prometheusrules will only be reconciled in the openshift-logging namespace")
		return nil
	}

	rule := runtime.NewPrometheusRule(namespace, constants.CollectorName)

	spec, err := runtime.NewPrometheusRuleSpecFrom(alerts.CollectorPrometheusAlert)
	if err != nil {
		return fmt.Errorf("failure creating the collector PrometheusRule: %w", err)
	}

	rule.Spec = *spec

	utils.AddOwnerRefToObject(rule, owner)
	return reconcile.PrometheusRule(r, k8sClient, rule)
}
