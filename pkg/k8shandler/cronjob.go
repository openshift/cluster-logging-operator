package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler/factory"
	"k8s.io/apimachinery/pkg/api/errors"

	batch "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetCronJobList retrieves the list of cronjobs with a given selector in namespace
func (clusterRequest *ClusterLoggingRequest) GetCronJobList(selector map[string]string) (*batch.CronJobList, error) {
	list := &batch.CronJobList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: batch.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.List(
		selector,
		list,
	)

	return list, err
}

//RemoveCronJob with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveCronJob(cronjobName string) error {

	cronjob := factory.NewCronJob(
		cronjobName,
		clusterRequest.cluster.Namespace,
		cronjobName,
		cronjobName,
		batch.CronJobSpec{},
	)

	err := clusterRequest.Delete(cronjob)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v cronjob %v", cronjobName, err)
	}

	return nil
}
