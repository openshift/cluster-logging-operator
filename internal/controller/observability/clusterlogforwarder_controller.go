package observability

import (
	"context"
	"encoding/json"
	"time"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalinit "github.com/openshift/cluster-logging-operator/internal/api/initialize"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	validations "github.com/openshift/cluster-logging-operator/internal/validations/observability"
	"github.com/openshift/cluster-logging-operator/version"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/set"
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
	Scheme *runtime.Scheme

	NewForwarderContext func() internalcontext.ForwarderContext

	PollInterval time.Duration

	TimeOut time.Duration
}

func (r *ClusterLogForwarderReconciler) Reconcile(_ context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := log.WithName(loggerName)
	log.V(3).Info("reconcile", "namespace", req.Namespace, "name", req.Name)

	cxt := r.NewForwarderContext()
	if cxt.Forwarder, err = FetchClusterLogForwarder(cxt.Client, req.Namespace, req.Name); err != nil {
		if !errors.IsNotFound(err) {
			// Other error, so requeue the request
			return defaultRequeue, err
		}
		// Stop reconciliation because resource is not present anymore
		return defaultRequeue, nil
	}

	if cxt.Forwarder.DeletionTimestamp != nil {
		// Resource is being deleted, no further reconciliation
		return defaultRequeue, nil
	}

	removeStaleStatuses(cxt.Forwarder)

	readyCond := internalobs.NewCondition(obsv1.ConditionTypeReady, obsv1.ConditionUnknown, obsv1.ReasonUnknownState, "")
	defer func() {
		updateStatus(cxt.Client, cxt.Forwarder, readyCond)
	}()

	if cxt.Forwarder.Spec.ManagementState == obsv1.ManagementStateUnmanaged {
		readyCond.Reason = obsv1.ReasonManagementStateUnmanaged
		readyCond.Message = "Updates are ignored when the managementState is Unmanaged"
		return defaultRequeue, nil
	}

	readyCond.Status = obsv1.ConditionFalse
	if cxt, err = initialize(cxt); err != nil {
		readyCond.Reason = obsv1.ReasonInitializationFailed
		readyCond.Message = err.Error()
		return defaultRequeue, nil
	}

	if !validateForwarder(cxt) {
		readyCond.Reason = obsv1.ReasonValidationFailure
		readyCond.Message = "collector not ready"
		if validations.MustUndeployCollector(cxt.Forwarder.Status.Conditions) {
			if deleteErr := collector.Remove(cxt.Client, cxt.Forwarder.Namespace, cxt.Forwarder.Name); deleteErr != nil {
				log.V(0).Error(deleteErr, "Unable to remove collector deployment")
			}
		}
		return defaultRequeue, err
	}

	if err = RemoveStaleWorkload(cxt.Client, cxt.Forwarder); err != nil {
		readyCond.Reason = obsv1.ReasonFailureToRemoveStaleWorkload
		readyCond.Message = err.Error()
		return defaultRequeue, err
	}

	reconcileErr := ReconcileCollector(cxt, r.PollInterval, r.TimeOut)
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

// initialize evaluates the spec and initializes any values that can not be enforced with annotations or are implied
// in their usage (i.e. reserved input names)
func initialize(cxt internalcontext.ForwarderContext) (internalcontext.ForwarderContext, error) {
	log.V(4).Info("Initialize")
	var err error
	cxt.AdditionalContext = utils.Options{}
	migrated := internalinit.ClusterLogForwarder(*cxt.Forwarder, cxt.AdditionalContext)
	cxt.Forwarder = &migrated

	if cxt.Secrets, err = MapSecrets(cxt.Client, cxt.Forwarder.Namespace, cxt.Forwarder.Spec.Inputs, cxt.Forwarder.Spec.Outputs); err != nil {
		return cxt, err
	}

	if generatedSecrets, found := utils.GetOption[[]*corev1.Secret](cxt.AdditionalContext, internalinit.GeneratedSecrets, []*corev1.Secret{}); found {
		for _, secret := range generatedSecrets {
			cxt.Secrets[secret.Name] = secret
		}
	}

	if cxt.ConfigMaps, err = MapConfigMaps(cxt.Client, cxt.Forwarder.Namespace, cxt.Forwarder.Spec.Inputs, cxt.Forwarder.Spec.Outputs); err != nil {
		return cxt, err
	}

	// Determine if on HCP and use the appropriate cluster Version/id
	clusterVersion, clusterID := version.HostedClusterVersion(context.TODO(), cxt.Reader, cxt.Forwarder.Namespace)
	if clusterVersion != "" && clusterID != "" {
		cxt.ClusterVersion = clusterVersion
		cxt.ClusterID = clusterID
	}
	return cxt, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterLogForwarderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&obsv1.ClusterLogForwarder{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.DaemonSet{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&appsv1.Deployment{}).
		Owns(&networkingv1.NetworkPolicy{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&v1.ServiceMonitor{}).
		Complete(r)
}

func validateForwarder(forwarderContext internalcontext.ForwarderContext) (valid bool) {
	validations.ValidateClusterLogForwarder(forwarderContext)

	validCond := internalobs.NewCondition(obsv1.ConditionTypeValid, obsv1.ConditionTrue, obsv1.ReasonValidationSuccess, "")
	if valid = internalobs.IsValidSpec(*forwarderContext.Forwarder); !valid {
		validCond.Status = obsv1.ConditionFalse
		validCond.Reason = obsv1.ReasonValidationFailure
		validCond.Message = "one or more conditions [inputs, outputs, pipelines, filters] have failed validation"
	}
	internalobs.SetCondition(&forwarderContext.Forwarder.Status.Conditions, validCond)
	return valid
}

func updateStatus(k8Client client.Client, instance *obsv1.ClusterLogForwarder, ready metav1.Condition) {
	internalobs.SetCondition(&instance.Status.Conditions, ready)
	jsonPatch, _ := json.Marshal(map[string]interface{}{
		"status": instance.Status,
	})
	if err := k8Client.Status().Patch(context.TODO(), instance, client.RawPatch(types.MergePatchType, jsonPatch)); err != nil {
		log.Error(err, "Error updating status", "status", instance.Status)
	}
}

func removeStaleStatuses(forwarder *obsv1.ClusterLogForwarder) {
	inputs := internalobs.Inputs(forwarder.Spec.Inputs)
	outputs := internalobs.Outputs(forwarder.Spec.Outputs)
	filters := internalobs.Filters(forwarder.Spec.Filters)
	pipelines := internalobs.Pipelines(forwarder.Spec.Pipelines)
	internalobs.PruneConditions(&forwarder.Status.InputConditions, inputs, obsv1.ConditionTypeValidInputPrefix)
	internalobs.PruneConditions(&forwarder.Status.OutputConditions, outputs, obsv1.ConditionTypeValidOutputPrefix)
	internalobs.PruneConditions(&forwarder.Status.FilterConditions, filters, obsv1.ConditionTypeValidFilterPrefix)
	internalobs.PruneConditions(&forwarder.Status.PipelineConditions, pipelines, obsv1.ConditionTypeValidPipelinePrefix)
}
