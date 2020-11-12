package k8shandler

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateServices ensures the existence of the services for Elasticsearch cluster
func (elasticsearchRequest *ElasticsearchRequest) CreateOrUpdateServices() error {

	dpl := elasticsearchRequest.cluster
	annotations := make(map[string]string)

	err := elasticsearchRequest.createOrUpdateService(
		fmt.Sprintf("%s-%s", dpl.Name, "cluster"),
		dpl.Namespace,
		dpl.Name,
		"cluster",
		9300,
		selectorForES("es-node-master", dpl.Name),
		annotations,
		true,
		map[string]string{},
	)
	if err != nil {
		return fmt.Errorf("Failure creating service %v", err)
	}

	err = elasticsearchRequest.createOrUpdateService(
		dpl.Name,
		dpl.Namespace,
		dpl.Name,
		"restapi",
		9200,
		selectorForES("es-node-client", dpl.Name),
		annotations,
		false,
		map[string]string{},
	)
	if err != nil {
		return fmt.Errorf("Failure creating service %v", err)
	}

	//legacy metrics service that likely can be rolled into the single service that goes through the proxy
	annotations["service.alpha.openshift.io/serving-cert-secret-name"] = fmt.Sprintf("%s-%s", dpl.Name, "metrics")
	err = elasticsearchRequest.createOrUpdateService(
		fmt.Sprintf("%s-%s", dpl.Name, "metrics"),
		dpl.Namespace,
		dpl.Name,
		"metrics",
		60001,
		selectorForES("es-node-client", dpl.Name),
		annotations,
		false,
		map[string]string{
			"scrape-metrics": "enabled",
		},
	)
	if err != nil {
		return fmt.Errorf("Failure creating service %v", err)
	}
	return nil
}

func (er *ElasticsearchRequest) createOrUpdateService(serviceName, namespace, clusterName, targetPortName string, port int32, selector, annotations map[string]string, publishNotReady bool, labels map[string]string) error {

	client := er.client
	cluster := er.cluster

	labels = appendDefaultLabel(clusterName, labels)

	service := newService(
		serviceName,
		namespace,
		clusterName,
		targetPortName,
		port,
		selector,
		annotations,
		labels,
		publishNotReady,
	)

	cluster.AddOwnerRefTo(service)

	err := client.Create(context.TODO(), service)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing %v service: %v", service.Name, err)
		}

		current := service.DeepCopy()
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = client.Get(context.TODO(), types.NamespacedName{Name: current.Name, Namespace: current.Namespace}, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v service: %v", service.Name, err)
			}

			current.Spec.Ports = service.Spec.Ports
			current.Spec.Selector = service.Spec.Selector
			current.Spec.PublishNotReadyAddresses = service.Spec.PublishNotReadyAddresses
			current.Labels = service.Labels
			if err = client.Update(context.TODO(), current); err != nil {
				return err
			}
			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func newService(serviceName, namespace, clusterName, targetPortName string, port int32, selector, annotations, labels map[string]string, publishNotReady bool) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.ServiceSpec{
			Selector: selector,
			Ports: []v1.ServicePort{
				{
					Port:       port,
					Protocol:   "TCP",
					TargetPort: intstr.FromString(targetPortName),
					Name:       clusterName,
				},
			},
			PublishNotReadyAddresses: publishNotReady,
		},
	}
}
