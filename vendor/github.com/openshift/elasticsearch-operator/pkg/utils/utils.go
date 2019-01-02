package utils

import (
	"io/ioutil"
	"os"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

func GetFileContents(filePath string) []byte {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("Unable to read file to get contents: %v", err)
		return nil
	}

	return contents
}

func Secret(secretName string, namespace string, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: "Opaque",
		Data: data,
	}
}

func LookupEnvWithDefault(envName, defaultValue string) string {
	if value, ok := os.LookupEnv(envName); ok {
		return value
	}
	return defaultValue
}

func GetESNodeCondition(status *api.ElasticsearchStatus, conditionType api.ClusterConditionType) (int, *api.ClusterCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}

func UpdateESNodeCondition(status *api.ElasticsearchStatus, condition *api.ClusterCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this node condition.
	conditionIndex, oldCondition := GetESNodeCondition(status, condition.Type)

	if oldCondition == nil {
		// We are adding new node condition.
		status.Conditions = append(status.Conditions, *condition)
		return true
	}
	// We are updating an existing condition, so we need to check if it has changed.
	if condition.Status == oldCondition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	status.Conditions[conditionIndex] = *condition
	// Return true if one of the fields have changed.
	return !isEqual
}

func UpdateConditionWithRetry(dpl *api.Elasticsearch, value api.ConditionStatus,
	executeUpdateCondition func(*api.ElasticsearchStatus, api.ConditionStatus) bool) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := sdk.Get(dpl); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch %v: %v", dpl.Name, getErr)
			return getErr
		}

		executeUpdateCondition(&dpl.Status, value)

		if updateErr := sdk.Update(dpl); updateErr != nil {
			logrus.Debugf("Failed to update Elasticsearch %v status: %v", dpl.Name, updateErr)
			return updateErr
		}
		return nil
	})
	return retryErr
}

func UpdateUpdatingSettingsCondition(status *api.ElasticsearchStatus, value api.ConditionStatus) bool {
	var message string
	if value == api.ConditionTrue {
		message = "Config Map is different"
	} else {
		message = "Config Map is up to date"
	}
	return UpdateESNodeCondition(status, &api.ClusterCondition{
		Type:    api.UpdatingSettings,
		Status:  value,
		Reason:  "ConfigChange",
		Message: message,
	})
}

func UpdateScalingUpCondition(status *api.ElasticsearchStatus, value api.ConditionStatus) bool {
	return UpdateESNodeCondition(status, &api.ClusterCondition{
		Type:   api.ScalingUp,
		Status: value,
	})
}

func UpdateScalingDownCondition(status *api.ElasticsearchStatus, value api.ConditionStatus) bool {
	return UpdateESNodeCondition(status, &api.ClusterCondition{
		Type:   api.ScalingDown,
		Status: value,
	})
}

func UpdateRestartingCondition(status *api.ElasticsearchStatus, value api.ConditionStatus) bool {
	return UpdateESNodeCondition(status, &api.ClusterCondition{
		Type:   api.Restarting,
		Status: value,
	})
}

func IsUpdatingSettings(status *api.ElasticsearchStatus) bool {
	_, settingsUpdateCondition := GetESNodeCondition(status, api.UpdatingSettings)
	if settingsUpdateCondition != nil && settingsUpdateCondition.Status == api.ConditionTrue {
		return true
	}
	return false
}

func IsClusterScalingUp(status *api.ElasticsearchStatus) bool {
	_, scaleUpCondition := GetESNodeCondition(status, api.ScalingUp)
	if scaleUpCondition != nil && scaleUpCondition.Status == api.ConditionTrue {
		return true
	}
	return false
}

func IsClusterScalingDown(status *api.ElasticsearchStatus) bool {
	_, scaleDownCondition := GetESNodeCondition(status, api.ScalingDown)
	if scaleDownCondition != nil && scaleDownCondition.Status == api.ConditionTrue {
		return true
	}
	return false
}

func IsRestarting(status *api.ElasticsearchStatus) bool {
	_, restartingCondition := GetESNodeCondition(status, api.Restarting)
	if restartingCondition != nil && restartingCondition.Status == api.ConditionTrue {
		return true
	}
	return false
}

func UpdateNodeUpgradeStatusWithRetry(dpl *api.Elasticsearch, deployName string, value *api.ElasticsearchNodeUpgradeStatus) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := sdk.Get(dpl); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch %v: %v", dpl.Name, getErr)
			return getErr
		}

		for i, node := range dpl.Status.Nodes {
			if node.DeploymentName == deployName {
				dpl.Status.Nodes[i].UpgradeStatus = *value
			}
		}

		if updateErr := sdk.Update(dpl); updateErr != nil {
			logrus.Debugf("Failed to update Elasticsearch %v status: %v", dpl.Name, updateErr)
			return updateErr
		}
		return nil
	})
	return retryErr
}

func NodeRestarting() *api.ElasticsearchNodeUpgradeStatus {
	return &api.ElasticsearchNodeUpgradeStatus{
		UnderUpgrade: api.UnderUpgradeTrue,
		UpgradePhase: api.NodeRestarting,
	}
}

func NodeRecoveringData() *api.ElasticsearchNodeUpgradeStatus {
	return &api.ElasticsearchNodeUpgradeStatus{
		UnderUpgrade: api.UnderUpgradeTrue,
		UpgradePhase: api.RecoveringData,
	}
}

func NodeControllerUpdated() *api.ElasticsearchNodeUpgradeStatus {
	return &api.ElasticsearchNodeUpgradeStatus{
		UnderUpgrade: api.UnderUpgradeTrue,
		UpgradePhase: api.ControllerUpdated,
	}
}

func NodeNormalOperation() *api.ElasticsearchNodeUpgradeStatus {
	return &api.ElasticsearchNodeUpgradeStatus{
		UnderUpgrade: api.UnderUpgradeFalse,
	}
}
