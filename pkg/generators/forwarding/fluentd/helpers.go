package fluentd

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"strings"
	"text/template"
)

var (
	helperRegistry = &template.FuncMap{
		"applicationTag":      applicationTag,
		"labelName":           labelName,
		"sourceTypelabelName": sourceTypeLabelName,
		"routeMapValues":      routeMapValues,
	}
)

func applicationTag(namespace string) string {
	if "" == namespace {
		return "**"
	}
	return strings.ToLower(fmt.Sprintf("kubernetes.**_%s_**", namespace))
}

func labelName(name string) string {
	return strings.ToUpper(fmt.Sprintf("@%s", replacer.Replace(name)))
}

func sourceTypeLabelName(name string) string {
	return strings.ToUpper(fmt.Sprintf("@_%s", replacer.Replace(name)))
}

func routeMapValues(routeMap logging.RouteMap, key string) []string {
	if values, found := routeMap[key]; found {
		return values.List()
	}
	return []string{}
}
