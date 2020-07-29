package factory

import (
	batch "k8s.io/api/batch/v1beta1"
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
