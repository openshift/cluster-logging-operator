package helpers

import (
	"fmt"
	"sort"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Replacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")

func LabelName(name string) string {
	return strings.ToUpper(fmt.Sprintf("@%s", Replacer.Replace(name)))
}

func LabelNames(names []string) []string {
	asLabels := make([]string, len(names))
	for i, n := range names {
		asLabels[i] = LabelName(n)
	}
	return asLabels
}

func SourceTypeLabelName(name string) string {
	return strings.ToUpper(fmt.Sprintf("@_%s", Replacer.Replace(name)))
}

func LabelsKV(ls *loggingv1.LabelSelector) []string {
	m, _ := metav1.LabelSelectorAsMap(&metav1.LabelSelector{MatchLabels: ls.MatchLabels})
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	kv := make([]string, len(m))
	for i, k := range keys {
		kv[i] = fmt.Sprintf("%s:%s", k, m[k])
	}
	return kv
}

func StoreID(prefix, name, suffix string) string {
	return strings.ToLower(fmt.Sprintf("%v%v%v", prefix, Replacer.Replace(name), suffix))
}
