package v1_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = test.Debug // Load the test package for side-effects.

var _ = Describe("Conditions", func() {
	It("serializes as an array", func() {
		bad := NewCondition("bad", true, "smelly", "Peeew")
		good := NewCondition("good", true, "lovely", "Ahhh")
		bad.LastTransitionTime = metav1.Time{} // Clear time for comparison
		good.LastTransitionTime = metav1.Time{}
		conds := NewConditions(bad, good)
		json, err := json.Marshal(conds)
		Expect(err).To(Succeed())
		Expect(string(json)).To(Equal(`[{"type":"bad","status":"True","reason":"smelly","message":"Peeew","lastTransitionTime":null},{"type":"good","status":"True","reason":"lovely","message":"Ahhh","lastTransitionTime":null}]`))
	})

	It("updates timestamps", func() {
		before := time.Now()
		cs := Conditions{}
		cs.SetNew("x", false, "y")
		Expect(cs["x"].LastTransitionTime.Time).To(BeTemporally(">=", before))
	})

	Context("evaluates conditions", func() {
		cs := Conditions{}
		cs.SetNew("TrueType", true, "")
		cs.SetNew("FalseType", false, "")
		// Absence means Unknown

		DescribeTable("test status values",
			func(ctype string, isTrue, isFalse, isUnknown bool) {
				c := cs[ConditionType(ctype)]
				Expect(c.IsTrue()).To(Equal(isTrue))
				Expect(c.IsFalse()).To(Equal(isFalse))
				Expect(c.IsUnknown()).To(Equal(isUnknown))
			},
			Entry("true", "TrueType", true, false, false),
			Entry("false", "FalseType", false, true, false),
			Entry("unkown", "UnknownType", false, false, true),
		)
	})
})
