package unit

import (
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

type ConfGenerateTest struct {
	Desc    string
	CLFSpec logging.ClusterLogForwarderSpec
	CLSpec  logging.ClusterLoggingSpec
	// key:Output Name, value: secret for the Output
	Secrets      map[string]*corev1.Secret
	Options      generator.Options
	ExpectedConf string
}

type GenerateFunc func(logging.ClusterLoggingSpec, map[string]*corev1.Secret, logging.ClusterLogForwarderSpec, generator.Options) []generator.Element

func TestGenerateConfWith(gf GenerateFunc) func(ConfGenerateTest) {
	return func(testcase ConfGenerateTest) {
		g := generator.MakeGenerator()
		if testcase.Options == nil {
			testcase.Options = generator.Options{}
		}
		e := gf(testcase.CLSpec, testcase.Secrets, testcase.CLFSpec, testcase.Options)
		conf, err := g.GenerateConf(e...)
		Expect(err).To(BeNil())
		Expect(conf).To(EqualTrimLines(testcase.ExpectedConf))
	}
}
