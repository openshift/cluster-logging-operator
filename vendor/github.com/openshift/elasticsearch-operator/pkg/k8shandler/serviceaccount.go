package k8shandler

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateServiceAccount ensures the existence of the serviceaccount for Elasticsearch cluster
func CreateOrUpdateServiceAccount(dpl *v1alpha1.Elasticsearch) (err error) {

	err = createOrUpdateServiceAccount(dpl.Name, dpl.Namespace, getOwnerRef(dpl))
	if err != nil {
		return fmt.Errorf("Failure creating ServiceAccount %v", err)
	}

	return nil
}

func createOrUpdateServiceAccount(serviceAccountName, namespace string, ownerRef metav1.OwnerReference) error {
	serviceAccount := newServiceAccount(serviceAccountName, namespace)
	addOwnerRefToObject(serviceAccount, ownerRef)

	err := sdk.Create(serviceAccount)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing serviceaccount for the Elasticsearch cluster: %v", err)
		}
	}

	return nil
}

// serviceAccount returns a v1.ServiceAccount object
func newServiceAccount(serviceAccountName string, namespace string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}
}
