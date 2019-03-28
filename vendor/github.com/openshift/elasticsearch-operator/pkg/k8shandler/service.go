package k8shandler

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateServices ensures the existence of the services for Elasticsearch cluster
func CreateOrUpdateServices(dpl *api.Elasticsearch) error {
	ownerRef := getOwnerRef(dpl)
	annotations := make(map[string]string)

	err := createOrUpdateService(
		fmt.Sprintf("%s-%s", dpl.Name, "cluster"),
		dpl.Namespace,
		dpl.Name,
		"cluster",
		9300,
		selectorForES("es-node-master", dpl.Name),
		annotations,
		true,
		ownerRef,
	)
	if err != nil {
		return fmt.Errorf("Failure creating service %v", err)
	}

	err = createOrUpdateService(
		dpl.Name,
		dpl.Namespace,
		dpl.Name,
		"restapi",
		9200,
		selectorForES("es-node-client", dpl.Name),
		annotations,
		false,
		ownerRef,
	)
	if err != nil {
		return fmt.Errorf("Failure creating service %v", err)
	}

	annotations["service.alpha.openshift.io/serving-cert-secret-name"] = fmt.Sprintf("%s-%s", dpl.Name, "metrics")
	err = createOrUpdateService(
		fmt.Sprintf("%s-%s", dpl.Name, "metrics"),
		dpl.Namespace,
		dpl.Name,
		"metrics",
		9200,
		selectorForES("es-node-client", dpl.Name),
		annotations,
		false,
		ownerRef,
	)
	if err != nil {
		return fmt.Errorf("Failure creating service %v", err)
	}
	return nil
}

func createOrUpdateService(serviceName, namespace, clusterName, targetPortName string, port int32, selector, annotations map[string]string, publishNotReady bool, owner metav1.OwnerReference) error {

	labels := appendDefaultLabel(clusterName, map[string]string{})

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
	addOwnerRefToObject(service, owner)

	err := sdk.Create(service)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing %v service: %v", service.Name, err)
		}

		current := service.DeepCopy()
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = sdk.Get(current); err != nil {
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
			if err = sdk.Update(current); err != nil {
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
				v1.ServicePort{
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
