package k8shandler

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	alertsFilePath = "/etc/elasticsearch-operator/files/prometheus_alerts.yml"
	rulesFilePath  = "/etc/elasticsearch-operator/files/prometheus_rules.yml"
)

func (elasticsearchRequest *ElasticsearchRequest) CreateOrUpdatePrometheusRules() error {

	dpl := elasticsearchRequest.cluster

	ruleName := fmt.Sprintf("%s-%s", dpl.Name, "prometheus-rules")
	owner := getOwnerRef(dpl)

	promRule, err := buildPrometheusRule(ruleName, dpl.Namespace, dpl.Labels)
	if err != nil {
		return err
	}

	addOwnerRefToObject(promRule, owner)

	err = elasticsearchRequest.client.Create(context.TODO(), promRule)
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
			Kind:       monitoringv1.DefaultCrdKinds.PrometheusRule.Kind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
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
