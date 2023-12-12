package k8shandler

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"

	"github.com/openshift/cluster-logging-operator/internal/tls"

	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
)

// EvaluateAnnotationsForEnabledCapabilities populates generator options with capabilities enabled by the ClusterLogForwarder
func EvaluateAnnotationsForEnabledCapabilities(forwarder *logging.ClusterLogForwarder, options framework.Options) {
	if forwarder == nil {
		return
	}
	for key, value := range forwarder.Annotations {
		switch key {
		case constants.UseOldRemoteSyslogPlugin:
			if strings.ToLower(value) == constants.Enabled {
				options[key] = ""
			}
		case constants.AnnotationDebugOutput:
			if strings.ToLower(value) == "true" {
				options[helpers.EnableDebugOutput] = "true"
			}

		case constants.AnnotationEnableSchema:
			if strings.ToLower(value) == constants.Enabled {
				options[constants.AnnotationEnableSchema] = "true"
			}
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {

	op := framework.Options{}
	tlsProfile, _ := tls.FetchAPIServerTlsProfile(clusterRequest.Client)
	op[framework.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(tlsProfile)
	EvaluateAnnotationsForEnabledCapabilities(clusterRequest.Forwarder, op)
	g := forwardergenerator.New(clusterRequest.Cluster.Spec.Collection.Type)
	generatedConfig, err := g.GenerateConf(clusterRequest.Cluster.Spec.Collection, clusterRequest.OutputSecrets, &clusterRequest.Forwarder.Spec, clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.ResourceNames, op)

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

	// Use logcollector SA token/ca.crt for the legacy case
	if clusterRequest.Forwarder.Spec.ServiceAccountName == constants.CollectorServiceAccountName {
		tokenSecret, err := clusterRequest.GetSecret(constants.LogCollectorToken)
		if err == nil {
			clusterRequest.OutputSecrets[constants.LogCollectorToken] = tokenSecret
		}
	}
}
