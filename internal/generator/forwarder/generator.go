package forwarder

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector"
	corev1 "k8s.io/api/core/v1"
	"net/url"
	"strings"
)

var (
	ErrNoOutputs        = errors.New("No outputs defined in ClusterLogForwarder")
	ErrNoValidInputs    = errors.New("No valid inputs found in ClusterLogForwarder")
	ErrInvalidOutputURL = func(o logging.OutputSpec) error {
		return fmt.Errorf("Invalid URL in %s output in ClusterLogForwarder", o.Name)
	}
	ErrInvalidInput = errors.New("Invalid Input")
)

type ConfigGenerator struct {
	g      generator.Generator
	conf   func(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Section
	format func(conf string) string
}

func New(collectorType logging.LogCollectionType) *ConfigGenerator {
	g := &ConfigGenerator{
		format: func(conf string) string { return conf },
	}
	switch collectorType {
	case logging.LogCollectionTypeFluentd:
		g.format = formatFluentConf
		g.conf = fluentd.Conf
	case logging.LogCollectionTypeVector:
		g.conf = vector.Conf
	default:
		log.Error(errors.New("Unsupported collector implementation"), "type", collectorType)
		return nil
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) (string, error) {
	sections := cg.conf(clspec, secrets, clfspec, op)
	conf, err := cg.g.GenerateConf(generator.MergeSections(sections)...)
	return cg.format(conf), err
}

func (cg *ConfigGenerator) Verify(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) error {
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
		if _, err := url.Parse(o.URL); err != nil {
			return ErrInvalidOutputURL(o)
		}
	}
	return err
}

func formatFluentConf(conf string) string {
	indent := 0
	lines := strings.Split(conf, "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(trimmed, "</") && strings.HasSuffix(trimmed, ">"):
			indent--
			trimmed = pad(trimmed, indent)
		case strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">"):
			trimmed = pad(trimmed, indent)
			indent++
		default:
			trimmed = pad(trimmed, indent)
		}
		if len(strings.TrimSpace(trimmed)) == 0 {
			trimmed = ""
		}
		lines[i] = trimmed
	}
	return strings.Join(lines, "\n")
}

func pad(line string, indent int) string {
	prefix := ""
	if indent >= 0 {
		prefix = strings.Repeat("  ", indent)
	}
	return prefix + line
}
