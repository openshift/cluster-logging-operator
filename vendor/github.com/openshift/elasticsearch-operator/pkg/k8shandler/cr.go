package k8shandler

import (
	"context"
	"fmt"

	loggingv1 "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetElasticsearchCR(c client.Client, ns string) (*loggingv1.Elasticsearch, error) {
	esl := &loggingv1.ElasticsearchList{}
	opts := &client.ListOptions{Namespace: ns}

	if err := c.List(context.TODO(), opts, esl); err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}

		return nil, fmt.Errorf("unable to get elasticsearch instance in %q: %w", ns, err)
	}

	if len(esl.Items) == 0 {
		return nil, fmt.Errorf("failed to find elasticsearch instance in %q: empty result set", ns)
	}

	return &esl.Items[0], nil
}
