package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
)

var _ = Describe("Component ID", func() {
	Context("ParseComponentID", func() {
		It("should parse type only ID", func() {
			componentType, name := types.ParseComponentID("file_log")
			Expect(componentType).To(Equal("file_log"))
			Expect(name).To(Equal(""))
		})

		It("should parse type with name", func() {
			componentType, name := types.ParseComponentID("file_log/containers")
			Expect(componentType).To(Equal("file_log"))
			Expect(name).To(Equal("containers"))
		})

		It("should parse otlphttp type only", func() {
			componentType, name := types.ParseComponentID("otlphttp")
			Expect(componentType).To(Equal("otlphttp"))
			Expect(name).To(Equal(""))
		})

		It("should parse otlphttp with name", func() {
			componentType, name := types.ParseComponentID("otlphttp/loki")
			Expect(componentType).To(Equal("otlphttp"))
			Expect(name).To(Equal("loki"))
		})

		It("should handle name with multiple slashes (edge case)", func() {
			componentType, name := types.ParseComponentID("file_log/my/custom/name")
			Expect(componentType).To(Equal("file_log"))
			Expect(name).To(Equal("my/custom/name"))
		})
	})

	Context("MakeComponentID", func() {
		It("should create ID from type only", func() {
			result := types.MakeComponentID("file_log", "")
			Expect(result).To(Equal("file_log"))
		})

		It("should create ID from type with name", func() {
			result := types.MakeComponentID("file_log", "containers")
			Expect(result).To(Equal("file_log/containers"))
		})

		It("should create otlphttp ID from type only", func() {
			result := types.MakeComponentID("otlphttp", "")
			Expect(result).To(Equal("otlphttp"))
		})

		It("should create otlphttp ID from type with name", func() {
			result := types.MakeComponentID("otlphttp", "loki")
			Expect(result).To(Equal("otlphttp/loki"))
		})
	})

	Context("Round trip", func() {
		It("should preserve type-only IDs", func() {
			id := types.MakeComponentID("file_log", "")
			parsedType, parsedName := types.ParseComponentID(id)
			Expect(parsedType).To(Equal("file_log"))
			Expect(parsedName).To(Equal(""))
		})

		It("should preserve type with name IDs", func() {
			id := types.MakeComponentID("file_log", "containers")
			parsedType, parsedName := types.ParseComponentID(id)
			Expect(parsedType).To(Equal("file_log"))
			Expect(parsedName).To(Equal("containers"))
		})

		It("should preserve otlphttp type-only IDs", func() {
			id := types.MakeComponentID("otlphttp", "")
			parsedType, parsedName := types.ParseComponentID(id)
			Expect(parsedType).To(Equal("otlphttp"))
			Expect(parsedName).To(Equal(""))
		})

		It("should preserve otlphttp with name IDs", func() {
			id := types.MakeComponentID("otlphttp", "loki")
			parsedType, parsedName := types.ParseComponentID(id)
			Expect(parsedType).To(Equal("otlphttp"))
			Expect(parsedName).To(Equal("loki"))
		})
	})
})
