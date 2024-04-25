package forwarder

import (
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/conf"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrNoOutputs        = errors.New("No outputs defined in ClusterLogForwarder")
	ErrNoValidInputs    = errors.New("No valid inputs found in ClusterLogForwarder")
	ErrInvalidOutputURL = func(o logging.OutputSpec) error {
		return fmt.Errorf("Invalid URL in %s output in ClusterLogForwarder", o.Name)
	}
	ErrInvalidInput      = errors.New("Invalid Input")
	ErrTLSOutputNoSecret = func(o logging.OutputSpec) error {
		return fmt.Errorf("No secret defined in output %s, but URL has TLS Scheme %s", o.Name, o.URL)
	}
	ErrGCL = errors.New("Exactly one of billingAccountId, folderId, organizationId, or projectId must be set.")
)

type ConfigGenerator struct {
	g      framework.Generator
	conf   func(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace, forwarderName string, resNames *factory.ForwarderResourceNames, op framework.Options) []framework.Section
	format func(conf string) string
}

func New(collectorType logging.LogCollectionType) *ConfigGenerator {
	g := &ConfigGenerator{
		format: func(conf string) string { return conf },
	}
	switch collectorType {
	case logging.LogCollectionTypeVector:
		g.format = helpers.FormatVectorToml
		g.conf = conf.Conf
	default:
		log.Error(errors.New("Unsupported collector implementation"), "type", collectorType)
		return nil
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace, forwarderName string, resNames *factory.ForwarderResourceNames, op framework.Options) (string, error) {
	sections := cg.conf(clspec, secrets, clfspec, namespace, forwarderName, resNames, op)
	conf, err := cg.g.GenerateConf(framework.MergeSections(sections)...)
	return cg.format(conf), err
}
