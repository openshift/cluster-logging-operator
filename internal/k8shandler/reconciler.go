package k8shandler

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder"

	"github.com/openshift/cluster-logging-operator/internal/migrations"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/metrics"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/telemetry"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"k8s.io/apimachinery/pkg/types"
)

func Reconcile(cl *logging.ClusterLogging, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterID string) (instance *logging.ClusterLogging, err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Cluster:       cl,
		Client:        requestClient,
		Reader:        reader,
		EventRecorder: r,
		ClusterID:     clusterID,
	}

	// Cancel previous info metrics
	telemetry.SetCLFMetrics(0)
	telemetry.SetCLMetrics(0)
	defer func() {
		telemetry.SetCLMetrics(1)
		telemetry.SetCLFMetrics(1)
	}()

	if instance, err = clusterLoggingRequest.getClusterLogging(false); err != nil {
		return nil, err
	}
	clusterLoggingRequest.Cluster = instance

	if instance.GetDeletionTimestamp() != nil {
		// ClusterLogging is being deleted, remove resources that can not be garbage-collected.
		if err := clusterLoggingRequest.removeLokiStackRbac(); err != nil {
			log.Error(err, "Error removing RBAC for accessing LokiStack.")
		}
	}

	if !clusterLoggingRequest.isManaged() {
		// if cluster is set to unmanaged then set managedStatus as 0
		telemetry.Data.CLInfo.Set("managedStatus", constants.UnManagedStatus)
		return clusterLoggingRequest.Cluster, nil
	}
	// CL is managed by default set it as 1
	telemetry.Data.CLInfo.Set("managedStatus", constants.ManagedStatus)
	updateInfofromCL(&clusterLoggingRequest)
	forwarder := clusterLoggingRequest.getLogForwarder()
	if forwarder != nil {
		if err := clusterlogforwarder.Validate(*forwarder); err != nil {
			return nil, err
		}
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	} else if !clusterLoggingRequest.IncludesManagedStorage() {
		// No clf and no logStore so remove the collector https://issues.redhat.com/browse/LOG-2703
		removeCollectorAndUpdate(clusterLoggingRequest)
		return clusterLoggingRequest.Cluster, nil
	}

	if clusterLoggingRequest.IncludesManagedStorage() {
		// Reconcile Log Store
		if err = clusterLoggingRequest.CreateOrUpdateLogStore(); err != nil {
			telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
			return clusterLoggingRequest.Cluster, fmt.Errorf("unable to create or update logstore for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		}

		// Reconcile Visualization
		if err = clusterLoggingRequest.CreateOrUpdateVisualization(); err != nil {
			telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
			return clusterLoggingRequest.Cluster, fmt.Errorf("unable to create or update visualization for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		}

	} else {
		removeManagedStorage(clusterLoggingRequest)
	}

	// Remove Curator
	if err := clusterLoggingRequest.removeCurator(); err != nil {
		log.V(0).Error(err, "Error removing curator component")
	}
	clusterLoggingRequest.Cluster.Status.Conditions.SetCondition(status.Condition{
		Type:    "CuratorRemoved",
		Status:  corev1.ConditionTrue,
		Reason:  "ResourceDeprecated",
		Message: "curator is deprecated in favor of defining retention policy",
	})

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(); err != nil {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		telemetry.Data.CollectorErrorCount.Inc("CollectorErrorCount")
		return clusterLoggingRequest.Cluster, fmt.Errorf("unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile metrics Dashboards
	if err = metrics.ReconcileDashboards(clusterLoggingRequest.Client, reader, clusterLoggingRequest.Cluster.Spec.Collection); err != nil {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		log.Error(err, "Unable to create or update metrics dashboards", "clusterName", clusterLoggingRequest.Cluster.Name)
	}

	//if there is no early exit from reconciler then new CL spec is applied successfully hence healthStatus is set to true or 1
	telemetry.Data.CLInfo.Set("healthStatus", constants.HealthyStatus)
	return clusterLoggingRequest.Cluster, nil
}

func removeCollectorAndUpdate(clusterRequest ClusterLoggingRequest) {
	log.V(3).Info("forwarder not found and logStore not found so removing collector")
	if err := clusterRequest.removeCollector(constants.CollectorName); err != nil {
		log.Error(err, "Error removing collector")
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
	}

	if updateError := clusterRequest.UpdateCondition(
		logging.CollectorDeadEnd,
		"Collectors are defined but there is no defined LogStore or LogForward destinations",
		"No defined logstore or logforward destination",
		corev1.ConditionTrue,
	); updateError != nil {
		log.Error(updateError, "Unable to update the clusterlogging status", "conditionType", logging.CollectorDeadEnd)
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
	}
}

func removeManagedStorage(clusterRequest ClusterLoggingRequest) {
	log.V(1).Info("Removing managed store components...")
	for _, remove := range []func() error{clusterRequest.removeElasticsearch, clusterRequest.removeKibana, clusterRequest.removeLokiStackRbac} {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		if err := remove(); err != nil {
			log.Error(err, "Error removing component")
		}
	}
}

func ReconcileForClusterLogForwarder(forwarder *logging.ClusterLogForwarder, requestClient client.Client, er record.EventRecorder, clusterID string) (err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Client:        requestClient,
		EventRecorder: er,
		ClusterID:     clusterID,
	}
	if forwarder != nil {
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	}

	var clusterLogging *logging.ClusterLogging
	if clusterLogging, err = clusterLoggingRequest.getClusterLogging(false); err != nil {
		return err
	}
	if clusterLogging == nil {
		return nil
	}
	clusterLoggingRequest.Cluster = clusterLogging

	if clusterLogging.Spec.ManagementState == logging.ManagementStateUnmanaged {
		telemetry.Data.CLInfo.Set("managedStatus", constants.UnManagedStatus)
		return nil
	}

	// Reconcile Collection
	err = clusterLoggingRequest.CreateOrUpdateCollection()
	forwarder.Status = clusterLoggingRequest.ForwarderRequest.Status

	if err != nil {
		msg := fmt.Sprintf("Unable to reconcile collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
		log.Error(err, msg)
		return errors.New(msg)
	}

	///////
	// if it reaches to this point without throwing any errors than mark CLF in healthy state as with '1' value and also CL in healthy state with '1' value
	telemetry.Data.CLFInfo.Set("healthStatus", constants.HealthyStatus)
	updateInfofromCLF(&clusterLoggingRequest)
	///////

	return nil
}

func (clusterRequest *ClusterLoggingRequest) getClusterLogging(skipMigrations bool) (*logging.ClusterLogging, error) {
	clusterLoggingNamespacedName := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.WatchNamespace}
	clusterLogging := &logging.ClusterLogging{}

	if err := clusterRequest.Client.Get(context.TODO(), clusterLoggingNamespacedName, clusterLogging); err != nil {
		return nil, err
	}

	// Do not modify cached copy
	clusterLogging = clusterLogging.DeepCopy()

	if skipMigrations {
		return clusterLogging, nil
	}

	// TODO Drop migration upon introduction of v2
	clusterLogging.Spec = migrations.MigrateCollectionSpec(clusterLogging.Spec)

	return clusterLogging, nil
}

func (clusterRequest *ClusterLoggingRequest) getLogForwarder() *logging.ClusterLogForwarder {
	nsname := types.NamespacedName{Name: constants.SingletonName, Namespace: clusterRequest.Cluster.Namespace}
	forwarder := runtime.NewClusterLogForwarder(clusterRequest.Cluster.Namespace, clusterRequest.Cluster.Name)
	if err := clusterRequest.Client.Get(context.TODO(), nsname, forwarder); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Encountered unexpected error getting", "forwarder", nsname)
		}
		forwarder.Spec = logging.ClusterLogForwarderSpec{}
	}
	forwarder.Spec = migrations.MigrateClusterLogForwarderSpec(forwarder.Spec, clusterRequest.Cluster.Spec.LogStore)
	return forwarder
}

func updateInfofromCL(request *ClusterLoggingRequest) {
	clspec := request.Cluster.Spec
	if clspec.LogStore != nil && clspec.LogStore.Type != "" {
		log.V(1).Info("LogStore Type", "clspecLogStoreType", clspec.LogStore.Type)
		telemetry.Data.CLLogStoreType.Set(string(clspec.LogStore.Type), constants.IsPresent)
	}
}

func updateInfofromCLF(request *ClusterLoggingRequest) {

	var npipelines = 0
	var output *logging.OutputSpec
	var found bool

	//CLO got two custom resources CL, CFL, CLF here is meant for forwarding logs to third party systems

	//CLO CLF pipelines and set of output specs
	lgpipeline := request.ForwarderSpec.Pipelines
	outputs := request.ForwarderSpec.OutputMap()
	log.V(1).Info("OutputMap", "outputs", outputs)

	for _, pipeline := range lgpipeline {
		npipelines++
		log.V(1).Info("pipelines", "npipelines", npipelines)
		inref := pipeline.InputRefs
		outref := pipeline.OutputRefs

		for labelname := range telemetry.Data.CLFInputType.M {
			log.V(1).Info("iter over labelnames", "labelname", labelname)
			telemetry.Data.CLFInputType.Set(labelname, constants.IsNotPresent) //reset to zero
			for _, inputtype := range inref {
				log.V(1).Info("iter over inputtype", "inputtype", inputtype)
				if inputtype == labelname {
					log.V(1).Info("labelname and inputtype", "labelname", labelname, "inputtype", inputtype) //when matched print matched labelname with input type stated in CLF spec
					telemetry.Data.CLFInputType.Set(labelname, constants.IsPresent)                          //input type present in CLF spec
				}
			}
		}

		for labelname := range telemetry.Data.CLFOutputType.M {
			log.V(1).Info("iter over labelnames", "labelname", labelname)
			telemetry.Data.CLFOutputType.Set(labelname, constants.IsNotPresent) //reset to zero
			for _, outputname := range outref {
				log.V(1).Info("iter over outref", "outputname", outputname)
				if outputname == "default" {
					telemetry.Data.CLFOutputType.Set("default", constants.IsPresent)
					continue
				}
				output, found = outputs[outputname]
				if found {
					outputtype := output.Type
					if outputtype == labelname {
						log.V(1).Info("labelname and outputtype", "labelname", labelname, "outputtype", outputtype)
						telemetry.Data.CLFOutputType.Set(labelname, constants.IsPresent) //when matched print matched labelname with output type stated in CLF spec
					}
				}
			}
		}
		log.V(1).Info("post updating inputtype and outputtype")
		telemetry.Data.CLFInfo.Set("pipelineInfo", strconv.Itoa(npipelines))
	}
}
