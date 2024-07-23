package observability

import (
	"context"
	"github.com/openshift/cluster-logging-operator/internal/api/initialize"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	validations "github.com/openshift/cluster-logging-operator/internal/validations/observability"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	loggerName = "controller.observability"
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
	log := log.WithName(loggerName)
	log.V(3).Info("reconcile", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)

	if r.Forwarder, err = FetchClusterLogForwarder(r.Client, req.NamespacedName.Namespace, req.NamespacedName.Name); err != nil {
		return defaultRequeue, err
	}

	removeStaleStatuses(r.Forwarder)

	readyCond := internalobs.NewCondition(obsv1.ConditionTypeReady, obsv1.ConditionUnknown, obsv1.ReasonUnknownState, "")
	defer func() {
		updateStatus(r.Client, r.Forwarder, readyCond)
	}()

	if r.Forwarder.Spec.ManagementState == obsv1.ManagementStateUnmanaged {
		readyCond.Reason = obsv1.ReasonManagementStateUnmanaged
		readyCond.Message = "Updates are ignored when the managementState is Unmanaged"
		return defaultRequeue, nil
	}

	readyCond.Status = obsv1.ConditionFalse
	if err = r.Initialize(); err != nil {
		readyCond.Reason = obsv1.ReasonInitializationFailed
		readyCond.Message = err.Error()
		return defaultRequeue, nil
	}

	if !validateForwarder(r.ForwarderContext) {
		readyCond.Reason = obsv1.ReasonValidationFailure
		if validations.MustUndeployCollector(r.Forwarder.Status.Conditions) {
			if deleteErr := collector.Remove(r.Client, r.Forwarder.Namespace, r.Forwarder.Name); deleteErr != nil {
				log.V(0).Error(deleteErr, "Unable to remove collector deployment")
			}
		}
		return defaultRequeue, err
	}

	if err = RemoveStaleWorkload(r.Client, r.Forwarder); err != nil {
		readyCond.Reason = obsv1.ReasonFailureToRemoveStaleWorkload
		readyCond.Message = err.Error()
		return defaultRequeue, err
	}

	reconcileErr := ReconcileCollector(r.ForwarderContext, collector.DefaultPollInterval, collector.DefaultTimeOut)
	if reconcileErr != nil {
		log.V(2).Error(reconcileErr, "reconcile error")
		readyCond.Reason = obsv1.ReasonDeploymentError
		readyCond.Message = reconcileErr.Error()
		return defaultRequeue, reconcileErr
	}
	readyCond.Reason = obsv1.ReasonReconciliationComplete
	readyCond.Status = obsv1.ConditionTrue

	return periodicRequeue, nil
}

// RemoveStaleWorkload removes existing workload if the ClusterLogForwarder was modified such that the deployment will change
// from a daemonSet to a deployment or vise versa
func RemoveStaleWorkload(k8Client client.Client, forwarder *obsv1.ClusterLogForwarder) error {
	remove := collector.RemoveDeployment
	if internalobs.DeployAsDeployment(*forwarder) {
		remove = collector.Remove
	}
	return remove(k8Client, forwarder.Namespace, forwarder.Name)
}

func MapSecrets(k8Client client.Client, namespace string, inputs internalobs.Inputs, outputs internalobs.Outputs) (secretMap map[string]*corev1.Secret, err error) {
	names := set.New(inputs.SecretNames()...)
	names.Insert(outputs.SecretNames()...)
	log.WithName(loggerName).V(4).Info("MapSecrets", "names", names.SortedList())
	secretMap = map[string]*corev1.Secret{}
	var secrets []*corev1.Secret
	if secrets, err = FetchSecrets(k8Client, namespace, names.UnsortedList()...); err != nil {
		return nil, err
	}
	for _, s := range secrets {
		log.WithName(loggerName).V(4).Info("fetched", "name", s.Name)
		secretMap[s.Name] = s
	}
	return secretMap, nil
}

