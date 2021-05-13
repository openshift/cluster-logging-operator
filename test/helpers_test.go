package test_test

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

func nameRegexp(name string) string {
	return fmt.Sprintf("%v-[0-9a-z]{8}", name)
}

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
				Expect(name).To(MatchRegexp(nameRegexp("x")))
				Expect(names).NotTo(HaveKey(name), "not unique")
				names[name] = true
			}
		})

		It("cleans up an illegal name", func() {
			name := UniqueName("x--y!-@#z--")
			Expect(validation.IsDNS1123Label(name)).To(BeNil(), name)
			Expect(name).To(MatchRegexp(nameRegexp("x-y-z")))
		})

		It("truncates a long prefix", func() {
			longName := strings.Repeat("ghijklmnop", 100)
			name := UniqueName(longName)
			Expect(validation.IsDNS1035Label(name)).To(BeNil())
			Expect(longName).To(HavePrefix(name[:10]))
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

	Describe("HTMLBodyText", func() {
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<!--[if IE]><meta http-equiv="X-UA-Compatible" content="IE=edge"><![endif]-->
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Go profiling cheat-sheet</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Open+Sans:300,300italic,400,400italic,600,600italic%7CNoto+Serif:400,400italic,700,700italic%7CDroid+Sans+Mono:400,700">
</head>
<body class="article">
<div id="header">
<h1>Go profiling cheat-sheet</h1>
</div>
<div id="content">
<div id="preamble">
<div class="sectionbody">
<div class="paragraph">
<p>This is a short list of things I do to profile in go, not an introduction to profiling.</p>
</div>
</div>
</div>
</div>
</body>
</html>
`
		It("extracts the HTML body text", func() {
			r := strings.NewReader(html)
			b, err := test.HTMLBodyText(r)
			ExpectOK(err)
			Expect(string(b)).To(Equal("Go profiling cheat-sheet\n\nThis is a short list of things I do to profile in go, not an introduction to profiling."))
		})
	})

	Describe("MapIndex", func() {
		It("indexes into a map[string]interface{}", func() {
			m := map[string]interface{}{
				"a": "1",
				"b": map[string]interface{}{
					"bb": "2",
					"c": map[string]interface{}{
						"cc": "3",
					},
				},
			}
			Expect(MapIndices(m, "a")).To(Equal("1"))
			Expect(MapIndices(m, "b", "bb")).To(Equal("2"))
			Expect(MapIndices(m, "b", "c", "cc")).To(Equal("3"))
		})
	})
})
