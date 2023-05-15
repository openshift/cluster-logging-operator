package k8shandler

import (
	"errors"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"strings"

	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"

	corev1 "k8s.io/api/core/v1"
)

// EvaluateAnnotationsForEnabledCapabilities populates generator options with capabilities enabled by the ClusterLogForwarder
func EvaluateAnnotationsForEnabledCapabilities(forwarder *logging.ClusterLogForwarder, options generator.Options) {
	if forwarder == nil {
		return
	}
	for key, value := range forwarder.Annotations {
		switch key {
		case constants.PreviewTLSSecurityProfile:
			fallthrough
		case constants.UseOldRemoteSyslogPlugin:
			if strings.ToLower(value) == constants.Enabled {
				options[key] = ""
			}
		case constants.AnnotationDebugOutput:
			if strings.ToLower(value) == "true" {
				options[helpers.EnableDebugOutput] = "true"
			}
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {
	if clusterRequest.Cluster == nil || clusterRequest.Cluster.Spec.Collection == nil {
		log.V(2).Info("skipping collection config generation as 'collection' section is not specified in CLO's CR")
		return "", nil
	}
	switch clusterRequest.Cluster.Spec.Collection.Type {
	case logging.LogCollectionTypeFluentd:
		break
	case logging.LogCollectionTypeVector:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.Cluster.Spec.Collection.Type)
	}

	if clusterRequest.Forwarder == nil {
		clusterRequest.Forwarder = &logging.ClusterLogForwarder{}
	}

	// Set the output secrets if any
	clusterRequest.SetOutputSecrets()

	tokenSecret, err := clusterRequest.getLogCollectorServiceAccountTokenSecret()
	if err == nil {
		clusterRequest.OutputSecrets[constants.LogCollectorToken] = tokenSecret
	}

	op := generator.Options{}
	tlsProfile, _ := tls.FetchAPIServerTlsProfile(clusterRequest.Client)
	op[generator.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(tlsProfile)
	EvaluateAnnotationsForEnabledCapabilities(clusterRequest.Forwarder, op)

	var collectorType = clusterRequest.Cluster.Spec.Collection.Type
	g := forwardergenerator.New(collectorType)
	generatedConfig, err := g.GenerateConf(clusterRequest.Cluster.Spec.Collection, clusterRequest.OutputSecrets, &clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Namespace, op)

	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		if updateError := clusterRequest.UpdateCondition(
			logging.CollectorDeadEnd,
			"Collectors are defined but there is no defined LogStore or LogForward destinations",
			"No defined logstore destination",
			corev1.ConditionTrue,
		); updateError != nil {
			log.Error(updateError, "Unable to update the clusterlogging status", "conditionType", logging.CollectorDeadEnd)
		}
		return "", err
	}
	// else
	err = clusterRequest.UpdateCondition(
		logging.CollectorDeadEnd,
		"",
		"",
		corev1.ConditionFalse,
	)
	log.V(3).Info("ClusterLogForwarder generated config", generatedConfig)
	return generatedConfig, err
}

func (clusterRequest *ClusterLoggingRequest) SetOutputSecrets() {
	clusterRequest.OutputSecrets = make(map[string]*corev1.Secret, len(clusterRequest.Forwarder.Spec.Outputs))
	for _, output := range clusterRequest.Forwarder.Spec.Outputs {
		output := output // Don't bind range variable.
		if output.Secret == nil {
			continue
		}
		// Should be able to get secret because output has been validated
		secret, _ := clusterRequest.GetSecret(output.Secret.Name)
		clusterRequest.OutputSecrets[output.Name] = secret
	}
}

func (clusterRequest *ClusterLoggingRequest) getLogCollectorServiceAccountTokenSecret() (*corev1.Secret, error) {
	s := &corev1.Secret{}
	log.V(9).Info("Fetching Secret", "Name", constants.LogCollectorToken)
	if err := clusterRequest.Get(constants.LogCollectorToken, s); err != nil {
		log.V(3).Error(err, "Could not find Secret", "Name", constants.LogCollectorToken)
		return nil, errors.New("Could not retrieve ServiceAccount token")
	}

	if _, ok := s.Data[constants.TokenKey]; !ok {
		log.V(9).Info("did not find token in secret", "Name", s.Name)
		return nil, errors.New("logcollector secret is missing token")
	}

	return s, nil
}
