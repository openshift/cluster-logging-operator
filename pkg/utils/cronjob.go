package utils

import (
	"fmt"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	batch "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewCronJob stubs an instance of a cronjob
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

//GetCronJobList retrieves the list of cronjobs with a given selector in namespace
func GetCronJobList(namespace, selector string) (*batch.CronJobList, error) {
	list := &batch.CronJobList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: batch.SchemeGroupVersion.String(),
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}

//RemoveCronJob with given name and namespace
func RemoveCronJob(namespace, cronjobName string) error {

	cronjob := NewCronJob(
		cronjobName,
		namespace,
		cronjobName,
		cronjobName,
		batch.CronJobSpec{},
	)

	err := sdk.Delete(cronjob)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v cronjob %v", cronjobName, err)
	}

	return nil
}
