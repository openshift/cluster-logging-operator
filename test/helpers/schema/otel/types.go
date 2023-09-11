package otel

import (
	"encoding/json"
)

type AllOTELLog struct {
	ContainerLog `json:",inline,omitempty"`
}

type OTELLogs []AllOTELLog

type ContainerLog struct {
	ApplicationLog `json:",inline,omitempty"`
}

type ApplicationLog struct {
	Resources      Resources `json:"resources,omitempty"`
	SeverityNumber int       `json:"severityNumber,omitempty"`
	SeverityText   string    `json:"severityText,omitempty"`
	TimeUnixNano   int64     `json:"timeUnixNano,omitempty"`
}

type Resources struct {
	Attributes map[string]string `json:"attributes,omitempty"`
	Container  Container         `json:"container,omitempty"`
	Host       Host              `json:"host,omitempty"`
	K8s        K8s               `json:"k8s,omitempty"`
}

type Container struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Image Image  `json:"image,omitempty"`
}

type Image struct {
	Name string `json:"name,omitempty"`
	Tag  string `json:"tag,omitempty"`
}

type Host struct {
	Name string `json:"name,omitempty"`
}

type K8s struct {
	Namespace Namespace `json:"namespace,omitempty"`
	Pod       Pod       `json:"pod,omitempty"`
	Logs      Logs      `json:"logs,omitempty"`
}

type Namespace struct {
	Name   string            `json:"name,omitempty"`
	Id     string            `json:"id,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type Pod struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Ip          string            `json:"ip,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Name        string            `json:"name,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	UID         string            `json:"uid,omitempty"`
}

type Logs struct {
	File File `json:"file,omitempty"`
}

type File struct {
	Path string `json:"path,omitempty"`
}

func ParseLogs(in string) (OTELLogs, error) {
	logs := OTELLogs{}
	if in == "" {
		return logs, nil
	}

	err := json.Unmarshal([]byte(in), &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
