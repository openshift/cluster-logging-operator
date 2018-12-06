package k8shandler

import (
	"html/template"
	"io"
)

// esYmlStruct is used to render esYmlTmpl to a proper elasticsearch.yml format
type esYmlStruct struct {
	KibanaIndexMode       string
	EsUnicastHost         string
	NodeQuorum            string
	RecoverExpectedShards string
}

func renderEsYml(w io.Writer, kibanaIndexMode, esUnicastHost, nodeQuorum, recoverExpectedShards string) error {
	t := template.New("elasticsearch.yml")
	config := esYmlTmpl
	t, err := t.Parse(config)
	if err != nil {
		return err
	}
	esy := esYmlStruct{
		KibanaIndexMode:       kibanaIndexMode,
		EsUnicastHost:         esUnicastHost,
		NodeQuorum:            nodeQuorum,
		RecoverExpectedShards: recoverExpectedShards,
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

type indexSettingsStruct struct {
	PrimaryShards string
	ReplicaShards string
}

func renderIndexSettings(w io.Writer, primaryShardsCount, replicaShardsCount string) error {
	t := template.New("index_settings")
	t, err := t.Parse(indexSettingsTmpl)
	if err != nil {
		return err
	}

	indexSettings := indexSettingsStruct{
		PrimaryShards: primaryShardsCount,
		ReplicaShards: replicaShardsCount,
	}

	return t.Execute(w, indexSettings)
}
