package telemetry

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func UpdateInfofromCL(cl logging.ClusterLogging) {
	clspec := cl.Spec
	if clspec.LogStore != nil && clspec.LogStore.Type != "" {
		log.V(1).Info("LogStore Type", "clspecLogStoreType", clspec.LogStore.Type)
		Data.CLLogStoreType.Set(string(clspec.LogStore.Type), constants.IsPresent)
	}
}
