package runtime

import (
	"bytes"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

func NewPrometheusRule(namespace, ruleName string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ruleName,
			Namespace: namespace,
		},
	}
}

func NewPrometheusRuleSpecFrom(groups string) (*monitoringv1.PrometheusRuleSpec, error) {
	ruleSpec := monitoringv1.PrometheusRuleSpec{}
	if err := k8sYAML.NewYAMLOrJSONDecoder(bytes.NewBufferString(groups), 1000).Decode(&ruleSpec); err != nil {
		return nil, err
	}
	return &ruleSpec, nil
}
