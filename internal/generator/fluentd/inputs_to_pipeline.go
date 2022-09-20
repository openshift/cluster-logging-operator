package fluentd

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
)

type ApplicationToPipeline struct {
	// Labels is an array of "<key>:<value>" strings
	Labels     []string
	Namespaces []string
	Pipeline   string
}

type ApplicationsToPipelines []ApplicationToPipeline

func (a ApplicationsToPipelines) Name() string {
	return "applicationToPipeline"
}

func (a ApplicationsToPipelines) Template() string {
	return `{{define "` + a.Name() + `"  -}}
# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
	{{- range $index, $a := .}}
    <route>
      @label {{$a.Pipeline}}
      <match>
        {{- if $a.Namespaces}}
        namespaces {{comma_separated $a.Namespaces}}
		{{- end}}
        {{- if $a.Labels}}
        labels {{comma_separated $a.Labels }}
		{{- end}}
      </match>
    </route>
	{{- end}}
  </match>
</label>
{{- end}}`
}

func SourceTypeToPipeline(sourceType string, spec *logging.ClusterLogForwarderSpec, op Options) Element {
	srcTypePipeline := []string{}
	for _, pipeline := range spec.Pipelines {
		for _, inRef := range pipeline.InputRefs {
			if inRef == sourceType {
				srcTypePipeline = append(srcTypePipeline, pipeline.Name)
			}
		}
	}
	switch len(srcTypePipeline) {
	case 0:
		return Nil
	case 1:
		return FromLabel{
			Desc:    fmt.Sprintf("Sending %s source type to pipeline", sourceType),
			InLabel: helpers.SourceTypeLabelName(sourceType),
			SubElements: []Element{
				AddLogSourceType(sourceType),
				Match{
					MatchTags: "**",
					MatchElement: Relabel{
						OutLabel: helpers.LabelName(srcTypePipeline[0]),
					},
				},
			},
		}
	default:
		return FromLabel{
			Desc:    fmt.Sprintf("Copying %s source type to pipeline", sourceType),
			InLabel: helpers.SourceTypeLabelName(sourceType),
			SubElements: []Element{
				AddLogSourceType(sourceType),
				Match{
					MatchTags: "**",
					MatchElement: Copy{
						Stores: CopyToLabels(helpers.LabelNames(srcTypePipeline)),
					},
				},
			},
		}
	}
}

func InputsToPipeline(spec *logging.ClusterLogForwarderSpec, op Options) []Element {
	return MergeElements(
		AppToPipeline1(spec, op),
		InfraToPipeline(spec, op),
		AuditToPipeline(spec, op),
	)
}

func AppToPipeline1(spec *logging.ClusterLogForwarderSpec, op Options) []Element {
	userDefined := spec.InputMap()
	// routed by namespace, or labels
	routes := []Element{}
	unRoutedPipelines := []string{}
	for _, pipeline := range spec.Pipelines {
		for _, inRef := range pipeline.InputRefs {
			if input, ok := userDefined[inRef]; ok {
				// user defined input
				if input.Application != nil {
					app := input.Application
					if len(app.Namespaces) != 0 || (app.Selector != nil && len(app.Selector.MatchLabels) != 0) {
						rd := RouteData{}
						if len(app.Namespaces) != 0 {
							rd.Namespaces = KV("namespaces", strings.Join(app.Namespaces, ", "))
						}
						if app.Selector != nil && len(app.Selector.MatchLabels) != 0 {
							selector := &logging.LabelSelector{MatchLabels: app.Selector.MatchLabels}
							rd.Labels = KV("labels", strings.Join(helpers.LabelsKV(selector), ", "))
						}
						routes = append(routes, Route{
							RoutePipeline: RoutePipeline{
								Pipeline:  helpers.LabelName(pipeline.Name),
								RouteData: rd,
							},
						})
					} else {
						unRoutedPipelines = append(unRoutedPipelines, pipeline.Name)
					}
				}
			} else if inRef == logging.InputNameApplication {
				unRoutedPipelines = append(unRoutedPipelines, pipeline.Name)
			}
		}
	}
	// No pipelines need Label/Namespace based routing
	if len(routes) == 0 {
		return []Element{
			SourceTypeToPipeline(logging.InputNameApplication, spec, op),
		}
	}

	switch len(unRoutedPipelines) {
	case 0:
		return []Element{
			FromLabel{
				Desc:    "Routing Application to pipelines",
				InLabel: helpers.SourceTypeLabelName(logging.InputNameApplication),
				SubElements: []Element{
					AddLogSourceType(logging.InputNameApplication),
					Match{
						MatchTags: "**",
						MatchElement: LabelRouter{
							Routes: routes,
						},
					},
				},
			},
		}
	case 1:
		routes = append(routes, Route{
			RoutePipeline: RoutePipeline{
				Pipeline: helpers.SourceTypeLabelName("APPLICATION_ALL"),
			},
		})
		return []Element{
			FromLabel{
				Desc:    "Routing Application to pipelines",
				InLabel: helpers.SourceTypeLabelName(logging.InputNameApplication),
				SubElements: []Element{
					AddLogSourceType(logging.InputNameApplication),
					Match{
						MatchTags: "**",
						MatchElement: LabelRouter{
							Routes: routes,
						},
					},
				},
			},
			FromLabel{
				Desc:    "Sending unrouted application to pipelines",
				InLabel: helpers.SourceTypeLabelName("APPLICATION_ALL"),
				SubElements: []Element{
					Match{
						MatchTags: "**",
						MatchElement: Relabel{
							OutLabel: helpers.LabelName(unRoutedPipelines[0]),
						},
					},
				},
			},
		}
	default:
		routes = append(routes, Route{
			RoutePipeline: RoutePipeline{
				Pipeline: helpers.SourceTypeLabelName("APPLICATION_ALL"),
			},
		})
		return []Element{
			FromLabel{
				Desc:    "Routing Application to pipelines",
				InLabel: helpers.SourceTypeLabelName(logging.InputNameApplication),
				SubElements: []Element{
					AddLogSourceType(logging.InputNameApplication),
					Match{
						MatchTags: "**",
						MatchElement: LabelRouter{
							Routes: routes,
						},
					},
				},
			},
			FromLabel{
				Desc:    "Copying unrouted application to pipelines",
				InLabel: helpers.SourceTypeLabelName("APPLICATION_ALL"),
				SubElements: []Element{
					Match{
						MatchTags: "**",
						MatchElement: Copy{
							Stores: CopyToLabels(helpers.LabelNames(unRoutedPipelines)),
						},
					},
				},
			},
		}
	}
}

