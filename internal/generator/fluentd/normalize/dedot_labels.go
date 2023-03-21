package normalize

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
)

func DedotLabels() Element {
	modRecord := RecordModifier{
		Records: []Record{
			{
				Key: "_dummy_",
				// Replace namespace label names that have '.' & '/' with '_'
				Expression: `${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}`,
			},
			{
				Key: "_dummy2_",
				// Replace label names that have '.' & '/' with '_'
				Expression: `${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}`,
			},
			{
				Key: "_dummy3_",
				// Replace flattened label names that have '.' & '/' with '_'
				Expression: `${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}`,
			},
		},
		RemoveKeys: []string{"_dummy_, _dummy2_, _dummy3_"},
	}

	return Filter{
		Desc:      "dedot namespace_labels and rebuild message field if present",
		MatchTags: "**",
		Element:   modRecord,
	}
}
