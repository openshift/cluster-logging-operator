package k8shandler

import (
	"html/template"
	"io"
)

// esYmlStruct is used to render esYmlTmpl to a proper elasticsearch.yml format
type esYmlStruct struct {
	KibanaIndexMode string
	EsUnicastHost   string
}

func renderEsYml(w io.Writer, kibanaIndexMode, esUnicastHost string) error {
	t := template.New("elasticsearch.yml")
	config := esYmlTmpl
	t, err := t.Parse(config)
	if err != nil {
		return err
	}
	esy := esYmlStruct{
		KibanaIndexMode: kibanaIndexMode,
		EsUnicastHost:   esUnicastHost,
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
