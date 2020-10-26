package test_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// Match unique name suffix (-HHMMSSxxxxxxxx).
	suffix    = "-[0-9]{6}[0-9a-f]{8}"
	suffixLen = 1 + 6 + 8
)

var _ = Describe("Helpers", func() {
	Describe("Unmarshal", func() {
		It("unmarshals YAML", func() {
			var m map[string]string
			MustUnmarshal("a: b\nx: \"y\"\n", &m)
			Expect(m).To(Equal(map[string]string{"a": "b", "x": "y"}))
		})

		It("unmarshals JSON", func() {
			var m map[string]string
			MustUnmarshal(`{"a":"b", "x":"y"}`, &m)
			Expect(m).To(Equal(map[string]string{"a": "b", "x": "y"}))
		})
	})

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
			name := UniqueName(strings.Repeat("ghijklmnop", 100))
			Expect(validation.IsDNS1035Label(name)).To(BeNil())
			Expect(name).To(MatchRegexp(name[:validation.DNS1123LabelMaxLength-suffixLen] + suffix))
		})
	})

	Describe("CurrentUniqueName", func() {
		It("uses test name", func() {
			Expect(UniqueNameForTest()).To(MatchRegexp("uses-test-name" + suffix))
		})
	})
})
