package k8shandler

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	eslogstore "github.com/openshift/cluster-logging-operator/internal/logstore/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	logmetricexporter "github.com/openshift/cluster-logging-operator/internal/metrics/logfilemetricexporter"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"
	"github.com/openshift/cluster-logging-operator/internal/migrations"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"k8s.io/apimachinery/pkg/types"
)

func Reconcile(cl *logging.ClusterLogging, forwarder *logging.ClusterLogForwarder, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterVersion, clusterID string, resourceNames *factory.ForwarderResourceNames) (instance *logging.ClusterLogging, err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Cluster:        cl,
		Client:         requestClient,
		Reader:         reader,
		EventRecorder:  r,
		ClusterVersion: clusterVersion,
		ClusterID:      clusterID,
		ResourceOwner:  utils.AsOwner(cl),
		ResourceNames:  resourceNames,
	}
	if forwarder != nil {
		clusterLoggingRequest.Forwarder = forwarder
	}

	if collectionSpec, err := CheckCollectionType(cl, forwarder); err != nil {
		return clusterLoggingRequest.Cluster, err
	} else {
		clusterLoggingRequest.CollectionSpec = collectionSpec
	}

	// Cancel previous info metrics
	telemetry.SetCLFMetrics(0)
	telemetry.SetCLMetrics(0)
	telemetry.SetLFMEMetrics(0)
	defer func() {
		telemetry.SetCLMetrics(1)
		telemetry.SetCLFMetrics(1)
	}()

	if cl.GetDeletionTimestamp() != nil {
		// ClusterLogging is being deleted, remove resources that can not be garbage-collected.
		if err := lokistack.RemoveRbac(clusterLoggingRequest.Client, clusterLoggingRequest.removeFinalizer); err != nil {
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

	if !forwarder.Status.IsReady() && !clusterLoggingRequest.IncludesManagedStorage() {
		// No clf and no logStore so remove the collector https://issues.redhat.com/browse/LOG-2703
		removeCollectorAndUpdate(clusterLoggingRequest)
		return clusterLoggingRequest.Cluster, nil
	}

	// Create a LogFileMetricExporter if none exists
	if clusterLoggingRequest.Cluster.Name == constants.SingletonName {
		var lfmeInstance *loggingv1alpha1.LogFileMetricExporter
		if lfmeInstance, err = getLogFileMetricExporter(clusterLoggingRequest.Client); err != nil {
			return nil, err
		}

		// // No LFME so make a new one
		if lfmeInstance == nil {
			lfmeInstance = runtime.NewLogFileMetricExporter(constants.WatchNamespace, clusterLoggingRequest.Cluster.Name)
			if err := ReconcileForLogFileMetricExporter(
				*clusterLoggingRequest.Cluster,
				lfmeInstance,
				clusterLoggingRequest.Client,
				clusterLoggingRequest.EventRecorder,
				clusterLoggingRequest.ClusterID,
				utils.AsOwner(clusterLoggingRequest.Cluster)); err != nil {
				log.V(2).Error(err, "virtual logfilemetricexporter could not be created")
				r.Event(lfmeInstance, "Error", string(logging.ReasonInvalid), err.Error())
				telemetry.Data.LFMEInfo.Set(telemetry.HealthStatus, constants.UnHealthyStatus)
			}
		}
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
	if err := clusterRequest.removeCollector(); err != nil {
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
	for _, remove := range []func() error{
		func() error {
			return eslogstore.Remove(clusterRequest.Client, clusterRequest.Cluster.Namespace, clusterRequest.ResourceNames.InternalLogStoreSecret)
		},
		clusterRequest.removeKibana,
		func() error { return lokistack.RemoveRbac(clusterRequest.Client, clusterRequest.removeFinalizer) }} {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		if err := remove(); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "Error removing component")
		}
	}
}

func ReconcileForClusterLogForwarder(forwarder *logging.ClusterLogForwarder, clusterLogging *logging.ClusterLogging, requestClient client.Client, er record.EventRecorder, clusterID string, resourceNames *factory.ForwarderResourceNames) (err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Client:        requestClient,
		EventRecorder: er,
		Cluster:       clusterLogging,
		ClusterID:     clusterID,
	}

	if forwarder != nil {
		clusterLoggingRequest.Forwarder = forwarder
		clusterLoggingRequest.ResourceNames = resourceNames
	}

	// Owner will be CL instance if CLF is named instance
	if forwarder.Name == constants.SingletonName {
		clusterLoggingRequest.ResourceOwner = utils.AsOwner(clusterLogging)
	} else {
		clusterLoggingRequest.ResourceOwner = utils.AsOwner(forwarder)
	}

	if collectionSpec, err := CheckCollectionType(clusterLogging, forwarder); err != nil {
		return err
	} else {
		clusterLoggingRequest.CollectionSpec = collectionSpec
	}

	clusterLoggingRequest.Cluster = clusterLogging

	if clusterLogging != nil && clusterLogging.Spec.ManagementState == logging.ManagementStateUnmanaged {
		telemetry.Data.CLInfo.Set("managedStatus", constants.UnManagedStatus)
		return nil
	}

	// Rejected if clf condition is not ready
	// Do not create or update the collection
	if clusterLogging.Status.Conditions.IsFalseFor(logging.ConditionReady) {
		telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
		return nil
	}

	// If valid, generate the appropriate config
	if err = clusterLoggingRequest.CreateOrUpdateCollection(); err != nil {
		msg := fmt.Sprintf("Unable to reconcile collection for %q/%q: %v", clusterLoggingRequest.Forwarder.Namespace, clusterLoggingRequest.Forwarder.Name, err)
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

func CheckCollectionType(cl *logging.ClusterLogging, forwarder *logging.ClusterLogForwarder) (*logging.CollectionSpec, error) {

	var collectionSpec *logging.CollectionSpec

	// Forwarder named instance must have ClusterLogging instance specified
	if forwarder.Name == constants.SingletonName {
		if cl == nil || cl.Spec.Collection == nil {
			log.V(2).Info("skipping collection config generation as 'collection' section is not specified in CLO's CR")
			return nil, fmt.Errorf("'collection' section is not specified in the ClusterLogging resource")
		}

		// Check if collection type is valid
		switch cl.Spec.Collection.Type {
		case logging.LogCollectionTypeFluentd:
			break
		case logging.LogCollectionTypeVector:
			break
		default:
			return nil, fmt.Errorf("%s collector does not support pipelines feature", cl.Spec.Collection.Type)
		}

		collectionSpec = cl.Spec.Collection
	} else {
		// Make a default collection spec with just the type
		collectionSpec = &logging.CollectionSpec{
			Type: logging.LogCollectionTypeVector,
		}
	}

	return collectionSpec, nil
}

func ReconcileForLogFileMetricExporter(clusterLogging logging.ClusterLogging,
	lfmeInstance *loggingv1alpha1.LogFileMetricExporter,
	requestClient client.Client,
	er record.EventRecorder,
	clusterID string,
	owner metav1.OwnerReference) (err error) {

	// Make the daemonset along with metric services for Log file metric exporter
	if err := logmetricexporter.Reconcile(lfmeInstance, requestClient, er, clusterLogging, owner); err != nil {
		return err
	}
	// Successfully reconciled a LFME, set telemetry to deployed
	telemetry.SetLFMEMetrics(1)
	return nil
}

func getLogFileMetricExporter(reqClient client.Client) (lfmeInstance *loggingv1alpha1.LogFileMetricExporter, err error) {
	nsname := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.WatchNamespace}
	logFileMetricExporter := &loggingv1alpha1.LogFileMetricExporter{}
	if err = reqClient.Get(context.TODO(), nsname, logFileMetricExporter); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, errors.New("error getting logFileMetricExporter")
		}
		return nil, nil
	}

	return logFileMetricExporter, nil
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
	lgpipeline := request.Forwarder.Spec.Pipelines
	outputs := request.Forwarder.Spec.OutputMap()
	log.V(1).Info("OutputMap", "outputs", outputs)

	for _, pipeline := range lgpipeline {
		npipelines++
		log.V(1).Info("pipelines", "npipelines", npipelines)
		inref := pipeline.InputRefs
		outref := pipeline.OutputRefs

		telemetry.Data.CLFInputType.Range(func(labelname, value interface{}) bool {
			log.V(1).Info("iter over labelnames", "labelname", labelname)
			telemetry.Data.CLFInputType.Set(labelname.(string), constants.IsNotPresent) //reset to zero
			for _, inputtype := range inref {
				log.V(1).Info("iter over inputtype", "inputtype", inputtype)
				if inputtype == labelname {
					log.V(1).Info("labelname and inputtype", "labelname", labelname, "inputtype", inputtype) //when matched print matched labelname with input type stated in CLF spec
					telemetry.Data.CLFInputType.Set(labelname.(string), constants.IsPresent)                 //input type present in CLF spec
				}
			}
			return true // continue iterating
		})

		telemetry.Data.CLFOutputType.Range(func(labelname, value interface{}) bool {
			log.V(1).Info("iter over labelnames", "labelname", labelname)
			telemetry.Data.CLFOutputType.Set(labelname.(string), constants.IsNotPresent) //reset to zero
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
						telemetry.Data.CLFOutputType.Set(labelname.(string), constants.IsPresent) //when matched print matched labelname with output type stated in CLF spec
					}
				}
			}
			return true // continue iterating
		})
		log.V(1).Info("post updating inputtype and outputtype")
		telemetry.Data.CLFInfo.Set("pipelineInfo", strconv.Itoa(npipelines))
	}
}
