package helpers

import (
	. "github.com/onsi/ginkgo"

	"encoding/json"
	"fmt"
)

func NewFakeElasticsearchChatter(responses map[string]FakeElasticsearchResponse) *FakeElasticsearchChatter {
	return &FakeElasticsearchChatter{
		Requests:  map[string]string{},
		Responses: responses,
	}
}

type FakeElasticsearchChatter struct {
	Requests  map[string]string
	Responses map[string]FakeElasticsearchResponse
}

type FakeElasticsearchResponse struct {
	Error      error
	StatusCode int
	Body       string
}

func (chat *FakeElasticsearchChatter) GetRequest(key string) (string, bool) {
	request, found := chat.Requests[key]
	return request, found
}

func (chat *FakeElasticsearchChatter) GetResponse(key string) (FakeElasticsearchResponse, bool) {
	response, found := chat.Responses[key]
	return response, found
}

func (response *FakeElasticsearchResponse) BodyAsResponseBody() map[string]interface{} {
	body := &map[string]interface{}{}
	if err := json.Unmarshal([]byte(response.Body), body); err != nil {
		Fail(fmt.Sprintf("Unable to convert to response body %q: %v", response.Body, err))
	}
	return *body
}
