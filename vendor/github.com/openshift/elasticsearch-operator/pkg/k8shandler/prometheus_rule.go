package k8shandler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	alertsFilePath = "/etc/elasticsearch-operator/files/prometheus_alerts.yml"
	rulesFilePath  = "/etc/elasticsearch-operator/files/prometheus_rules.yml"
)

func CreateOrUpdatePrometheusRules(dpl *v1alpha1.Elasticsearch) error {
	ruleName := fmt.Sprintf("%s-%s", dpl.Name, "prometheus-rules")
	owner := asOwner(dpl)

	promRule, err := buildPrometheusRule(ruleName, dpl.Namespace, dpl.Labels)
	if err != nil {
		return err
	}

	addOwnerRefToObject(promRule, owner)

	err = sdk.Create(promRule)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	//TODO: handle update - use retry.RetryOnConflict

	return nil
}

func buildPrometheusRule(ruleName string, namespace string, labels map[string]string) (*monitoringv1.PrometheusRule, error) {
	alertsRuleSpec, err := ruleSpec(utils.LookupEnvWithDefault("ALERTS_FILE_PATH", alertsFilePath))
	if err != nil {
		return nil, err
	}
	rulesRuleSpec, err := ruleSpec(utils.LookupEnvWithDefault("RULES_FILE_PATH", rulesFilePath))
	if err != nil {
		return nil, err
	}

	alertsRuleSpec.Groups = append(alertsRuleSpec.Groups, rulesRuleSpec.Groups...)

	rule := prometheusRule(ruleName, namespace, labels)
	rule.Spec = *alertsRuleSpec

	return rule, nil
}

func prometheusRule(ruleName, namespace string, labels map[string]string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: monitoringv1.Group + "/" + monitoringv1.Version,
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      ruleName,
			Namespace: namespace,
			Labels:    labels,
		},
	}
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
