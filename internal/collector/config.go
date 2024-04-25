package collector

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
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
