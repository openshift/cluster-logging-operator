package elasticsearch

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/elasticsearch-operator/internal/constants"
	estypes "github.com/openshift/elasticsearch-operator/internal/types/elasticsearch"
	"github.com/openshift/elasticsearch-operator/internal/utils"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (ec *esClient) CreateIndexTemplate(name string, template *estypes.IndexTemplate) error {
	body, err := utils.ToJSON(template)
	if err != nil {
		return err
	}
	payload := &EsRequest{
		Method:      http.MethodPut,
		URI:         fmt.Sprintf("_template/%s", name),
		RequestBody: body,
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error != nil || (payload.StatusCode != 200 && payload.StatusCode != 201) {
		return ec.errorCtx().New("failed to create index template",
			"template", name,
			"response_status", payload.StatusCode,
			"response_body", payload.ResponseBody,
			"response_error", payload.Error,
		)
	}
	return nil
}

func (ec *esClient) DeleteIndexTemplate(name string) error {
	payload := &EsRequest{
		Method: http.MethodDelete,
		URI:    fmt.Sprintf("_template/%s", name),
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error == nil && (payload.StatusCode == 404 || payload.StatusCode < 300) {
		return nil
	}

	return ec.errorCtx().New("failed to delete index template",
		"template", name,
		"response_status", payload.StatusCode,
		"response_body", payload.ResponseBody,
		"response_error", payload.Error)
}

// ListTemplates returns a list of templates
func (ec *esClient) ListTemplates() (sets.String, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_template",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error != nil || payload.StatusCode != 200 {
		return nil, ec.errorCtx().New("failed to get list of index templates",
			"response_status", payload.StatusCode,
			"response_body", payload.ResponseBody,
			"response_error", payload.Error)
	}
	response := sets.NewString()
	for name := range payload.ResponseBody {
		response.Insert(name)
	}
	return response, nil
}

func (ec *esClient) GetIndexTemplates() (map[string]estypes.GetIndexTemplate, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    fmt.Sprintf("_template/common.*,%s-*", constants.OcpTemplatePrefix),
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	// unmarshal response body and return that
	templates := map[string]estypes.GetIndexTemplate{}
	err := json.Unmarshal([]byte(payload.RawResponseBody), &templates)
	if err != nil {
		return templates, fmt.Errorf("failed decoding raw response body into `map[string]estypes.GetIndexTemplate` for %s in namespace %s: %v", ec.cluster, ec.namespace, err)
	}

	return templates, payload.Error
}

func (ec *esClient) updateAllIndexTemplateReplicas(replicaCount int32) (bool, error) {
	// get the index template and then update the replica and put it
	indexTemplates, err := ec.GetIndexTemplates()
	if err != nil {
		return false, err
	}

	replicaString := fmt.Sprintf("%d", replicaCount)

	for templateName, template := range indexTemplates {

		currentReplicas := template.Settings.Index.NumberOfReplicas
		if currentReplicas != replicaString {
			template.Settings.Index.NumberOfReplicas = replicaString

			templateJSON, err := json.Marshal(template)
			if err != nil {
				return false, err
			}

			payload := &EsRequest{
				Method:      http.MethodPut,
				URI:         fmt.Sprintf("_template/%s", templateName),
				RequestBody: string(templateJSON),
			}

			ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

			acknowledged := false
			if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
				acknowledged = acknowledgedBool
			}

			if !(payload.StatusCode == 200 && acknowledged) {
				log.Error(payload.Error, "unable to update template", "cluster", ec.cluster, "namespace", ec.namespace, "template", templateName)
			}
		}
	}

	return true, nil
}

func (ec *esClient) UpdateTemplatePrimaryShards(shardCount int32) error {
	// get the index template and then update the shards and put it
	indexTemplates, err := ec.GetIndexTemplates()
	if err != nil {
		return err
	}

	shardString := fmt.Sprintf("%d", shardCount)

	for templateName, template := range indexTemplates {

		currentShards := template.Settings.Index.NumberOfShards
		if currentShards != shardString {
			template.Settings.Index.NumberOfShards = shardString

			templateJSON, err := json.Marshal(template)
			if err != nil {
				return err
			}

			payload := &EsRequest{
				Method:      http.MethodPut,
				URI:         fmt.Sprintf("_template/%s", templateName),
				RequestBody: string(templateJSON),
			}

			ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

			acknowledged := false
			if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
				acknowledged = acknowledgedBool
			}

			if !(payload.StatusCode == 200 && acknowledged) {
				log.Error(payload.Error, "unable to update template", "cluster", ec.cluster, "namespace", ec.namespace, "template", templateName)
			}
		}
	}

	return nil
}
