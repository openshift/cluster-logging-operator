package clusterlogforwarder

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var vectorLogLevelSet = sets.NewString("trace", "debug", "info", "warn", "error", "off")

func validateAnnotations(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	// No annotations to validate
	if len(clf.Annotations) == 0 {
		return nil, nil
	}

	// log level annotation
	if level, ok := clf.Annotations[constants.AnnotationVectorLogLevel]; ok {
		if !vectorLogLevelSet.Has(level) {
			return errors.NewValidationError("log level: %q is not valid. Must be one of trace, debug, info, warn, error, off.", level), nil
		}
	}

	return nil, nil
}
