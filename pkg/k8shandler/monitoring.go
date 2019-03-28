package k8shandler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const fluentdAlertsFile = "/etc/fluentd/prometheus/fluentd_prometheus_alerts.yaml"


func BuildPrometheusRule(ruleName string, namespace string) (*monitoringv1.PrometheusRule, error) {
	alertsRuleFile, ok := os.LookupEnv("ALERTS_FILE_PATH")
	if !ok {
		alertsRuleFile = fluentdAlertsFile
	}

	alertsRuleSpec, err := newPrometheusRuleSpecFrom(alertsRuleFile)
	if err != nil {
		return nil, err
	}

	rule := utils.NewPrometheusRule(ruleName, namespace)
	rule.Spec = *alertsRuleSpec

	return rule, nil
}

func newPrometheusRuleSpecFrom(filePath string) (*monitoringv1.PrometheusRuleSpec, error) {
	if err := verifyFileExists(filePath); err != nil {
		return nil, err
	}
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("'%s' not readable", filePath)
	}
	ruleSpec := monitoringv1.PrometheusRuleSpec{}
	if err := k8sYAML.NewYAMLOrJSONDecoder(bytes.NewBufferString(string(fileContent)), 1000).Decode(&ruleSpec); err != nil {
		return nil, err
	}
	return &ruleSpec, nil
}

func verifyFileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("'%s' not found", filePath)
		}
		return err
	}
	return nil
}

func createOrUpdateServiceMonitor(cluster *ClusterLogging) error {
	serviceMonitor := utils.NewServiceMonitor("fluentd", cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   metricsPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", "fluentd", cluster.Namespace),
			// ServerName can be e.g. fluentd.openshift-logging.svc
		},
	}

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"logging-infra": "support",
		},
	}

	serviceMonitor.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-fluentd",
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{cluster.Namespace},
		},
	}

	utils.AddOwnerRefToObject(serviceMonitor, utils.AsOwner(cluster))

	err := sdk.Create(serviceMonitor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd ServiceMonitor: %v", err)
	}

	return nil
}

func createOrUpdatePrometheusRule(cluster *ClusterLogging) error {
	promRule, err := BuildPrometheusRule("fluentd", cluster.Namespace)
	if err != nil {
		return fmt.Errorf("Failure creating the fluentd PrometheusRule: %v", err)
	}

	utils.AddOwnerRefToObject(promRule, utils.AsOwner(cluster))

	err = sdk.Create(promRule)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating the fluentd PrometheusRule: %v", err)
	}

	return nil
}
