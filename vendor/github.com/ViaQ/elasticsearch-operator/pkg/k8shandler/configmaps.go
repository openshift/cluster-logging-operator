package k8shandler

import (
	"bytes"
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/ViaQ/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	//"github.com/sirupsen/logrus"
)

// CreateOrUpdateConfigMaps ensures the existens of ConfigMaps with Elasticsearch configuration
func CreateOrUpdateConfigMaps(dpl *v1alpha1.Elasticsearch) (string, error) {
	owner := asOwner(dpl)
	var configMapName string
	if dpl.Spec.ConfigMapName == "" {
		configMapName = dpl.Name
	} else {
		configMapName = dpl.Spec.ConfigMapName
	}

	// TODO: take all vars from CRD
	pathData := "- /elasticsearch/persistent/"
	err := createOrUpdateConfigMap(configMapName, dpl.Namespace, dpl.Name, defaultKibanaIndexMode, pathData, false, dpl.Spec.Secure.Disabled, owner, dpl.Labels)
	if err != nil {
		return configMapName, fmt.Errorf("Failure creating ConfigMap %v", err)
	}
	return configMapName, nil
}

func createOrUpdateConfigMap(configMapName, namespace, clusterName, kibanaIndexMode, pathData string,
	allowClusterReader bool, insecureCluster bool, owner metav1.OwnerReference, labels map[string]string) error {
	elasticsearchCM, err := createConfigMap(configMapName, namespace, clusterName, kibanaIndexMode, pathData, allowClusterReader, insecureCluster, labels)
	if err != nil {
		return err
	}
	addOwnerRefToObject(elasticsearchCM, owner)
	err = sdk.Create(elasticsearchCM)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Elasticsearch ConfigMap: %v", err)
	} else if errors.IsAlreadyExists(err) {
		// Get existing configMap to check if it is same as what we want
		existingCM := configMap(configMapName, namespace, labels)
		err = sdk.Get(existingCM)
		if err != nil {
			return fmt.Errorf("Unable to get Elasticsearch cluster configMap: %v", err)
		}

		// TODO: Compare existing configMap labels, selectors and port
	}
	return nil
}

func createConfigMap(configMapName, namespace, clusterName, kibanaIndexMode, pathData string,
	allowClusterReader bool, insecureCluster bool, labels map[string]string) (*v1.ConfigMap, error) {
	cm := configMap(configMapName, namespace, labels)
	cm.Data = map[string]string{}
	buf := &bytes.Buffer{}
	if err := renderEsYml(buf, allowClusterReader, kibanaIndexMode, pathData, insecureCluster); err != nil {
		return cm, err
	}
	cm.Data["elasticsearch.yml"] = buf.String()

	buf = &bytes.Buffer{}
	if err := renderLog4j2Properties(buf, defaultRootLogger); err != nil {
		return cm, err
	}
	cm.Data["log4j2.properties"] = buf.String()

	return cm, nil
}

// configMap returns a v1.ConfigMap object
func configMap(configMapName string, namespace string, labels map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
