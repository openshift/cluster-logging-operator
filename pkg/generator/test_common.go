package generator

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

type ConfGenerateTest struct {
	Desc    string
	CLFSpec logging.ClusterLogForwarderSpec
	CLSpec  logging.ClusterLoggingSpec
	// key:Output Name, value: secret for the Output
	Secrets      map[string]*corev1.Secret
	Options      Options
	ExpectedConf string
}

type GenerateFunc func(logging.ClusterLoggingSpec, map[string]*corev1.Secret, logging.ClusterLogForwarderSpec, Options) []Element

func TestGenerateConfWith(gf GenerateFunc) func(ConfGenerateTest) {
	return func(testcase ConfGenerateTest) {
		g := MakeGenerator()
		e := gf(testcase.CLSpec, testcase.Secrets, testcase.CLFSpec, testcase.Options)
		conf, err := g.GenerateConf(e...)
		Expect(err).To(BeNil())
		diff := cmp.Diff(
			strings.Split(strings.TrimSpace(testcase.ExpectedConf), "\n"),
			strings.Split(strings.TrimSpace(conf), "\n"))
		if diff != "" {
			//b, _ := json.MarshalIndent(e, "", " ")
			//fmt.Printf("elements:\n%s\n", string(b))
			fmt.Println(conf)
			fmt.Printf("diff: %s", diff)
		}
		Expect(diff).To(Equal(""))
	}
}
