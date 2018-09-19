package k8shandler

import (
	"html/template"
	"io"
	"strconv"
	"strings"
)

// esYmlStruct is used to render esYmlTmpl to a proper elasticsearch.yml format
type esYmlStruct struct {
	AllowClusterReader string
	KibanaIndexMode    string
	PathData           string
}

func renderEsYml(w io.Writer, allowClusterReader bool, kibanaIndexMode string, pathData string, insecureCluster bool) error {
	t := template.New("elasticsearch.yml")

	var config string
	if !insecureCluster {
		config = strings.Join([]string{esYmlTmpl, secureYmlTmpl}, "")
	} else {
		config = esYmlTmpl
	}

	t, err := t.Parse(config)
	if err != nil {
		return err
	}
	esy := esYmlStruct{
		AllowClusterReader: strconv.FormatBool(allowClusterReader),
		KibanaIndexMode:    kibanaIndexMode,
		PathData:           pathData,
	}

	return t.Execute(w, esy)
}

type log4j2PropertiesStruct struct {
	RootLogger string
}

func renderLog4j2Properties(w io.Writer, rootLogger string) error {
	t := template.New("log4j2.properties")
	t, err := t.Parse(log4j2PropertiesTmpl)
	if err != nil {
		return err
	}

	log4jProp := log4j2PropertiesStruct{
		RootLogger: rootLogger,
	}

	return t.Execute(w, log4jProp)
}
