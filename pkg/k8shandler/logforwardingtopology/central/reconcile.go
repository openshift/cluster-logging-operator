package central

import (
	"fmt"
	"reflect"

	"github.com/ViaQ/logerr/log"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/factory"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentbit"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology"
	topologyapi "github.com/openshift/cluster-logging-operator/pkg/k8shandler/logforwardingtopology"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/pkg/utils/comparators/services"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	collectorConfigName  = "fluentbit"
	normalizerConfigName = "normalizer"
)

type CentralNormalizerTopology struct {
	OwnerRef                metav1.OwnerReference
	APIClient               topologyapi.APIClient
	ReconcilePriorityClass  func() error
	ReconcileServiceAccount func() (*core.ServiceAccount, error)
	ReconcileConfigMap      func(configMap *core.ConfigMap) error
	ReconcileSecrets        func() error
	ReconcileTopology       func(proxyConfig *configv1.Proxy) error
	RemoveServiceAccount    func() error
	RemoveSecrets           func() error
	RemovePriorityClass     func() error
	GenerateCollectorConfig func() (string, error)
}

func (topology CentralNormalizerTopology) Name() string {
	return logforwardingtopology.LogForwardingCentralNormalizationTopology
}
func (topology CentralNormalizerTopology) ProcessConfigMap(cm *v1.ConfigMap) *v1.ConfigMap {
	return cm
}
func (topology CentralNormalizerTopology) ProcessPodSpec(podSpec *v1.PodSpec) *v1.PodSpec {
	return podSpec
}
func (topology CentralNormalizerTopology) ProcessService(service *v1.Service) *v1.Service {
	return service
}
func (topology CentralNormalizerTopology) ProcessServiceMonitor(sm *monitoringv1.ServiceMonitor) *monitoringv1.ServiceMonitor {
	return sm
}
func (topology CentralNormalizerTopology) Reconcile(proxyConfig *configv1.Proxy) (err error) {

	if err = topology.ReconcilePriorityClass(); err != nil {
		return
	}
	var serviceAccount *core.ServiceAccount
	if serviceAccount, err = topology.ReconcileServiceAccount(); err != nil {
		return
	}
	if serviceAccount != nil {
		// remove our finalizer from the list and update it.
		serviceAccount.ObjectMeta.Finalizers = utils.RemoveString(serviceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
		if err = topology.APIClient.Update(serviceAccount); err != nil {
			log.Error(err, "Unable to update the collector serviceaccount", "sa", serviceAccount.Name)
			return nil
		}
	}
	if err = ReconcileServices(topology.APIClient, topology.OwnerRef); err != nil {
		return
	}

	//SM
	//PromRules
	//genconfig
	if err = topology.reconcileConfigMaps(topology.ReconcileConfigMap); err != nil {
		return
	}
	if err = topology.ReconcileSecrets(); err != nil {
		return
	}

	//need proxy config here for normalizer?
	//add conf hash
	if err = reconcileNormalizer(topology.APIClient, topology.OwnerRef); err != nil {
		return
	}
	//add conf hash
	if err = reconcileCollector(topology.APIClient, topology.OwnerRef); err != nil {
		return
	}
	return nil
}

func (topology CentralNormalizerTopology) Undeploy() (err error) {
	if err = topology.APIClient.Delete(NewCollector()); err != nil {
		log.Error(err, "Unable to remove collector")
	}
	if err = topology.APIClient.Delete(NewNormalizer()); err != nil {
		log.Error(err, "Unable to remove normalizer")
	}
	if err = topology.RemoveSecrets(); err != nil {
		log.Error(err, "Unable to remove secrets")
	}
	config := factory.NewConfigMap(collectorConfigName, constants.OpenshiftNS, map[string]string{})
	if err = topology.APIClient.Delete(config); err != nil {
		log.Error(err, "Unable to remove configmap")
	}
	config = factory.NewConfigMap(normalizerConfigName, constants.OpenshiftNS, map[string]string{})
	if err = topology.APIClient.Delete(config); err != nil {
		log.Error(err, "Unable to remove configmap")
	}
	if err = topology.APIClient.Delete(NewService()); err != nil {
		log.Error(err, "Unable to remove normalizer service")
	}
	if err = topology.RemoveServiceAccount(); err != nil {
		log.Error(err, "Unable to remove serviceaccount")
	}
	if err = topology.RemovePriorityClass(); err != nil {
		log.Error(err, "Unable to remove priorityclass")
	}
	return nil
}

func ReconcileServices(apiClient topologyapi.APIClient, ownerRef metav1.OwnerReference) (err error) {

	desired := NewService()
	utils.AddOwnerRefToObject(desired, ownerRef)
	err = apiClient.Create(desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating service %q: %v", desired.Name, err)
		}

		current := &core.Service{}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := apiClient.Get(desired.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get service %q: %v", desired.Name, err)
			}
			if services.AreSame(current, desired) {
				return nil
			}
			//Explicitly copying because services are immutable
			current.Labels = desired.Labels
			current.Spec.Selector = desired.Spec.Selector
			current.Spec.Ports = desired.Spec.Ports
			return apiClient.Update(current)
		})
	}
	return err

}

func (topology CentralNormalizerTopology) reconcileConfigMaps(reconcileConfigMap func(configMap *core.ConfigMap) error) (err error) {
	//collector
	data := map[string]string{
		"fluent-bit.conf": fluentbitConf,
		"parsers.conf":    fluentbit.Parsers,
		"concat-crio.lua": fluentbit.ConcatCrioFilter,
	}
	configmap := factory.NewConfigMap(collectorConfigName, constants.OpenshiftNS, data)
	if err = reconcileConfigMap(configmap); err != nil {
		return err
	}
	//normalizer
	fluentConf, err := topology.GenerateCollectorConfig()
	if err != nil {
		return err
	}
	data = map[string]string{
		"fluent.conf": fluentConf,
		"run.sh":      string(utils.GetFileContents(utils.GetShareDir() + "/fluentd/run.sh")),
	}
	configmap = factory.NewConfigMap(normalizerConfigName, constants.OpenshiftNS, data)
	if err = reconcileConfigMap(configmap); err != nil {
		return err
	}
	return nil
}

func reconcileCollector(apiClient topologyapi.APIClient, ownerRef metav1.OwnerReference) (err error) {
	desired := NewCollector()
	utils.AddOwnerRefToObject(desired, ownerRef)
	err = apiClient.Create(desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating %s/%s: %v", desired.Kind, desired.Name, err)
		}

		current := &apps.DaemonSet{}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := apiClient.Get(desired.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %s/%s: %v", desired.Kind, desired.Name, err)
			}
			//TODO: More explicit comparison
			if reflect.DeepEqual(current.Spec, desired.Spec) {
				return nil
			}
			current.Spec = desired.Spec
			return apiClient.Update(current)
		})
	}

	return err
}

func reconcileNormalizer(apiClient topologyapi.APIClient, ownerRef metav1.OwnerReference) (err error) {
	desired := NewNormalizer()
	utils.AddOwnerRefToObject(desired, ownerRef)
	err = apiClient.Create(desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating %s/%s: %v", desired.Kind, desired.Name, err)
		}

		current := &apps.Deployment{}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := apiClient.Get(desired.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %s/%s: %v", desired.Kind, desired.Name, err)
			}
			//TODO: More explicit comparison
			if reflect.DeepEqual(current.Spec, desired.Spec) {
				return nil
			}
			current.Spec = desired.Spec
			return apiClient.Update(current)
		})
	}
	return err
}
