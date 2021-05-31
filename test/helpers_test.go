package test_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// Match unique name suffix (-HHMMSSxxxxxxxx).
	suffix    = "-[0-9]{6}-[0-9a-f]{8}"
	suffixLen = 2 + 6 + 8
)

var _ = Describe("Helpers", func() {

	// Test data for unmarshal and string functions.

	m := map[string]string{"a": "b", "c": "d"}

	cm := corev1.ConfigMap{Data: m}

	clf := loggingv1.ClusterLogForwarder{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogForwarder",
			APIVersion: "logging.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance",
			Namespace: "openshift-logging",
		},
		Spec: loggingv1.ClusterLogForwarderSpec{
			Outputs: []loggingv1.OutputSpec{
				{
					Name: "test-forward",
					Type: "fluentdForward",
					URL:  "tcp://localhost:24224",
				},
			},
			Pipelines: []loggingv1.PipelineSpec{
				{
					Name:       "test-forward",
					InputRefs:  []string{"application"},
					OutputRefs: []string{"test-forward"}},
			},
		},
	}

	Describe("Unmarshal", func() {
		It("unmarshals YAML", func() {
			var m2 map[string]string
			MustUnmarshal(`{"a":"b", "c":"d"}`, &m2)
			Expect(m2).To(Equal(m))
		})

		It("unmarshals JSON", func() {
			var m2 map[string]string
			MustUnmarshal(`{"a":"b", "x":"y"}`, &m2)
			Expect(m).To(Equal(map[string]string{"a": "b", "c": "d"}))
		})
	})

	DescribeTable("JSONLine",
		func(v interface{}, s string) { Expect(JSONLine(v)).To(EqualLines(s)) },
		Entry("map", m, `{"a":"b","c":"d"}`),
		Entry("configmap", cm, `{"metadata":{"creationTimestamp":null},"data":{"a":"b","c":"d"}}`),
		Entry("clf", clf, `{"kind":"ClusterLogForwarder","apiVersion":"logging.openshift.io/v1","metadata":{"name":"instance","namespace":"openshift-logging","creationTimestamp":null},"spec":{"outputs":[{"name":"test-forward","type":"fluentdForward","url":"tcp://localhost:24224"}],"pipelines":[{"outputRefs":["test-forward"],"inputRefs":["application"],"name":"test-forward"}]},"status":{}}`),
	)

	DescribeTable("JSONString",
		func(v interface{}, s string) { Expect(JSONString(v)).To(EqualLines(s)) },
		Entry("map", m, `{
  "a": "b",
  "c": "d"
}`),
		Entry("configmap", cm, `{
  "metadata": {
    "creationTimestamp": null
  },
  "data": {
    "a": "b",
    "c": "d"
  }
}`),
		Entry("clf", clf, `{
	"kind": "ClusterLogForwarder",
	"apiVersion": "logging.openshift.io/v1",
	"metadata": {
		"name": "instance",
		"namespace": "openshift-logging",
		"creationTimestamp": null
	},
	"spec": {
		"outputs": [
			{
				"name": "test-forward",
				"type": "fluentdForward",
				"url": "tcp://localhost:24224"
			}
		],
		"pipelines": [
			{
				"outputRefs": [
					"test-forward"
				],
				"inputRefs": [
					"application"
				],
				"name": "test-forward"
			}
		]
	},
	"status": {}
	}`),
	)

	DescribeTable("YAMLString",
		func(v interface{}, s string) { Expect(YAMLString(v)).To(EqualLines(s)) },
		Entry("map", m, `
  a: b
  c: d
`),
		Entry("configmap", cm, `
  data:
    a: b
    c: d
  metadata:
    creationTimestamp: null
`),
		Entry("clf", clf, `
   apiVersion: logging.openshift.io/v1
    kind: ClusterLogForwarder
    metadata:
      creationTimestamp: null
      name: instance
      namespace: openshift-logging
    spec:
      outputs:
      - name: test-forward
        type: fluentdForward
        url: tcp://localhost:24224
      pipelines:
      - inputRefs:
        - application
        name: test-forward
        outputRefs:
        - test-forward
    status: {}
`))

	Describe("UniqueName", func() {
		It("generates unique names", func() {
			names := map[string]bool{}
			for i := 0; i < 100; i++ {
				name := UniqueName("x")
				Expect(name).To(MatchRegexp("x" + suffix))
				Expect(names).NotTo(HaveKey(name), "not unique")
			}
		})

		It("cleans up an illegal name", func() {
			name := UniqueName("x--y!-@#z--")
			Expect(validation.IsDNS1123Label(name)).To(BeNil(), name)
			Expect(name).To(MatchRegexp("x-y-z" + suffix))
		})

		It("truncates a long prefix", func() {
			prefix := strings.Repeat("ghijklmnop", 100)
			name := UniqueName(prefix)
			Expect(name).To(HaveLen(validation.DNS1123LabelMaxLength))
			Expect(validation.IsDNS1035Label(name)).To(BeNil())
			Expect(name).To(MatchRegexp(prefix[:validation.DNS1123LabelMaxLength-suffixLen] + suffix))
		})
	})

	Describe("CurrentUniqueName", func() {
		It("uses test name", func() {
			Expect(UniqueNameForTest()).To(MatchRegexp("uses-test-name" + suffix))
		})
	})

	Describe("GitRoot", func() {
		It("finds the repository root", func() {
			wd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(wd).To(HavePrefix(GitRoot()))
			Expect(GitRoot("test", "helpers_test.go")).To(BeARegularFile())
		})
	})
})
