package common

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("source", func() {
	//Context("", func() {
	It("should generate appropriate vector config", func() {
		exp, err := embedFS.ReadFile("secret.toml")
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file secret.toml with exp config: %v", err))
		}
		id := helpers.MakeID("output", "foo")
		conf := NewVectorSecret(id, "./read-secret-data")
		Expect(string(exp)).To(EqualConfigFrom(conf), fmt.Sprintf("for exp. file secret.toml"))
	})
	//})
})
