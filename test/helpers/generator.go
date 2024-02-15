package helpers

import (
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

type ConfGenerateTest struct {
	Desc    string
	CLFSpec logging.ClusterLogForwarderSpec
	CLSpec  logging.CollectionSpec
	// key:Output Name, value: secret for the Output
	Secrets      map[string]*corev1.Secret
	Options      framework.Options
	ExpectedConf string
}

type GenerateFunc func(logging.CollectionSpec, map[string]*corev1.Secret, logging.ClusterLogForwarderSpec, framework.Options) []framework.Element

func TestGenerateConfWith(gf GenerateFunc) func(ConfGenerateTest) {
	return TestGenerateConfAndFormatWith(gf, nil)
}

func TestGenerateConfAndFormatWith(gf GenerateFunc, format func(string) string) func(ConfGenerateTest) {
	if format == nil {
		format = func(conf string) string {
			return conf
		}
	}
	return func(testcase ConfGenerateTest) {
		g := framework.MakeGenerator()
		if testcase.Options == nil {
			testcase.Options = framework.Options{}
		}
		e := gf(testcase.CLSpec, testcase.Secrets, testcase.CLFSpec, testcase.Options)
		conf, err := g.GenerateConf(e...)
		Expect(err).To(BeNil())
		Expect(format(testcase.ExpectedConf)).To(EqualTrimLines(format(conf)))
	}
}
