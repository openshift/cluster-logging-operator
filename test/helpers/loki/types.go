package loki

import (
	"encoding/json"
	"fmt"
)

type stream struct {
	ContainerID      string `json:"container_id,omitempty"`
	ContainerImage   string `json:"container_image,omitempty"`
	ContainerImageID string `json:"container_image_id,omitempty"`
	MasterURL        string `json:"master_url,omitempty"`
	NamespaceID      string `json:"namespace_id,omitempty"`
	NamespaceName    string `json:"namespace_name,omitempty"`
	IndexName        string `json:"index_name,omitempty"`
	PodName          string `json:"pod_name,omitempty"`
	Hostname         string `json:"hostname,omitempty"`
}

type result struct {
	Stream *stream    `json:"stream,omitempty"`
	Values [][]string `json:"values,omitempty"`
}

type data struct {
	ResultType string   `json:"resultType"`
	Result     []result `json:"result"`
}

type Response struct {
	Status string `json:"status"`
	Data   *data  `json:"data"`
}

func Parse(in string) (Response, error) {
	res := Response{}
	if in == "" {
		return res, nil
	}

	err := json.Unmarshal([]byte(in), &res)
	if err != nil {
		return res, fmt.Errorf("%s: %s", err, in)
	}

	return res, nil
}

func (r *Response) ToString() string {
	s, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(s)
}

func (r *Response) NonEmpty() bool {
	if r.Data == nil {
		return false
	}

	if r.Data.Result == nil {
		return false
	}

	return len(r.Data.Result) > 0
}
