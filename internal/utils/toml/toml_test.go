package toml

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// literalSource mimics transforms.Remap.Source / transforms.Filter.Condition: a string field
// serialized by go-toml v1 as an unescaped literal multiline string.
type literalSource string

type transform struct {
	Type   string        `toml:"type"`
	Source literalSource `toml:"source" multiline:"true" literal:"true"`
}

type config struct {
	Transforms []*transform `toml:"transforms"`
}

var _ = Describe("Marshal", func() {

	Context("with a literal multiline field", func() {

		It("should marshal a benign VRL source", func() {
			c := &config{Transforms: []*transform{{Type: "remap", Source: `. = parse_json!(.message)`}}}
			out, err := Marshal(c)
			Expect(err).To(BeNil())
			Expect(out).To(ContainSubstring("'''"))
			Expect(out).To(ContainSubstring("parse_json!"))
		})

		It("should allow a value with a single quote (not the terminator)", func() {
			c := &config{Transforms: []*transform{{Type: "remap", Source: `match(to_string(.foo) ?? "", r'x')`}}}
			out, err := Marshal(c)
			Expect(err).To(BeNil())
			Expect(out).To(ContainSubstring("r'x'"))
		})

		It("should reject a value containing the literal multiline terminator", func() {
			c := &config{Transforms: []*transform{{Type: "remap", Source: `x''' ` + "\n[sources.evil]\ntype = \"file\"\nsource = '''\n. = ."}}}
			_, err := Marshal(c)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("literal multiline terminator"))
		})

		It("should report the offending nested field path", func() {
			c := &config{Transforms: []*transform{{Type: "remap", Source: "ok"}, {Type: "remap", Source: "bad'''injection"}}}
			_, err := Marshal(c)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Transforms[1].Source"))
		})
	})

	It("should not scan fields that are not literal multiline", func() {
		type plain struct {
			Name string `toml:"name"`
		}
		out, err := Marshal(&plain{Name: "has ''' quotes"})
		Expect(err).To(BeNil())
		Expect(out).To(ContainSubstring("has ''' quotes"))
	})
})
