package forwarder

import (
	"errors"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/conf"

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
	conf   func(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) []framework.Section
	format func(conf string) string
}

func New() *ConfigGenerator {
	g := &ConfigGenerator{
		format: helpers.FormatVectorToml,
		conf:   conf.Conf,
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) (string, error) {
	sections := cg.conf(secrets, clfspec, namespace, forwarderName, resNames, op)
	conf, err := cg.g.GenerateConf(framework.MergeSections(sections)...)
	return cg.format(conf), err
}
