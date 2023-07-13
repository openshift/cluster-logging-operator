package migrations

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("MigrateDefaultOutput", func() {

	var (
		pipelines []logging.PipelineSpec
		outputs   []logging.OutputSpec
		spec      logging.ClusterLogForwarderSpec
		esSpec    *logging.Elasticsearch
		logstore  *logging.LogStoreSpec
		extras    map[string]bool
	)

	BeforeEach(func() {
		esSpec = &logging.Elasticsearch{
			ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
				StructuredTypeKey: "foo.bar",
			},
		}
		pipelines = []logging.PipelineSpec{
			{
				Name:       "test",
				OutputRefs: []string{"first", "second"},
				InputRefs:  []string{logging.InputNameApplication},
			},
		}
		outputs = []logging.OutputSpec{
			{
				Name: "first",
				Type: logging.OutputTypeElasticsearch,
				OutputTypeSpec: logging.OutputTypeSpec{
					Elasticsearch: esSpec,
				},
			},
		}
		spec = logging.ClusterLogForwarderSpec{
			Outputs:   outputs,
			Pipelines: pipelines,
		}
		extras = map[string]bool{}
	})

	It("should not add service account name if forwarder not named `instance` and not in `openshift-logging` namespace", func() {
		forwarderSpec, extras := MigrateClusterLogForwarderSpec("test-ns", "test-clf", spec, logstore, extras, constants.CollectorName)
		Expect(forwarderSpec).To(Equal(spec))
		Expect(extras).To(Equal(map[string]bool{}))
	})

	It("should not add the default OutputSpec when it is not referenced by a pipeline", func() {
		// This is equal to returning (spec, nil) and will only pass if 2nd param is nil
		forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)
		spec.ServiceAccountName = constants.CollectorServiceAccountName
		Expect(forwarderSpec).To(Equal(spec))
		Expect(extras).To(Equal(map[string]bool{}))
	})

	It("should add the default OutputSpec when default logstore exists and spec is empty ", func() {
		logstore = &logging.LogStoreSpec{Type: logging.OutputTypeElasticsearch}
		forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, logging.ClusterLogForwarderSpec{}, logstore, extras, constants.CollectorName)
		Expect(forwarderSpec).To(Equal(
			logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "default_pipeline_0_",
						InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
						OutputRefs: []string{logging.OutputNameDefault},
					},
				},
				Outputs:            []logging.OutputSpec{NewDefaultOutput(nil, constants.CollectorName)},
				ServiceAccountName: constants.CollectorServiceAccountName,
			},
		))
		Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
	})

	It("generates default configuration for empty spec with LokiStack log store", func() {
		logstore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{
				Name: "lokistack-testing",
			},
		}

		spec, extras = MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, logging.ClusterLogForwarderSpec{}, logstore, extras, constants.CollectorName)

		Expect(spec).To(Equal(
			logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "default_loki_pipeline_0_",
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{"default-loki-apps"},
					},
					{
						Name:       "default_loki_pipeline_1_",
						InputRefs:  []string{logging.InputNameInfrastructure},
						OutputRefs: []string{"default-loki-infra"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: "default-loki-apps",
						Type: logging.OutputTypeLoki,
						URL:  "https://lokistack-testing-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
					},
					{
						Name: "default-loki-infra",
						Type: logging.OutputTypeLoki,
						URL:  "https://lokistack-testing-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure",
					},
				},
				ServiceAccountName: constants.CollectorServiceAccountName,
			},
		))
		Expect(extras).To(Equal(map[string]bool{}))
	})

	It("processes custom pipelines to default LokiStack log store", func() {
		logstore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{
				Name: "lokistack-testing",
			},
		}
		spec = logging.ClusterLogForwarderSpec{
			Pipelines: []logging.PipelineSpec{
				{
					InputRefs:  []string{"audit"},
					OutputRefs: []string{"default"},
				},
			},
		}

		spec, extras = MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)

		Expect(spec).To(Equal(
			logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "default_loki_pipeline_0_",
						InputRefs:  []string{logging.InputNameAudit},
						OutputRefs: []string{"default-loki-audit"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: "default-loki-audit",
						Type: logging.OutputTypeLoki,
						URL:  "https://lokistack-testing-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit",
					},
				},
				ServiceAccountName: constants.CollectorServiceAccountName,
			},
		))
		Expect(extras).To(Equal(map[string]bool{}))
	})

	Context("when a pipeline references 'default'", func() {

		var exp logging.ClusterLogForwarderSpec
		BeforeEach(func() {
			logstore = &logging.LogStoreSpec{Type: logging.OutputTypeElasticsearch}
			pipelines[0].OutputRefs = append(spec.Pipelines[0].OutputRefs, logging.OutputNameDefault)
			spec = logging.ClusterLogForwarderSpec{
				Outputs:   outputs,
				Pipelines: pipelines,
			}
		})

		Context("and outputs does not explicitly spec 'default'", func() {
			BeforeEach(func() {
				exp = *spec.DeepCopy()
				exp.Outputs = append(outputs, NewDefaultOutput(nil, constants.CollectorName))
				exp.ServiceAccountName = constants.CollectorServiceAccountName
			})

			It("should add the default OutputSpec", func() {
				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)
				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v", pipelines))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})

			It("should add the default OutputSpec and OutputDefaults when OutputDefaults are spec'd", func() {
				spec.OutputDefaults = &logging.OutputDefaults{
					Elasticsearch: &logging.ElasticsearchStructuredSpec{
						StructuredTypeKey: "foo.bar",
					},
				}
				exp.Outputs[1].Elasticsearch = &logging.Elasticsearch{ElasticsearchStructuredSpec: *spec.OutputDefaults.Elasticsearch}
				exp.OutputDefaults = spec.OutputDefaults

				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)

				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v and OutputDefault %v", pipelines, spec.OutputDefaults))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})
		})

		Context("and outputs includes an OutputSpec named 'default'", func() {
			var tobereplaced logging.OutputSpec
			BeforeEach(func() {
				tobereplaced = logging.OutputSpec{
					Name:   logging.OutputNameDefault,
					Type:   logging.OutputTypeElasticsearch,
					URL:    "thiswillgetreplaced",
					Secret: &logging.OutputSecretSpec{Name: "replacem"},
				}

			})

			It("should replace the OutputSpec with the default OutputSpec", func() {
				spec = logging.ClusterLogForwarderSpec{
					Outputs:   append(outputs, tobereplaced),
					Pipelines: pipelines,
				}
				exp = *spec.DeepCopy()
				exp.Outputs = append(outputs, NewDefaultOutput(nil, constants.CollectorName))
				exp.ServiceAccountName = constants.CollectorServiceAccountName

				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)
				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v", pipelines))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})

			It("should replace the OutputSpec with the default OutputSpec and use the config (e.g. structureTypeKey) defined in the original OutputSpec", func() {
				tobereplaced.Elasticsearch = esSpec
				spec = logging.ClusterLogForwarderSpec{
					Outputs:        append(outputs, tobereplaced),
					Pipelines:      pipelines,
					OutputDefaults: &logging.OutputDefaults{Elasticsearch: &logging.ElasticsearchStructuredSpec{StructuredTypeKey: "abc"}},
				}
				exp = *spec.DeepCopy()
				exp.Outputs = append(outputs, NewDefaultOutput(&logging.OutputDefaults{Elasticsearch: &esSpec.ElasticsearchStructuredSpec}, constants.CollectorName))
				exp.ServiceAccountName = constants.CollectorServiceAccountName

				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)
				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v and ElasticsearchSpec %v", pipelines, esSpec))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})
		})

	})
	Context("when a pipeline references 'default'", func() {

		var exp logging.ClusterLogForwarderSpec
		BeforeEach(func() {
			logstore = &logging.LogStoreSpec{Type: logging.OutputTypeElasticsearch}
			pipelines[0].OutputRefs = append(spec.Pipelines[0].OutputRefs, logging.OutputNameDefault)
			spec = logging.ClusterLogForwarderSpec{
				Outputs:   outputs,
				Pipelines: pipelines,
			}
		})

		Context("and outputs does not explicitly spec 'default'", func() {
			BeforeEach(func() {
				exp = *spec.DeepCopy()
				exp.Outputs = append(outputs, NewDefaultOutput(nil, constants.CollectorName))
				exp.ServiceAccountName = constants.CollectorServiceAccountName
			})

			It("should add the default OutputSpec", func() {
				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)
				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v", pipelines))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})
			It("should add the default OutputSpec and OutputDefaults when OutputDefaults are spec'd", func() {
				spec.OutputDefaults = &logging.OutputDefaults{
					Elasticsearch: &logging.ElasticsearchStructuredSpec{
						StructuredTypeKey: "foo.bar",
					},
				}
				exp.Outputs[1].Elasticsearch = &logging.Elasticsearch{ElasticsearchStructuredSpec: *spec.OutputDefaults.Elasticsearch}
				exp.OutputDefaults = spec.OutputDefaults

				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)

				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v and OutputDefault %v", pipelines, spec.OutputDefaults))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})
		})

		Context("and outputs includes an OutputSpec named 'default'", func() {
			var tobereplaced logging.OutputSpec
			BeforeEach(func() {
				tobereplaced = logging.OutputSpec{
					Name:   logging.OutputNameDefault,
					Type:   logging.OutputTypeElasticsearch,
					URL:    "thiswillgetreplaced",
					Secret: &logging.OutputSecretSpec{Name: "replacem"},
				}

			})

			It("should replace the OutputSpec with the default OutputSpec", func() {
				spec = logging.ClusterLogForwarderSpec{
					Outputs:   append(outputs, tobereplaced),
					Pipelines: pipelines,
				}
				exp = *spec.DeepCopy()
				exp.Outputs = append(outputs, NewDefaultOutput(nil, constants.CollectorName))
				exp.ServiceAccountName = constants.CollectorServiceAccountName

				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)
				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v", pipelines))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})

			It("should replace the OutputSpec with the default OutputSpec and use the config (e.g. structureTypeKey) defined in the original OutputSpec", func() {
				tobereplaced.Elasticsearch = esSpec
				spec = logging.ClusterLogForwarderSpec{
					Outputs:        append(outputs, tobereplaced),
					Pipelines:      pipelines,
					OutputDefaults: &logging.OutputDefaults{Elasticsearch: &logging.ElasticsearchStructuredSpec{StructuredTypeKey: "abc"}},
				}
				exp = *spec.DeepCopy()
				exp.Outputs = append(outputs, NewDefaultOutput(&logging.OutputDefaults{Elasticsearch: &esSpec.ElasticsearchStructuredSpec}, constants.CollectorName))
				exp.ServiceAccountName = constants.CollectorServiceAccountName

				forwarderSpec, extras := MigrateClusterLogForwarderSpec(constants.OpenshiftNS, constants.SingletonName, spec, logstore, extras, constants.CollectorName)
				Expect(forwarderSpec).To(Equal(exp), fmt.Sprintf("Exp. default output because of pipeline %v and ElasticsearchSpec %v", pipelines, esSpec))
				Expect(extras).To(Equal(map[string]bool{constants.MigrateDefaultOutput: true}))
			})
		})
	})
})
