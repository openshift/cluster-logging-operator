package forwarder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

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
	conf   func(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Section
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

func (cg *ConfigGenerator) GenerateConf(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) (string, error) {
	sections := cg.conf(clspec, secrets, clfspec, op)
	conf, err := cg.g.GenerateConf(generator.MergeSections(sections)...)
	return cg.format(conf), err
}

func (cg *ConfigGenerator) Verify(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) error {
	var err error
	types := generator.GatherSources(clfspec, op)
	if !types.HasAny(logging.InputNameApplication, logging.InputNameInfrastructure, logging.InputNameAudit) {
		return ErrNoValidInputs
	}
	if len(clfspec.Outputs) == 0 {
		return ErrNoOutputs
	}
	for _, p := range clfspec.Pipelines {
		if _, err := json.Marshal(p.Labels); err != nil {
			return ErrInvalidInput
		}
	}
	for _, o := range clfspec.Outputs {
		if _, err = url.Parse(o.URL); err != nil {
			return ErrInvalidOutputURL(o)
		}
		if o.GoogleCloudLogging != nil {
			gcl := o.GoogleCloudLogging
			i := 0
			if gcl.ProjectID != "" {
				i += 1
			}
			if gcl.FolderID != "" {
				i += 1
			}
			if gcl.BillingAccountID != "" {
				i += 1
			}
			if gcl.OrganizationID != "" {
				i += 1
			}
			if i > 1 {
				return ErrGCL
			}
		}
	}
	return err
}
