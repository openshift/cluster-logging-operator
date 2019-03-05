package k8shandler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

const fluentdAlertsFile = "/usr/bin/files/fluentd_prometheus_alerts.yaml"

func BuildPrometheusRule(ruleName string, namespace string) (*monitoringv1.PrometheusRule, error) {
	alertsRuleFile, ok := os.LookupEnv("ALERTS_FILE_PATH")
	if !ok {
		alertsRuleFile = fluentdAlertsFile
	}

	alertsRuleSpec, err := ruleSpec(alertsRuleFile)
	if err != nil {
		return nil, err
	}

	rule := utils.PrometheusRule(ruleName, namespace)
	rule.Spec = *alertsRuleSpec

	return rule, nil
}

func ruleSpec(filePath string) (*monitoringv1.PrometheusRuleSpec, error) {
	if err := checkFile(filePath); err != nil {
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

func checkFile(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("'%s' not found", filePath)
		}
		return err
	}
	return nil
}
