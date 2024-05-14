package source_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	env   = "environment"
	prod  = "production"
	qa    = "qa"
	tier  = "tier"
	front = "frontend"
	back  = "backend"

	NotIn         = LabelSelectorOpNotIn
	In            = LabelSelectorOpIn
	Exists        = LabelSelectorOpExists
	DoesNotExists = LabelSelectorOpDoesNotExist
)

var _ = DescribeTable("#LabelSelectorFrom", func(s *LabelSelector, exp string) {
	Expect(source.LabelSelectorFrom(s)).To(Equal(exp))
},
	Entry("should be empty for a nil selector", nil, ""),
	Entry("should format a selector that exactly matches all the defined match labels", &LabelSelector{
		MatchLabels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}, `key1=value1,key2=value2`),
	Entry("should format a selector that only includes the label when the value is empty", &LabelSelector{
		MatchLabels: map[string]string{
			"key1": "",
			"key2": "",
		},
	}, `key1,key2`),
	Entry("should format matchExpressions with In operator", &LabelSelector{
		MatchExpressions: []LabelSelectorRequirement{
			{Key: env, Operator: In, Values: []string{prod, qa}},
			{Key: tier, Operator: In, Values: []string{front}},
		},
	}, `environment in (production,qa),tier in (frontend)`),
	Entry("should format matchExpressions with NotIn operator", &LabelSelector{
		MatchExpressions: []LabelSelectorRequirement{
			{Key: env, Operator: NotIn, Values: []string{prod, qa}},
			{Key: tier, Operator: NotIn, Values: []string{front}},
		},
	}, `environment notin (production,qa),tier notin (frontend)`),
	Entry("should format matchExpressions with Exists operator", &LabelSelector{
		MatchExpressions: []LabelSelectorRequirement{
			{Key: env, Operator: Exists, Values: []string{}},
			{Key: tier, Operator: Exists, Values: []string{}},
		},
	}, `environment,tier`),
	Entry("should format matchExpressions with NotExists operator", &LabelSelector{
		MatchExpressions: []LabelSelectorRequirement{
			{Key: env, Operator: DoesNotExists, Values: []string{}},
			{Key: tier, Operator: DoesNotExists, Values: []string{}},
		},
	}, `!environment,!tier`),
	Entry("should format matchExpressions with Exists,NotIn operator", &LabelSelector{
		MatchExpressions: []LabelSelectorRequirement{
			{Key: env, Operator: Exists, Values: []string{}},
			{Key: tier, Operator: NotIn, Values: []string{front}},
		},
	}, `environment,tier notin (frontend)`),
	Entry("should format matchExpressions with NotExists,In operator", &LabelSelector{
		MatchExpressions: []LabelSelectorRequirement{
			{Key: env, Operator: DoesNotExists, Values: []string{}},
			{Key: tier, Operator: In, Values: []string{front}},
		},
	}, `!environment,tier in (frontend)`),
	Entry("should format matchlabels with matchExpressions", &LabelSelector{
		MatchLabels: map[string]string{
			"key1": "",
			"key2": "value2",
		},
		MatchExpressions: []LabelSelectorRequirement{
			{Key: env, Operator: DoesNotExists, Values: []string{}},
			{Key: tier, Operator: In, Values: []string{front}},
		},
	}, `key1,key2=value2,!environment,tier in (frontend)`),
)
