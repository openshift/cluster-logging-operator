package service

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	kubernetes "sigs.k8s.io/controller-runtime/pkg/client"
)

// List returns a list of services in a namespace with a based on a key/value label and namespace
func List(client kubernetes.Reader, namespace, key, value string) (*core.ServiceList, error) {
	labelSelector, _ := labels.Parse(fmt.Sprintf("%s=%s", key, value))
	services := &core.ServiceList{}
	if err := client.List(context.TODO(), services, &kubernetes.ListOptions{LabelSelector: labelSelector, Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("failure listing services with label: %s,  %v", fmt.Sprintf("%s=%s", key, value), err)
	}
	return services, nil
}

func Delete(client kubernetes.Client, namespace, name string) error {
	service := runtime.NewService(namespace, name)
	if err := client.Delete(context.TODO(), service); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failure deleting service %s/%s: %v", namespace, name, err)
	}

	return nil
}
