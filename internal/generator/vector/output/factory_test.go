package output

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("output/factory.go", func() {

	DescribeTable("#New", func(o obs.OutputSpec, secrets map[string]*corev1.Secret, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", exp, err))
		}
		Expect(string(exp)).To(EqualConfigFrom(New(o, []string{"application"}, secrets, &Output{}, framework.Options{})))

	},
		Entry("should add output throttling when present",
			obs.OutputSpec{
				Type: obs.OutputTypeLoki,
				Name: "default-loki-apps",
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
					},
					Authentication: &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							Key: constants.TokenKey,
							Secret: &corev1.LocalObjectReference{
								Name: constants.LogCollectorToken,
							},
						},
					},
				},
				Limit: &obs.LimitSpec{
					MaxRecordsPerSecond: 100,
				},
				TLS: &obs.OutputTLSSpec{
					TLSSpec: obs.TLSSpec{
						CA: &obs.ConfigMapOrSecretKey{
							Secret: &corev1.LocalObjectReference{
								Name: constants.LogCollectorToken,
							},
							Key: "service-ca.crt",
						},
					},
				},
			},
			map[string]*corev1.Secret{
				constants.LogCollectorToken: {
					Data: map[string][]byte{
						"token": []byte("token-for-loki"),
					},
				},
			},
			"factory_test_loki_with_throttle.toml",
		),
	)
})
