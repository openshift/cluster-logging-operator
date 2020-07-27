package v1alpha1

import (
  "testing"

  "github.com/google/go-cmp/cmp"
  loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConversionV1alpha1toV1(t *testing.T) {
  tests := []struct {
    desc    string
    src     *LogForwarding
    want    *loggingv1.ClusterLogForwarder
    wantErr bool
  }{
    {
      desc: "successful conversion",
      src: &LogForwarding{
        ObjectMeta: metav1.ObjectMeta{
          Name:      "instance",
          Namespace: "openshift-logging",
        },
        Spec: ForwardingSpec{
          Outputs: []OutputSpec{
            {
              Name:     "elasticsearch",
              Type:     OutputTypeElasticsearch,
              Endpoint: "https://elasticsearch.nirvana.svc.:9200",
            },
            {
              Name:     "fluentd-forward",
              Type:     OutputTypeForward,
              Endpoint: "tls://fluentforward.nirvana.svc.:8888",
              Secret: &OutputSecretSpec{
                Name: "secret-name",
              },
            },
            {
              Name:     "syslog",
              Type:     OutputTypeSyslog,
              Endpoint: "tcp://syslog.nirvana.svc.:8888",
            },
          },
          Pipelines: []PipelineSpec{
            {
              Name:       "application-logs",
              SourceType: LogSourceTypeApp,
              Namespaces: []string{"nirvana", "nowhere"},
              OutputRefs: []string{
                "elastichsearch",
                "fluentd-forward",
              },
            },
            {
              Name:       "audit-logs",
              SourceType: LogSourceTypeAudit,
              OutputRefs: []string{"syslog"},
            },
            {
              Name:       "infra-logs",
              SourceType: LogSourceTypeInfra,
              OutputRefs: []string{"syslog"},
            },
          },
        },
      },
      want: &loggingv1.ClusterLogForwarder{
        ObjectMeta: metav1.ObjectMeta{
          Name:      "instance",
          Namespace: "openshift-logging",
        },
        Spec: loggingv1.ClusterLogForwarderSpec{
          Inputs: []loggingv1.InputSpec{
            {
              Name: "application",
              Application: &loggingv1.Application{
                Namespaces: []string{"nirvana", "nowhere"},
              },
              Audit:          &loggingv1.Audit{},
              Infrastructure: &loggingv1.Infrastructure{},
            },
            {
              Name:           "audit",
              Application:    &loggingv1.Application{Namespaces: []string{}},
              Audit:          &loggingv1.Audit{},
              Infrastructure: &loggingv1.Infrastructure{},
            },
            {
              Name:           "infrastructure",
              Application:    &loggingv1.Application{Namespaces: []string{}},
              Audit:          &loggingv1.Audit{},
              Infrastructure: &loggingv1.Infrastructure{},
            },
          },
          Outputs: []loggingv1.OutputSpec{
            {
              Name: "elasticsearch",
              Type: loggingv1.OutputTypeElasticsearch,
              URL:  "https://elasticsearch.nirvana.svc.:9200",
            },
            {
              Name: "fluentd-forward",
              Type: loggingv1.OutputTypeFluentdForward,
              URL:  "tls://fluentforward.nirvana.svc.:8888",
              Secret: &loggingv1.OutputSecretSpec{
                Name: "secret-name",
              },
            },
            {
              Name: "syslog",
              Type: loggingv1.OutputTypeSyslog,
              URL:  "tcp://syslog.nirvana.svc.:8888",
              OutputTypeSpec: loggingv1.OutputTypeSpec{
                Syslog: &loggingv1.Syslog{RFC: "RFC3164"},
              },
            },
          },
          Pipelines: []loggingv1.PipelineSpec{
            {
              Name:       "application-logs",
              InputRefs:  []string{"application"},
              OutputRefs: []string{"elastichsearch", "fluentd-forward"},
            },
            {
              Name:       "audit-logs",
              InputRefs:  []string{"audit"},
              OutputRefs: []string{"syslog"},
            },
            {
              Name:       "infra-logs",
              InputRefs:  []string{"infrastructure"},
              OutputRefs: []string{"syslog"},
            },
          },
        },
      },
    },
    {
      desc: "unsupported log source type",
      src: &LogForwarding{
        Spec: ForwardingSpec{
          Pipelines: []PipelineSpec{
            {
              SourceType: "not-supported",
            },
          },
        },
      },
      want:    &loggingv1.ClusterLogForwarder{},
      wantErr: true,
    },
    {
      desc: "unsupported output type",
      src: &LogForwarding{
        Spec: ForwardingSpec{
          Outputs: []OutputSpec{
            {
              Type: "not-supported",
            },
          },
        },
      },
      want:    &loggingv1.ClusterLogForwarder{},
      wantErr: true,
    },
    {
      desc: "invalid endpoint",
      src: &LogForwarding{
        Spec: ForwardingSpec{
          Outputs: []OutputSpec{
            {
              Type:     "elasticsearch",
              Endpoint: "%%whatever:1234",
            },
          },
        },
      },
      want:    &loggingv1.ClusterLogForwarder{},
      wantErr: true,
    },
  }
  for _, tc := range tests {
    tc := tc
    t.Run(tc.desc, func(t *testing.T) {
      t.Parallel()

      dst := &loggingv1.ClusterLogForwarder{}
      err := tc.src.ConvertTo(dst)
      if err != nil && !tc.wantErr {
        t.Fatalf("no err provided: %s", err)
      }

      if diff := cmp.Diff(dst, tc.want); diff != "" {
        t.Errorf("got diff: %s", diff)
      }
    })
  }
}
