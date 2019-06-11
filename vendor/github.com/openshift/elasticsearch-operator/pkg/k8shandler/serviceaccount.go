package k8shandler

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateServiceAccount ensures the existence of the serviceaccount for Elasticsearch cluster
func (elasticsearchRequest *ElasticsearchRequest) CreateOrUpdateServiceAccount() (err error) {

	dpl := elasticsearchRequest.cluster

	err = createOrUpdateServiceAccount(dpl.Name, dpl.Namespace, getOwnerRef(dpl), elasticsearchRequest.client)
	if err != nil {
		return fmt.Errorf("Failure creating ServiceAccount %v", err)
	}

	return nil
}

func createOrUpdateServiceAccount(serviceAccountName, namespace string, ownerRef metav1.OwnerReference, client client.Client) error {
	serviceAccount := newServiceAccount(serviceAccountName, namespace)
	addOwnerRefToObject(serviceAccount, ownerRef)

	err := client.Create(context.TODO(), serviceAccount)
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