func AppToPipeline(spec *logging.ClusterLogForwarderSpec, op Options) []Element {
	userDefined := spec.InputMap()
	// routed by namespace, or labels
	routedPipelines := ApplicationsToPipelines{}
	unRoutedPipelines := []string{}
	for _, pipeline := range spec.Pipelines {
		for _, inRef := range pipeline.InputRefs {
			if input, ok := userDefined[inRef]; ok {
				// user defined input
				if input.Application != nil {
					app := input.Application
					var a *ApplicationToPipeline = nil
					if len(app.Namespaces) != 0 {
						a = &ApplicationToPipeline{
							Pipeline: helpers.LabelName(pipeline.Name),
						}
						a.Namespaces = app.Namespaces
					}
					if app.Selector != nil && len(app.Selector.MatchLabels) != 0 {
						if a == nil {
							a = &ApplicationToPipeline{
								Pipeline: helpers.LabelName(pipeline.Name),
							}
						}
						a.Labels = helpers.LabelsKV(app.Selector)
					}
					if a != nil {
						routedPipelines = append(routedPipelines, *a)
					} else {
						unRoutedPipelines = append(unRoutedPipelines, pipeline.Name)
					}
				}
			} else if inRef == logging.InputNameApplication {
				unRoutedPipelines = append(unRoutedPipelines, pipeline.Name)
			}
		}
	}
	if len(routedPipelines) == 0 {
		return []Element{
			SourceTypeToPipeline(logging.InputNameApplication, spec, op),
		}
	}

	switch len(unRoutedPipelines) {
	case 0:
		return []Element{
			routedPipelines,
		}
	case 1:
		routedPipelines = append(routedPipelines, ApplicationToPipeline{
			Pipeline: helpers.SourceTypeLabelName("APPLICATION_ALL"),
		})
		return []Element{
			routedPipelines,
			FromLabel{
				Desc:    "Sending unrouted application to pipelines",
				InLabel: helpers.SourceTypeLabelName("APPLICATION_ALL"),
				SubElements: []Element{
					Match{
						MatchTags: "**",
						MatchElement: Relabel{
							OutLabel: helpers.LabelName(unRoutedPipelines[0]),
						},
					},
				},
			},
		}
	default:
		routedPipelines = append(routedPipelines, ApplicationToPipeline{
			Pipeline: helpers.SourceTypeLabelName("APPLICATION_ALL"),
		})
		return []Element{
			routedPipelines,
			FromLabel{
				Desc:    "Copying unrouted application to pipelines",
				InLabel: helpers.SourceTypeLabelName("APPLICATION_ALL"),
				SubElements: []Element{
					Match{
						MatchTags: "**",
						MatchElement: Copy{
							Stores: CopyToLabels(helpers.LabelNames(unRoutedPipelines)),
						},
					},
				},
			},
		}
	}
}

func AuditToPipeline(spec *logging.ClusterLogForwarderSpec, op Options) []Element {
	return []Element{
		SourceTypeToPipeline(logging.InputNameAudit, spec, op),
	}
}

func InfraToPipeline(spec *logging.ClusterLogForwarderSpec, op Options) []Element {
	return []Element{
		SourceTypeToPipeline(logging.InputNameInfrastructure, spec, op),
	}
}

func AddLogSourceType(logtype string) Element {
	return Filter{
		MatchTags: "**",
		Element: RecordModifier{
			Records: []Record{
				{
					Key:        "log_type",
					Expression: logtype,
				},
			},
		},
	}
}
