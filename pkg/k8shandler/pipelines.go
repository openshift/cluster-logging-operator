package k8shandler

// import (
// 	"fmt"

// 	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
// 	"github.com/openshift/cluster-logging-operator/pkg/utils"
// 	"k8s.io/apimachinery/pkg/api/errors"
// 	core "k8s.io/api/core/v1"

// 	yaml "gopkg.in/yaml.v2"
// )

// const (
// 	PipelineInputConfigMapName = "collector-pipeline-input"
// )

// func (cluster *ClusterLogging) createOrUpdatePipelines() (err error) {

// 	if cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeRsyslog {
// 		return fmt.Errorf("Rsyslog collector does not support pipelines feature")
// 	}
// 	cluster.updatePipelineContext()

// 	return cluster.createOrUpdatePipelineInputConfigMap()
// }

// func (cluster *ClusterLogging) updatePipelineContext () {

// 	cluster.Context.Pipelines = cluster.Spec.Pipelines
// 	//if logstore is not defined then pipeline is whatever is defined
// 	//in the ClusterLogging resource.
// 	if cluster.Spec.LogStore.Type != logging.LogStoreTypeElasticsearch {
// 		return
// 	}
// 	if cluster.Context.Pipelines == nil {
// 		cluster.Context.Pipelines = &PipelinesSpec{}
// 	}
// 	// defaultPipelineTarget := &PipelineTargetsSpec{
// 	// 	Targets: []PipelineTargetSpec{
// 	// 		{	PipelineTargetSpec{
// 	// 				Type: PipelineTargetTypeElasticsearch,
// 	// 				Endpoint: fmt.Sprintf("%s.%s.svc:9200", elasticsearchResourceName, cluster.Namespace),
// 	// 				Certificates: PipelineTargetCertificatesSpec {
// 	// 					SecretName: "",
// 	// 				}
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }
// 	// if cluster.Context.Pipelines.LogsApp == nil {
// 	// 	cluster.Context.Pipelines.LogsApp = defaultPipelineTarget
// 	// 	}
// 	// }
// 	// if cluster.Context.Pipelines.LogsInfra == nil {
// 	// 	cluster.Context.Pipelines.LogsInfra = defaultPipelineTarget
// 	// 	}
// 	// }
// }

// func (cluster *ClusterLogging) createOrUpdatePipelineInputConfigMap() (err error) {

// 	pipelines := cluster.ClusterLogging.Spec.Pipelines
// 	if conf, err = yaml.Marshal(pipelines); err != nil {
// 		return err
// 	}

// 	configMap := utils.NewConfigMap(
// 		PipelineInputConfigMapName,
// 		cluster.Namespace,
// 		map[string]string{
// 			"pipelines": string(conf),
// 		},
// 	)
// 	cluster.AddOwnerRefTo(configMap)

// 	current := pipelines.DeepCopy()
// 	runtime := cluster.Runtime
// 	if err = runtime.Get(current); err != nil {
// 		if errors.IsNotFound(err) && err = runtime.Create(configMap); err != nil {
// 				return fmt.Errorf("Failure creating %s configmap: %v", configMap.Name, err)
// 			}
// 		}
// 	}
// 	update := func() error {
// 		return runtime.Update(configMap)
// 	}
// 	if retryErr := runtime.RetryOnConflict(retry.DefaultRetry, update); retryErr != nil {
// 		return fmt.Errorf("Failed to update status for ClusterLogging %q: %v", configMap.Name, retryErr)
// 	}
// 	return nil
// }

// func (cluster *ClusterLogging) createOrUpdatePiplineInputVolumeMounts(current []v1.Volume) []v1.Volume {
// 	pipelines := cluster.ClusterLogging.Spec.Pipelines
// 	results := [] v1.Volume {}
// 	for _, volume := range current {
// 		if volume.Name != PipelineInputConfigMapName {
// 			append(results, volume)
// 		} else {

// 		}
// 	}
// 	return results
// }
