package elasticsearch

import (
	"fmt"
	"net/http"

	"github.com/openshift/elasticsearch-operator/pkg/logger"
	estypes "github.com/openshift/elasticsearch-operator/pkg/types/elasticsearch"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (ec *esClient) CreateIndexTemplate(name string, template *estypes.IndexTemplate) error {
	body, err := utils.ToJson(template)
	if err != nil {
		return err
	}
	payload := &EsRequest{
		Method:      http.MethodPut,
		URI:         fmt.Sprintf("_template/%s", name),
		RequestBody: body,
	}

	logger.DebugObject("CreateIndexTemplate with payload: %s", template)
	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error != nil {
		return payload.Error
	}
	if payload.StatusCode != 200 && payload.StatusCode != 201 {
		return fmt.Errorf("There was an error creating index template %s. Error code: %v, %v", name, payload.StatusCode != 200, payload.ResponseBody)
	}
	return nil
}

func (ec *esClient) DeleteIndexTemplate(name string) error {
	payload := &EsRequest{
		Method: http.MethodDelete,
		URI:    fmt.Sprintf("_template/%s", name),
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error != nil {
		return payload.Error
	}
	if payload.StatusCode != 200 && payload.StatusCode != 404 {
		return fmt.Errorf("There was an error deleting template %s. Error code: %v", name, payload.StatusCode)
	}
	return nil
}

//ListTemplates returns a list of templates
func (ec *esClient) ListTemplates() (sets.String, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_template",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error != nil {
		return nil, payload.Error
	}
	if payload.StatusCode != 200 {
		return nil, fmt.Errorf("There was an error retrieving list of templates. Error code: %v, %v", payload.StatusCode != 200, payload.ResponseBody)
	}
	response := sets.NewString()
	for name := range payload.ResponseBody {
		response.Insert(name)
	}
	return response, nil
}

func (ec *esClient) GetIndexTemplates() (map[string]interface{}, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_template/common.*",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	return payload.ResponseBody, payload.Error
}
