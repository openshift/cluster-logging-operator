package k8shandler

import (
	"errors"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/tls"

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

	op := generator.Options{}
	tlsProfile, _ := tls.FetchAPIServerTlsProfile(clusterRequest.Client)
	op[generator.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(tlsProfile)
	EvaluateAnnotationsForEnabledCapabilities(clusterRequest.Forwarder, op)

	g := forwardergenerator.New(clusterRequest.CollectionSpec.Type)
	generatedConfig, err := g.GenerateConf(clusterRequest.CollectionSpec, clusterRequest.OutputSecrets, &clusterRequest.Forwarder.Spec, clusterRequest.Forwarder.Namespace, op)

	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		return "", err
	}

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

func (clusterRequest *ClusterLoggingRequest) GetLogCollectorServiceAccountTokenSecret() (*corev1.Secret, error) {
	colTokenName := clusterRequest.ResourceNames.ServiceAccountTokenSecret
	s := &corev1.Secret{}
	log.V(9).Info("Fetching Secret", "Name", colTokenName)
	if err := clusterRequest.Get(colTokenName, s); err != nil {
		log.V(3).Error(err, "Could not find Secret", "Name", colTokenName)
		return nil, errors.New("Could not retrieve ServiceAccount token")
	}

	if _, ok := s.Data[constants.TokenKey]; !ok {
		log.V(9).Info("did not find token in secret", "Name", s.Name)
		return nil, errors.New(colTokenName + " secret is missing token")
	}

	return s, nil
}
