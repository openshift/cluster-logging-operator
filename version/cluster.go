package version

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/hostedcontrolplane"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	operatorConditionName = "OPERATOR_CONDITION_NAME"
)

var (
	clusterVersion string
	clusterID      string
)

// OperatorVersion get clo operator version from OPERATOR_CONDITION_NAME ENV variable
func OperatorVersion() (string, error) {
	operatorVersion, found := os.LookupEnv(operatorConditionName)
	if !found {
		return "", fmt.Errorf("%s must be set", operatorConditionName)
	}
	return operatorVersion, nil
}

// ClusterVersion retrieves the ClusterVersion spec
func ClusterVersion(k8client client.Reader) (string, string, error) {
	if clusterVersion == "" && clusterID == "" {
		proto := &configv1.ClusterVersion{}
		key := client.ObjectKey{Name: "version"}
		if err := k8client.Get(context.TODO(), key, proto); err != nil {
			return "", "", err
		}
		clusterVersion = proto.Status.Desired.Version
		clusterID = string(proto.Spec.ClusterID)
	}
	return clusterVersion, clusterID, nil
}

// HostedClusterVersion retrieves the version info of the hosted cluster or the clustser ID where the operator is deployed
// upon error
func HostedClusterVersion(ctx context.Context, k8client client.Reader, namespace string) (version, id string) {
	version, id, err := ClusterVersion(k8client)
	if err != nil {
		log.V(0).Error(err, "Unable to retrieve the cluster version")
	}
	// If reconciling in a hosted control plane namespace, use the guest cluster version/id
	// provided by the hostedcontrolplane resource.
	if info := hostedcontrolplane.GetVersionID(ctx, k8client, namespace); info != nil {
		return info.Version, info.ID
	}
	// Otherwise use the cluster-ID established at start-up.
	return version, id
}