func MapConfigMaps(k8Client client.Client, namespace string, inputs internalobs.Inputs, outputs internalobs.Outputs) (configMaps map[string]*corev1.ConfigMap, err error) {
	names := set.New(inputs.ConfigmapNames()...)
	names.Insert(outputs.ConfigmapNames()...)
	log.WithName(loggerName).V(4).Info("MapConfigMaps", "names", names.SortedList())
	configMaps = map[string]*corev1.ConfigMap{}
	var configs []*corev1.ConfigMap
	if configs, err = FetchConfigMaps(k8Client, namespace, names.UnsortedList()...); err != nil {
		return nil, err
	}
	for _, cm := range configs {
		log.WithName(loggerName).V(4).Info("fetched", "name", cm.Name)
		configMaps[cm.Name] = cm
	}
	return configMaps, nil
}

// Initialize evaluates the spec and initializes any values that can not be enforced with annotations or are implied
// in their usage (i.e. reserved input names)
func (r *ClusterLogForwarderReconciler) Initialize() (err error) {
	log.V(4).Info("Initialize")
	r.AdditionalContext = utils.Options{}
	migrated := initialize.ClusterLogForwarder(*r.Forwarder, r.AdditionalContext)
	r.Forwarder = &migrated

	if r.Secrets, err = MapSecrets(r.Client, r.Forwarder.Namespace, r.Forwarder.Spec.Inputs, r.Forwarder.Spec.Outputs); err != nil {
		return err
	}

	if generatedSecrets, found := utils.GetOption[[]*corev1.Secret](r.AdditionalContext, initialize.GeneratedSecrets, []*corev1.Secret{}); found {
		for _, secret := range generatedSecrets {
			r.Secrets[secret.Name] = secret
		}
	}

	if r.ConfigMaps, err = MapConfigMaps(r.Client, r.Forwarder.Namespace, r.Forwarder.Spec.Inputs, r.Forwarder.Spec.Outputs); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterLogForwarderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&obsv1.ClusterLogForwarder{}).
		Complete(r)
}

func validateForwarder(forwarderContext internalcontext.ForwarderContext) (valid bool) {
	validations.ValidateClusterLogForwarder(forwarderContext)

	validCond := internalobs.NewCondition(obsv1.ConditionTypeValid, obsv1.ConditionTrue, obsv1.ReasonValidationSuccess, "")
	if valid = internalobs.IsValid(*forwarderContext.Forwarder); !valid {
		validCond.Status = obsv1.ConditionFalse
		validCond.Reason = obsv1.ReasonValidationFailure
		validCond.Message = "one or more of inputs, outputs, pipelines, filters have a validation failure"
	}
	internalobs.SetCondition(&forwarderContext.Forwarder.Status.Conditions, validCond)
	return valid
}

func updateStatus(k8Client client.Client, instance *obsv1.ClusterLogForwarder, ready metav1.Condition) {
	internalobs.SetCondition(&instance.Status.Conditions, ready)
	if err := k8Client.Status().Update(context.TODO(), instance); err != nil {
		log.Error(err, "clusterlogforwarder-controller error updating status", "status", instance.Status)
	}
}

func removeStaleStatuses(forwarder *obsv1.ClusterLogForwarder) {
	inputs := internalobs.Inputs(forwarder.Spec.Inputs)
	outputs := internalobs.Outputs(forwarder.Spec.Outputs)
	filters := internalobs.Filters(forwarder.Spec.Filters)
	pipelines := internalobs.Pipelines(forwarder.Spec.Pipelines)
	internalobs.PruneConditions(&forwarder.Status.Inputs, inputs)
	internalobs.PruneConditions(&forwarder.Status.Outputs, outputs)
	internalobs.PruneConditions(&forwarder.Status.Filters, filters)
	internalobs.PruneConditions(&forwarder.Status.Pipelines, pipelines)
}
