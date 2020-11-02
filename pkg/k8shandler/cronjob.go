// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

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
