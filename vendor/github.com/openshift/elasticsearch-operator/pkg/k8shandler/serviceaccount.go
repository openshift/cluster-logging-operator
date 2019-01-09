package k8shandler

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// CreateOrUpdateServiceAccount ensures the existence of the serviceaccount for Elasticsearch cluster
func CreateOrUpdateServiceAccount(dpl *v1alpha1.Elasticsearch) (string, error) {
	serviceAccountName := v1alpha1.ServiceAccountName

	owner := asOwner(dpl)

	err := createOrUpdateServiceAccount(serviceAccountName, dpl.Namespace, owner)
	if err != nil {
		return serviceAccountName, fmt.Errorf("Failure creating ServiceAccount %v", err)
	}

	return serviceAccountName, nil
}

func createOrUpdateServiceAccount(serviceAccountName, namespace string, owner metav1.OwnerReference) error {
	elasticsearchSA := serviceAccount(serviceAccountName, namespace)
	addOwnerRefToObject(elasticsearchSA, owner)
	err := sdk.Get(elasticsearchSA)
	if err != nil {
		err = sdk.Create(elasticsearchSA)
		if err != nil {
			return fmt.Errorf("Failure constructing serviceaccount for the Elasticsearch cluster: %v", err)
		}
	}
	return nil
}

// serviceAccount returns a v1.ServiceAccount object
func serviceAccount(serviceAccountName string, namespace string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}
}
