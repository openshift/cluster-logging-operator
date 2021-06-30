package fluentd

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
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

func (conf *outputLabelConf) HasCABundle() bool {
	if conf.Name == logging.OutputNameDefault {
		return true
	}
	if conf.Secret == nil {
		return false
	}

	if _, ok := conf.Secret.Data[constants.TrustedCABundleKey]; !ok {
		return false
	}
	return true
}

func (conf *outputLabelConf) HasTLSKeyAndCrt() bool {
	if conf.Name == logging.OutputNameDefault {
		return true
	}
	if conf.Secret == nil {
		return false
	}

	if _, ok := conf.Secret.Data[constants.ClientCertKey]; !ok {
		return false
	}
	if _, ok := conf.Secret.Data[constants.ClientPrivateKey]; !ok {
		return false
	}
	return true
}
