package forwarder

import (
	"errors"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector"
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
	g      generator.Generator
	conf   func(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace string, op generator.Options) []generator.Section
	format func(conf string) string
}

func New(collectorType logging.LogCollectionType) *ConfigGenerator {
	g := &ConfigGenerator{
		format: func(conf string) string { return conf },
	}
	switch collectorType {
	case logging.LogCollectionTypeFluentd:
		g.format = helpers.FormatFluentConf
		g.conf = fluentd.Conf
	case logging.LogCollectionTypeVector:
		g.conf = vector.Conf
	default:
		log.Error(errors.New("Unsupported collector implementation"), "type", collectorType)
		return nil
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace string, op generator.Options) (string, error) {
	sections := cg.conf(clspec, secrets, clfspec, namespace, op)
	conf, err := cg.g.GenerateConf(generator.MergeSections(sections)...)
	return cg.format(conf), err
}
