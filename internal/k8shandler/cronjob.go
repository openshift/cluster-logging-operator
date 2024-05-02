package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	batch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCronJob stubs an instance of a cronjob
func NewCronJob(cronjobName, namespace, loggingComponent, component string, cronjobSpec batch.CronJobSpec) *batch.CronJob {
	return &batch.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: batch.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronjobName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
		},
		Spec: cronjobSpec,
	}
}

// GetCronJobList retrieves the list of cronjobs with a given selector in namespace
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

// RemoveCronJob with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveCronJob(cronjobName string) error {

	cronjob := NewCronJob(
		cronjobName,
		clusterRequest.Cluster.Namespace,
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
