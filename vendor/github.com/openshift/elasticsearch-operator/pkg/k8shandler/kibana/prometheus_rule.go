package kibana

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"
)

func NewPrometheusRule(ruleName, namespace string) *monitoringv1.PrometheusRule {
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

func (clusterRequest *KibanaRequest) RemovePrometheusRule(ruleName string) error {

	promRule := NewPrometheusRule(ruleName, clusterRequest.cluster.Namespace)

	err := clusterRequest.Delete(promRule)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v prometheus rule: %v", promRule, err)
	}

	return nil
}

func NewPrometheusRuleSpecFrom(filePath string) (*monitoringv1.PrometheusRuleSpec, error) {
	if err := utils.CheckFileExists(filePath); err != nil {
		return nil, err
	}
	fileContent, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("'%s' not readable", filePath)
	}
	ruleSpec := monitoringv1.PrometheusRuleSpec{}
	if err := k8sYAML.NewYAMLOrJSONDecoder(bytes.NewBufferString(string(fileContent)), 1000).Decode(&ruleSpec); err != nil {
		return nil, err
	}
	return &ruleSpec, nil
}
