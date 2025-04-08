package lokistack

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	lokioutput "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

var _ = Describe("#GenerateOutput", func() {
	const (
		lokistackOut      = "lokistack-out"
		lokistackTarget   = "test-lokistack"
		lokistackOutApp   = lokistackOut + "-" + string(obs.InputTypeApplication)
		lokistackOutAudit = lokistackOut + "-" + string(obs.InputTypeAudit)
	)

	var (
		initOutputSpec = func(outName string) obs.OutputSpec {
			spec := obs.OutputSpec{
				Name: outName,
				Type: obs.OutputTypeLokiStack,
				LokiStack: &obs.LokiStack{
					Target: obs.LokiStackTarget{
						Name:      lokistackTarget,
						Namespace: constants.OpenshiftNS,
					},
					Authentication: &obs.LokiStackAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				},
			}
			return spec
		}

		setOtelDataModel = func(spec *obs.OutputSpec) {
			spec.LokiStack.DataModel = obs.LokiStackDataModelOpenTelemetry
		}
	)

	DescribeTable("when generating a spec from a lokistack", func(expSpec obs.OutputSpec, tenant string, visit func(spec *obs.OutputSpec)) {
		lokiStack := initOutputSpec(lokistackOut)
		if visit != nil {
			visit(&lokiStack)
		}

		spec := GenerateOutput(lokiStack, tenant)
		Expect(spec).To(Equal(expSpec))
	},
		Entry("with ViaQ should generate a loki output spec with desired tenant",
			obs.OutputSpec{
				Name: lokistackOutApp,
				Type: obs.OutputTypeLoki,
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
					},
					Authentication: &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				},
			},
			string(obs.InputTypeApplication),
			nil,
		),
		Entry("with ViaQ and customized label keys should generate a loki output spec with desired tenant and label keys",
			obs.OutputSpec{
				Name: lokistackOutAudit,
				Type: obs.OutputTypeLoki,
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit",
					},
					Authentication: &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
					LabelKeys: []string{
						"log_type",
						"objectRef.apiGroup",
					},
				},
			},
			string(obs.InputTypeAudit),
			func(spec *obs.OutputSpec) {
				spec.LokiStack.LabelKeys = &obs.LokiStackLabelKeys{
					Audit: &obs.LokiStackTenantLabelKeys{
						IgnoreGlobal: true,
						LabelKeys: []string{
							"log_type",
							"objectRef.apiGroup",
						},
					},
				}

			},
		),
		Entry("with ViaQ and customized label keys that include globals should generate a loki output spec with desired tenant and label keys",
			obs.OutputSpec{
				Name: lokistackOutAudit,
				Type: obs.OutputTypeLoki,
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit",
					},
					Authentication: &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
					LabelKeys: sets.NewString(lokioutput.DefaultLabelKeys...).Insert("objectRef.apiGroup").List(),
				},
			},
			string(obs.InputTypeAudit),
			func(spec *obs.OutputSpec) {
				spec.LokiStack.LabelKeys = &obs.LokiStackLabelKeys{
					Audit: &obs.LokiStackTenantLabelKeys{
						IgnoreGlobal: false,
						LabelKeys: []string{
							"objectRef.apiGroup",
						},
					},
				}

			},
		),
		Entry("with OTel datamodel should generate an OTLP output spec to a loki OTLP endpoint the with desired tenant",
			obs.OutputSpec{
				Name: lokistackOutApp,
				Type: obs.OutputTypeOTLP,
				OTLP: &obs.OTLP{
					URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
					Authentication: &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				},
			},
			string(obs.InputTypeApplication),
			setOtelDataModel,
		),
	)

	DescribeTable(
		"LabelKeys logic for LokiStack tenants", func(labelKeys *obs.LokiStackLabelKeys, tenant string, wantKeys []string) {
			testDefaultKeys := []string{
				"default_one",
				"default_two",
			}
			keys := lokiStackLabelKeysForTenant(labelKeys, tenant, testDefaultKeys)
			Expect(keys).To(Equal(wantKeys))
		},
		Entry(
			"no config",
			nil,
			string(obs.InputTypeApplication),
			nil,
		),
		Entry(
			"empty slices -> still nil",
			&obs.LokiStackLabelKeys{
				Global: []string{},
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{},
				},
			},
			string(obs.InputTypeApplication),
			nil,
		),
		Entry(
			"only global",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"global_one",
				"global_two",
			},
		),
		Entry(
			"only tenant",
			&obs.LokiStackLabelKeys{
				Application: &obs.LokiStackTenantLabelKeys{
					IgnoreGlobal: true,
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"only tenant but with defaults",
			&obs.LokiStackLabelKeys{
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"default_one",
				"default_two",
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"global and tenant",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
				},
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"global_one",
				"global_two",
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"global and tenant, ignore global",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
				},
				Application: &obs.LokiStackTenantLabelKeys{
					IgnoreGlobal: true,
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"global and tenant, ignore duplicates",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
					"common_one",
				},
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
						"common_one",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"common_one",
				"global_one",
				"global_two",
				"tenant_one",
				"tenant_two",
			},
		),
	)
})
