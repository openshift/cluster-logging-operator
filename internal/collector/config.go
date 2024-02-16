package collector

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileCollectorConfig reconciles a collector config specifically for the collector defined by the factory
func (f *Factory) ReconcileCollectorConfig(er record.EventRecorder, k8sClient client.Client, reader client.Reader, namespace, collectorConfig string, owner metav1.OwnerReference) error {
	log.V(3).Info("Updating ConfigMap and Secrets")
	if f.CollectorType == logging.LogCollectionTypeFluentd {
		collectorConfigMap := runtime.NewConfigMap(
			namespace,
			f.ResourceNames.ConfigMap,
			map[string]string{
				"fluent.conf":         collectorConfig,
				"run.sh":              fluentd.RunScript,
				"cleanInValidJson.rb": fluentd.CleanInValidJson,
			},
			f.CommonLabelInitializer,
		)
		utils.AddOwnerRefToObject(collectorConfigMap, owner)
		return reconcile.Configmap(k8sClient, reader, collectorConfigMap)
	} else if f.CollectorType == logging.LogCollectionTypeVector {

		secret := runtime.NewSecret(
			namespace,
			f.ResourceNames.ConfigMap,
			map[string][]byte{
				vector.ConfigFile:    []byte(collectorConfig),
				vector.RunVectorFile: []byte(fmt.Sprintf(vector.RunVectorScript, vector.GetDataPath(namespace, f.ResourceNames.ForwarderName))),
			},
			f.CommonLabelInitializer)

		utils.AddOwnerRefToObject(secret, owner)
		return reconcile.Secret(er, k8sClient, secret, comparators.CompareLabels)
	}

	return nil
}
