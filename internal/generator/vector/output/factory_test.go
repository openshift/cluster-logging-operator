package output

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("output/factory.go", func() {
	defer GinkgoRecover()
	Skip("TODO: Enable me after rewire")
	// TODO: I don't this these tests are relevant anymore/ or need to redo them
	DescribeTable("#New", func(o logging.OutputSpec, secrets map[string]*corev1.Secret, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", exp, err))
		}
		Expect(string(exp)).To(EqualConfigFrom(New(o, []string{"application"}, secrets, &Output{}, framework.Options{})))

	},
		Entry("should honor global minTLSVersion & ciphers with loki as the default logstore regardless of the feature gate setting",
			logging.OutputSpec{
				Type: logging.OutputTypeLoki,
				Name: "default-loki-apps",
				URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
			},
			map[string]*corev1.Secret{
				constants.LogCollectorToken: {
					Data: map[string][]byte{
						"token": []byte("token-for-loki"),
					},
				},
			},
			"factory_test_loki_no_throttle.toml",
		),
		Entry("should add output throttling when present",
			logging.OutputSpec{
				Type: logging.OutputTypeLoki,
				Name: "default-loki-apps",
				URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
				Limit: &logging.LimitSpec{
					MaxRecordsPerSecond: 100,
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
