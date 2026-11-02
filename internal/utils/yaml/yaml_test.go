package yaml_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	utilyaml "github.com/openshift/cluster-logging-operator/internal/utils/yaml"
)

func TestYAML(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "YAML Utils Suite")
}

var _ = Describe("YAML Utilities", func() {
	type TestStruct struct {
		Name        string            `yaml:"name"`
		Age         int               `yaml:"age,omitempty"`
		Tags        []string          `yaml:"tags,omitempty"`
		Metadata    map[string]string `yaml:"metadata,omitempty"`
		OmittedZero int               `yaml:"omitted,omitempty"`
	}

	Context("MustMarshal", func() {
		It("should marshal struct to YAML string", func() {
			obj := TestStruct{
				Name: "test",
				Age:  30,
				Tags: []string{"tag1", "tag2"},
			}

			result := utilyaml.MustMarshal(obj)

			Expect(result).To(ContainSubstring("name: test"))
			Expect(result).To(ContainSubstring("age: 30"))
			Expect(result).To(ContainSubstring("tags:"))
			Expect(result).To(ContainSubstring("- tag1"))
			Expect(result).To(ContainSubstring("- tag2"))
		})

		It("should omit empty fields with omitempty tag", func() {
			obj := TestStruct{
				Name: "test",
			}

			result := utilyaml.MustMarshal(obj)

			Expect(result).To(ContainSubstring("name: test"))
			Expect(result).NotTo(ContainSubstring("age:"))
			Expect(result).NotTo(ContainSubstring("tags:"))
			Expect(result).NotTo(ContainSubstring("omitted:"))
		})

		It("should handle panics during marshal", func() {
			// Channels cause yaml.v3 to panic
			obj := make(chan int)

			// Should recover from panic and return empty string
			result := utilyaml.MustMarshal(obj)

			Expect(result).To(Equal(""))
		})
	})

	Context("Marshal", func() {
		It("should marshal to YAML bytes", func() {
			obj := TestStruct{
				Name: "test",
				Age:  25,
			}

			bytes, err := utilyaml.Marshal(obj)

			Expect(err).To(BeNil())
			Expect(string(bytes)).To(ContainSubstring("name: test"))
			Expect(string(bytes)).To(ContainSubstring("age: 25"))
		})

		It("should marshal simple types", func() {
			obj := map[string]interface{}{
				"key":   "value",
				"count": 42,
			}

			bytes, err := utilyaml.Marshal(obj)

			Expect(err).To(BeNil())
			Expect(string(bytes)).To(ContainSubstring("key: value"))
			Expect(string(bytes)).To(ContainSubstring("count: 42"))
		})
	})

	Context("Unmarshal", func() {
		It("should unmarshal YAML bytes to struct", func() {
			yamlData := `
name: test
age: 30
tags:
  - tag1
  - tag2
metadata:
  key1: value1
  key2: value2
`
			var result TestStruct

			err := utilyaml.Unmarshal([]byte(yamlData), &result)

			Expect(err).To(BeNil())
			Expect(result.Name).To(Equal("test"))
			Expect(result.Age).To(Equal(30))
			Expect(result.Tags).To(Equal([]string{"tag1", "tag2"}))
			Expect(result.Metadata).To(HaveKeyWithValue("key1", "value1"))
			Expect(result.Metadata).To(HaveKeyWithValue("key2", "value2"))
		})

		It("should return error for invalid YAML", func() {
			invalidYAML := `
name: test
  invalid: indentation
`
			var result TestStruct

			err := utilyaml.Unmarshal([]byte(invalidYAML), &result)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("MarshalToMap", func() {
		It("should convert struct to map", func() {
			obj := TestStruct{
				Name: "test",
				Age:  30,
				Tags: []string{"tag1", "tag2"},
			}

			result, err := utilyaml.MarshalToMap(obj)

			Expect(err).To(BeNil())
			Expect(result).To(HaveKeyWithValue("name", "test"))
			Expect(result).To(HaveKeyWithValue("age", 30))
			Expect(result).To(HaveKey("tags"))
		})

		It("should omit empty fields with omitempty", func() {
			obj := TestStruct{
				Name: "test",
			}

			result, err := utilyaml.MarshalToMap(obj)

			Expect(err).To(BeNil())
			Expect(result).To(HaveKeyWithValue("name", "test"))
			Expect(result).NotTo(HaveKey("age"))
			Expect(result).NotTo(HaveKey("tags"))
			Expect(result).NotTo(HaveKey("omitted"))
		})
	})

	Context("Round-trip", func() {
		It("should preserve data through marshal and unmarshal", func() {
			original := TestStruct{
				Name: "roundtrip",
				Age:  42,
				Tags: []string{"a", "b", "c"},
				Metadata: map[string]string{
					"env":  "test",
					"team": "platform",
				},
			}

			// Marshal to YAML
			bytes, err := utilyaml.Marshal(original)
			Expect(err).To(BeNil())

			// Unmarshal back
			var result TestStruct
			err = utilyaml.Unmarshal(bytes, &result)
			Expect(err).To(BeNil())

			// Verify equality
			Expect(result).To(Equal(original))
		})
	})
})
