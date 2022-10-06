package clusterlogforwarder

import (
	"fmt"
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func validateSingleton(clf v1.ClusterLogForwarder) error {
	if clf.Name != constants.SingletonName {
		return fmt.Errorf("Invalid name %q, singleton instance must be named %q", clf.Name, constants.SingletonName)
	}
	return nil
}
