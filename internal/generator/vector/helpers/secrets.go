package helpers

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	corev1 "k8s.io/api/core/v1"
)

func GetOutputSecret(o logging.OutputSpec, secrets map[string]*corev1.Secret) *corev1.Secret {
	if s, ok := secrets[o.Name]; ok {
		log.V(9).Info("Using secret configured in output: " + o.Name)
		return s
	}
	if o.Type == logging.OutputTypeLoki && lokistack.DefaultLokiOutputNames.Has(o.Name) {
		log.V(9).Info("Default lokiStack, using collector token", "output", o.Name, "secret", constants.LogCollectorToken)
		return secrets[constants.LogCollectorToken]
	}
	log.V(9).Info("No Secret found for output", "output", o.Name)
	return nil
}
