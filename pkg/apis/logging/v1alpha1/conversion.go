package v1alpha1

import (
  "fmt"
  "net/url"
  "regexp"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

  loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var matchURL = regexp.MustCompile("^[a-zA-Z]+://")

// makeValidURL converts s, which can be a URL or a host:port string, into a URL.
// If s is a host:port string, scheme is used as the URL scheme.
func makeValidURL(s, scheme string) (*url.URL, error) {
  if !matchURL.MatchString(s) {
    s = fmt.Sprintf("%v://%v", scheme, s)
  }
  return url.Parse(s)
}

func (src *LogForwarding) ConvertTo(dst *loggingv1.ClusterLogForwarder) error {
  spec, err := convertSpec(&src.Spec)
  if err != nil {
    return fmt.Errorf("failed converting LF spec to CLF spec: %s", err)
  }

  dst.ObjectMeta = metav1.ObjectMeta{
    Name:      src.Name,
    Namespace: src.Namespace,
  }
  dst.Spec = *spec
  dst.Status = loggingv1.ClusterLogForwarderStatus{}
  return nil
}

func convertSpec(src *ForwardingSpec) (*loggingv1.ClusterLogForwarderSpec, error) {
  inputs, err := convertInputs(src.Pipelines)
  if err != nil {
    return nil, fmt.Errorf("failed converting LF inputs to CLF inputs: %s", err)
  }

  outputs, err := convertOutputs(src.Outputs)
  if err != nil {
    return nil, fmt.Errorf("failed converting LF outputs to CLF outputs: %s", err)
  }

  pipelines, err := convertPipelines(src.Pipelines)
  if err != nil {
    return nil, fmt.Errorf("failed converting LF pipelines to CLF pipelines: %s", err)
  }

  dst := &loggingv1.ClusterLogForwarderSpec{
    Inputs:    inputs,
    Outputs:   outputs,
    Pipelines: pipelines,
  }

  return dst, nil
}

func convertInputs(src []PipelineSpec) ([]loggingv1.InputSpec, error) {
  dst := []loggingv1.InputSpec{}

  for _, pipeline := range src {
    name, err := convertLogSourceType(pipeline.SourceType)
    if err != nil {
      return nil, err
    }

    input := loggingv1.InputSpec{
      Name:           name,
      Application:    &loggingv1.Application{Namespaces: []string{}},
      Audit:          &loggingv1.Audit{},
      Infrastructure: &loggingv1.Infrastructure{},
    }

    if len(pipeline.Namespaces) > 0 {
      input.Application.Namespaces = append(input.Application.Namespaces, pipeline.Namespaces...)
    }

    dst = append(dst, input)
  }

  return dst, nil
}

func convertOutputs(src []OutputSpec) ([]loggingv1.OutputSpec, error) {
  dst := []loggingv1.OutputSpec{}

  for _, srcOut := range src {
    var (
      dstSpec   loggingv1.OutputTypeSpec
      dstType   string
      srcScheme string
    )

    switch srcOut.Type {
    case OutputTypeElasticsearch:
      dstType = loggingv1.OutputTypeElasticsearch
      dstSpec = loggingv1.OutputTypeSpec{}
      srcScheme = "https"

    case OutputTypeForward:
      dstType = loggingv1.OutputTypeFluentdForward
      dstSpec = loggingv1.OutputTypeSpec{}
      srcScheme = "tcp"
      if srcOut.Secret != nil {
        srcScheme = "tls"
      }

    case OutputTypeSyslog:
      dstType = loggingv1.OutputTypeSyslog
      dstSpec = loggingv1.OutputTypeSpec{
        Syslog: &loggingv1.Syslog{
          RFC: "RFC3164",
        },
      }

      srcScheme = "tcp"
      if srcOut.Secret != nil {
        srcScheme = "tls"
      }

    default:
      return nil, fmt.Errorf("failed to convert output %s, unrecognized output type %s", srcOut.Name, srcOut.Type)
    }

    url, err := makeValidURL(srcOut.Endpoint, srcScheme)
    if err != nil {
      return nil, fmt.Errorf("failed to convert output's %q endpoint to valid URL: %s", srcOut.Name, err)
    }

    dstOut := loggingv1.OutputSpec{
      Name:           srcOut.Name,
      Type:           dstType,
      URL:            url.String(),
      OutputTypeSpec: dstSpec,
    }

    if srcOut.Secret != nil {
      dstOut.Secret = &loggingv1.OutputSecretSpec{
        Name: srcOut.Secret.Name,
      }
    }

    dst = append(dst, dstOut)
  }

  return dst, nil
}

func convertPipelines(src []PipelineSpec) ([]loggingv1.PipelineSpec, error) {
  dst := []loggingv1.PipelineSpec{}

  for _, srcPipeline := range src {
    name, err := convertLogSourceType(srcPipeline.SourceType)
    if err != nil {
      return nil, fmt.Errorf("failed converting LF pipeline to CLF pipeline: %s", err)
    }

    dstPipeline := loggingv1.PipelineSpec{
      Name:       srcPipeline.Name,
      InputRefs:  []string{name},
      OutputRefs: srcPipeline.OutputRefs,
    }

    dst = append(dst, dstPipeline)
  }

  return dst, nil
}

func convertLogSourceType(name LogSourceType) (string, error) {
  switch name {
  case LogSourceTypeApp:
    return loggingv1.InputNameApplication, nil

  case LogSourceTypeAudit:
    return loggingv1.InputNameAudit, nil

  case LogSourceTypeInfra:
    return loggingv1.InputNameInfrastructure, nil

  default:
    return "", fmt.Errorf("failed converting LF log source type %q to CLF input type", name)
  }
}
