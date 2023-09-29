package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputTuningSpec", func() {
	Context("#Validate", func() {
		It("should fail when a key is not allowed", func() {
			Expect(OutputTuningSpec{"foo": "bar"}).To(Not(BeEmpty()))
		})
		It("should fail when a value is an unsupported enum", func() {
			Expect(OutputTuningSpec{bufferWhenFull: "bar"}).To(Not(BeEmpty()))
		})
		It("should fail when a value type is unsupported", func() {
			Expect(OutputTuningSpec{bufferMaxEvents: "abc"}.Validate()).To(Not(BeEmpty()), "string value")
			Expect(OutputTuningSpec{bufferMaxEvents: "0"}.Validate()).To(Not(BeEmpty()), "int value <=0")
			Expect(OutputTuningSpec{bufferMaxEvents: "-1"}.Validate()).To(Not(BeEmpty()), "int value <=0")
			Expect(OutputTuningSpec{bufferMaxEvents: "4.5"}.Validate()).To(Not(BeEmpty()), "float value")
		})
		It("should succeed for keys that are supported", func() {
			for _, key := range allowedOutputTuning.List() {
				spec := OutputTuningSpec{key: "5"}
				if key == bufferWhenFull {
					spec[bufferWhenFull] = "block"
				}
				Expect(spec.Validate()).To(BeEmpty())
			}
		})
	})
})
