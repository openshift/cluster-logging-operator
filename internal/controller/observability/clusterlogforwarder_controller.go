package observability

import (
	"context"
	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	obsmigrate "github.com/openshift/cluster-logging-operator/internal/migrations/observability"
	validations "github.com/openshift/cluster-logging-operator/internal/validations/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/set"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"

	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// periodicRequeue to ensure CLF collection permissions are still valid.  We can not watch
	// ClusterRoleBindings since there is no effective way to associate known CLF with a given binding to
	// avoid needing to reconcile all CRB events
	periodicRequeue = ctrl.Result{
		RequeueAfter: time.Minute * 5,
	}

	defaultRequeue = ctrl.Result{}
)

// ClusterLogForwarderReconciler reconciles a ClusterLogForwarder object
type ClusterLogForwarderReconciler struct {
	internalcontext.ForwarderContext
	Scheme *runtime.Scheme
}

func (r *ClusterLogForwarderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := log.WithName("ClusterLogForwarderReconciler.reconcile")
	log.V(3).Info("obs.clf controller reconciling resource", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)

	if r.Forwarder, err = FetchClusterLogForwarder(r.Client, req.NamespacedName.Namespace, req.NamespacedName.Name); err != nil {
		return defaultRequeue, err
	}
	if r.Forwarder.Spec.ManagementState == observabilityv1.ManagementStateUnmanaged {
		return defaultRequeue, nil
	}

	if result, err = Initialize(r.Client, r.Forwarder); err != nil {
		return result, err
	}

	if r.Secrets, err = MapSecrets(r.Client, req.Namespace, r.Forwarder.Spec.Inputs, r.Forwarder.Spec.Outputs); err != nil {
		return defaultRequeue, err
	}

	if r.ConfigMaps, err = MapConfigMaps(r.Client, req.Namespace, r.Forwarder.Spec.Inputs, r.Forwarder.Spec.Outputs); err != nil {
		return defaultRequeue, err
	}

	if result, err = validateForwarder(r.ForwarderContext); err != nil {
		return result, err
	}

	//TODO: Remove deployment if unready? - add to "validate" logic of 'must-undeploy'
	//TODO: Remove existing deployment/daemonset
	//TODO: Remove stale input services

	reconcileErr := ReconcileCollector(r.Client, r.Reader, *r.Forwarder, r.ClusterID, collector.DefaultPollInterval, collector.DefaultTimeOut)
	if reconcileErr != nil {
		log.V(2).Error(reconcileErr, "clusterlogforwarder-controller returning, error")
		//} else {
		//	//TODO: Update conditions
	}

	return periodicRequeue, reconcileErr
}

func MapSecrets(k8Client client.Client, namespace string, inputs internalobs.Inputs, outputs internalobs.Outputs) (map[string]*corev1.Secret, error) {
	names := set.New(inputs.SecretNames()...)
	names.Insert(outputs.SecretNames()...)
	secretMap := map[string]*corev1.Secret{}
	if secrets, err := FetchSecrets(k8Client, namespace, names.UnsortedList()...); err != nil {
		return nil, err
	} else {
		for _, s := range secrets {
			secretMap[s.Name] = s
		}
	}

	return secretMap, nil
}

func MapConfigMaps(k8Client client.Client, namespace string, inputs internalobs.Inputs, outputs internalobs.Outputs) (map[string]*corev1.ConfigMap, error) {
	names := set.New(inputs.ConfigmapNames()...)
	names.Insert(outputs.ConfigmapNames()...)
	configMapMap := map[string]*corev1.ConfigMap{}
	if configMaps, err := FetchConfigMaps(k8Client, namespace, names.UnsortedList()...); err != nil {
		return nil, err
	} else {
		for _, cm := range configMaps {
			configMapMap[cm.Name] = cm
		}
	}
	return configMapMap, nil
}

// Initialize evaluates the spec and initializes any values that can not be enforced with annotations or are implied
// in their usage (i.e. reserved input names)
func Initialize(k8Client client.Client, forwarder *observabilityv1.ClusterLogForwarder) (ctrl.Result, error) {
	forwarder.Spec, _ = obsmigrate.MigrateClusterLogForwarder(forwarder.Spec)
	//TODO: FIX Conditions
	//condition := logging.NewCondition(logging.ValidationCondition, corev1.ConditionTrue, logging.ValidationFailureReason, "%v", err)
	//instance.Status.Conditions.SetCondition(condition)
	//instance.Status.Conditions.SetCondition(logging.CondNotReady(logging.ValidationFailureReason, ""))
	return updateStatus(k8Client, forwarder)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterLogForwarderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&observabilityv1.ClusterLogForwarder{}).
		Complete(r)
}

func validateForwarder(forwarderContext internalcontext.ForwarderContext) (result ctrl.Result, err error) {
	log.V(3).Info("obs-clusterlogforwarder-controller Error getting instance. It will be retried if other then 'NotFound'", "error", err.Error())
	if failures := validations.ValidateClusterLogForwarder(forwarderContext); len(failures) > 0 {
		// TODO: Evaluate failures
		//if validationerrors.MustUndeployCollector(err) {
		//	if deleteErr := collector.Remove(k8Client, req.Namespace, req.Name); deleteErr != nil {
		//		log.V(0).Error(deleteErr, "Unable to remove collector deployment")
		//	}
		//}
		// TODO: Determine if we need to "sync" conditions like in 5.9
		for attributeType, conditions := range failures {
			switch attributeType {
			case common.AttributeConditionConditions:
				forwarderContext.Forwarder.Status.Conditions = conditions
			case common.AttributeConditionInputs:
				forwarderContext.Forwarder.Status.Inputs = conditions
			case common.AttributeConditionOutputs:
				forwarderContext.Forwarder.Status.Outputs = conditions
			case common.AttributeConditionPipelines:
				forwarderContext.Forwarder.Status.Pipelines = conditions
			case common.AttributeConditionFilters:
				forwarderContext.Forwarder.Status.Filters = conditions
			}
		}
		return updateStatus(forwarderContext.Client, forwarderContext.Forwarder)
	}
	return defaultRequeue, err
}

func updateStatus(k8Client client.Client, instance *obsv1.ClusterLogForwarder) (ctrl.Result, error) {
	if err := k8Client.Status().Update(context.TODO(), instance); err != nil {

		if strings.Contains(err.Error(), constants.OptimisticLockErrorMsg) {
			// do manual retry without error
			// more information about this error here: https://github.com/kubernetes/kubernetes/issues/28149
			return reconcile.Result{RequeueAfter: time.Second * 1}, nil
		}

		log.Error(err, "clusterlogforwarder-controller error updating status")
		return ctrl.Result{}, err
	}

	return periodicRequeue, nil
}
