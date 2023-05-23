package clusterlogforwarder

import v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"

func Validate(clf v1.ClusterLogForwarder) error {
	for _, validate := range validations {
		if err := validate(clf); err != nil {
			return err
		}
	}
	return nil
}

// validations are the set of admission rules for validating
// a ClusterLogForwarder
var validations = []func(clf v1.ClusterLogForwarder) error{
	validateSingleton,
	validateJsonParsingToElasticsearch,
	validateUrlAccordingToTls,
	validateHttpContentTypeHeaders,
}
