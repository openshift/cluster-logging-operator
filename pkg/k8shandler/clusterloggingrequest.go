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
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLoggingRequest struct {
	Client  client.Client
	Cluster *logging.ClusterLogging

	// ForwarderRequest is a logforwarder instance
	ForwarderRequest *logging.ClusterLogForwarder

	// ForwarderSpec is the normalized and sanitized logforwarder spec
	ForwarderSpec logging.ClusterLogForwarderSpec
}

// TODO: determine if this is even necessary
func (clusterRequest *ClusterLoggingRequest) isManaged() bool {
	return clusterRequest.Cluster.Spec.ManagementState == logging.ManagementStateManaged
}

func (clusterRequest *ClusterLoggingRequest) Create(object runtime.Object) error {
	err := clusterRequest.Client.Create(context.TODO(), object)
	return err
}

//Update the runtime Object or return error
func (clusterRequest *ClusterLoggingRequest) Update(object runtime.Object) (err error) {
	if err = clusterRequest.Client.Update(context.TODO(), object); err != nil {
		log.Error(err, "Error updating ", object.GetObjectKind())
	}
	return err
}

//Update the runtime Object status or return error
func (clusterRequest *ClusterLoggingRequest) UpdateStatus(object runtime.Object) (err error) {
	if err = clusterRequest.Client.Status().Update(context.TODO(), object); err != nil {
		// making this debug because we should be throwing the returned error if we are never
		// able to update the status
		log.V(2).Error(err, "Error updating status")
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) Get(objectName string, object runtime.Object) error {
	namespacedName := types.NamespacedName{Name: objectName, Namespace: clusterRequest.Cluster.Namespace}

	log.V(3).Info("Getting object", "namespacedName", namespacedName, "object", object)

	return clusterRequest.Client.Get(context.TODO(), namespacedName, object)
}

func (clusterRequest *ClusterLoggingRequest) GetClusterResource(objectName string, object runtime.Object) error {
	namespacedName := types.NamespacedName{Name: objectName}
	log.V(3).Info("Getting ClusterResource object", "namespacedName", namespacedName, "object", object)
	err := clusterRequest.Client.Get(context.TODO(), namespacedName, object)
	log.V(3).Error(err, "Response")
	return err
}

func (clusterRequest *ClusterLoggingRequest) List(selector map[string]string, object runtime.Object) error {
	log.V(3).Info("Listing selector object", "selector", selector, "object", object)

	listOpts := []client.ListOption{
		client.InNamespace(clusterRequest.Cluster.Namespace),
		client.MatchingLabels(selector),
	}

	return clusterRequest.Client.List(
		context.TODO(),
		object,
		listOpts...,
	)
}

func (clusterRequest *ClusterLoggingRequest) Delete(object runtime.Object) error {
	log.V(3).Info("Deleting", "object", object)
	return clusterRequest.Client.Delete(context.TODO(), object)
}

func (clusterRequest *ClusterLoggingRequest) UpdateCondition(t logging.ConditionType, message string, reason logging.ConditionReason, status v1.ConditionStatus) error {
	if logging.SetCondition(&clusterRequest.Cluster.Status.Conditions, t, status, reason, message) {
		return clusterRequest.UpdateStatus(clusterRequest.Cluster)
	}
	return nil
}
