package collector

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/runtime/service"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
	kubernetes "sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileInputServices evaluates receiver inputs and deploys services for them
func (f *Factory) ReconcileInputServices(k8sClient kubernetes.Client, k8sReader kubernetes.Reader, namespace string, owner metav1.OwnerReference, visitors func(o runtime.Object)) error {

	if err := RemoveOrphanedInputServices(k8sClient, k8sReader, namespace, f.ForwarderSpec, *f.ResourceNames, owner); err != nil {
		return err
	}

	for _, input := range f.ForwarderSpec.Inputs {
		var listenPort int32
		serviceName := f.ResourceNames.GenerateInputServiceName(input.Name)
		if input.Receiver != nil {
			listenPort = input.Receiver.Port
			if err := network.ReconcileInputService(k8sClient, namespace, serviceName, f.ResourceNames.CommonName, serviceName, listenPort, listenPort, input.Receiver.Type, owner, visitors); err != nil {
				return err
			}
		}
	}
	return nil
}

// RemoveOrphanedInputServices removes receiver input services owned by the current CLF but not defined in the spec
func RemoveOrphanedInputServices(client kubernetes.Client, reader kubernetes.Reader, namespace string, spec obs.ClusterLogForwarderSpec, resourceNames factory.ForwarderResourceNames, currOwner metav1.OwnerReference) error {

	for _, receiverType := range obs.ReceiverTypes {
		// Get list of input services by label/ namespace
		services, err := service.List(reader, namespace, constants.LabelLoggingInputServiceType, string(receiverType))
		if err != nil {
			return err
		}

		// Collect defined receiver inputs
		inputs := set.New[string]()
		for _, input := range spec.Inputs {
			if input.Type == obs.InputTypeReceiver && input.Receiver.Type == receiverType {
				inputs.Insert(resourceNames.GenerateInputServiceName(input.Name))
			}
		}

		// Remove services only if owned by current CLF and isn't defined
		for _, item := range services.Items {
			if utils.HasSameOwner(item.OwnerReferences, []metav1.OwnerReference{currOwner}) && !inputs.Has(item.Name) {
				if err := service.Delete(client, item.Namespace, item.Name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
