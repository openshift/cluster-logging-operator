package k8shandler

import (
	"bytes"
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"

	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateStatus(logging *v1alpha1.ClusterLogging) (err error) {

	changed := false

	if logging.Spec.LogStore.Type == "elasticsearch" {
		elasticsearchStatus, err := getElasticsearchStatus(logging.Namespace)

		if err != nil {
			logrus.Fatalf("Failed to get status for Elasticsearch")
		}

		if !reflect.DeepEqual(elasticsearchStatus, logging.Status.LogStore.ElasticsearchStatus) {
			logrus.Infof("Updating status of Elasticsearch")
			logging.Status.LogStore.ElasticsearchStatus = elasticsearchStatus
			changed = true
		}
	}

	if logging.Spec.Visualization.Type == "kibana" {
		kibanaStatus, err := getKibanaStatus(logging.Namespace)

		if err != nil {
			logrus.Fatalf("Failed to get status for Elasticsearch")
		}

		if !reflect.DeepEqual(kibanaStatus, logging.Status.Visualization.KibanaStatus) {
			logrus.Infof("Updating status of Kibana")
			logging.Status.Visualization.KibanaStatus = kibanaStatus
			changed = true
		}
	}

	if logging.Spec.Curation.Type == "curator" {
		curatorStatus, err := getCuratorStatus(logging.Namespace)

		if err != nil {
			logrus.Fatalf("Failed to get status for Curator")
		}

		if !reflect.DeepEqual(curatorStatus, logging.Status.Curation.CuratorStatus) {
			logrus.Infof("Updating status of Curator")
			logging.Status.Curation.CuratorStatus = curatorStatus
			changed = true
		}
	}

	if logging.Spec.Collection.Type == "fluentd" {
		fluentdStatus, err := getFluentdCollectorStatus(logging.Namespace)

		if err != nil {
			logrus.Fatalf("Failed to get status of Fluentd")
		}

		if !reflect.DeepEqual(fluentdStatus, logging.Status.Collection.FluentdStatus) {
			logrus.Infof("Updating status of Fluentd")
			logging.Status.Collection.FluentdStatus = fluentdStatus
			changed = true
		}
	}

	if changed {
		err = sdk.Update(logging)

		// TODO: check for invalid object error (the object was deleted on us...)

		if err != nil {
			logrus.Fatalf("Failed to update Cluster Logging status: %v", err)
		}
	}

	return nil
}

func getCuratorStatus(namespace string) (statuses []v1alpha1.CuratorStatus, err error) {

	status := []v1alpha1.CuratorStatus{}

	curatorCronJobList, err := utils.GetCronJobList(namespace, "logging-infra=curator")

	for _, cronjob := range curatorCronJobList.Items {

		curatorStatus := v1alpha1.CuratorStatus{
			CronJob:   cronjob.Name,
			Schedule:  cronjob.Spec.Schedule,
			Suspended: *cronjob.Spec.Suspend,
		}

		status = append(status, curatorStatus)
	}

	return status, nil
}

func getFluentdCollectorStatus(namespace string) (status v1alpha1.FluentdCollectorStatus, err error) {

	fluentdDaemonsetList, err := utils.GetDaemonSetList(namespace, "logging-infra=fluentd")
	fluentdStatus := v1alpha1.FluentdCollectorStatus{}

	if len(fluentdDaemonsetList.Items) != 0 {
		daemonset := fluentdDaemonsetList.Items[0]

		fluentdStatus.DaemonSet = daemonset.Name

		// use map to represent {pod: node}
		podList, _ := utils.GetPodList(namespace, "logging-infra=fluentd")
		pods := make(map[string]string)
		for _, pod := range podList.Items {
			pods[pod.Name] = pod.Spec.NodeName
		}
		fluentdStatus.Pods = pods
	}

	return fluentdStatus, nil
}

func getKibanaStatus(namespace string) (statuses []v1alpha1.KibanaStatus, err error) {

	status := []v1alpha1.KibanaStatus{}

	kibanaDeploymentList, err := utils.GetDeploymentList(namespace, "logging-infra=kibana")

	for _, deployment := range kibanaDeploymentList.Items {

		var selectorValue bytes.Buffer
		selectorValue.WriteString("component=")
		selectorValue.WriteString(deployment.Name)

		kibanaStatus := v1alpha1.KibanaStatus{
			Deployment: deployment.Name,
			Replicas:   *deployment.Spec.Replicas,
		}

		replicaSetList, _ := utils.GetReplicaSetList(namespace, selectorValue.String())
		replicaNames := []string{}
		for _, replicaSet := range replicaSetList.Items {
			replicaNames = append(replicaNames, replicaSet.Name)
		}
		kibanaStatus.ReplicaSets = replicaNames

		podList, _ := utils.GetPodList(namespace, selectorValue.String())
		podNames := []string{}
		for _, pod := range podList.Items {
			podNames = append(podNames, pod.Name)
		}
		kibanaStatus.Pods = podNames

		status = append(status, kibanaStatus)
	}

	return status, nil
}

func getElasticsearchStatus(namespace string) (statuses []v1alpha1.ElasticsearchStatus, err error) {

	// we can scrape the status provided by the elasticsearch-operator
	// get list of elasticsearch objects
	esList := &elasticsearch.ElasticsearchList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: "elasticsearch.redhat.com/v1alpha1",
		},
	}

	err = sdk.List(namespace, esList)
	status := []v1alpha1.ElasticsearchStatus{}

	if err != nil {
		logrus.Fatalf("Unable to get Elasticsearches: %v", err)
	}

	if len(esList.Items) != 0 {
		for _, node := range esList.Items {

			nodeStatus := v1alpha1.ElasticsearchStatus{
				ClusterName:  node.Name,
				Replicas:     node.Spec.Nodes[0].Replicas,
				Deployments:  getDeploymentNames(node.Status),
				ReplicaSets:  getReplicaSetNames(node.Status),
				StatefulSets: getStatefulSetNames(node.Status),
				Pods:         getPodNames(node.Status),
			}

			status = append(status, nodeStatus)
		}
	}

	return status, nil
}

func getPodNames(node elasticsearch.ElasticsearchStatus) []string {

	podNames := []string{}

	for _, nodeStatus := range node.Nodes {
		podNames = append(podNames, nodeStatus.PodName)
	}

	return podNames
}

func getDeploymentNames(node elasticsearch.ElasticsearchStatus) []string {

	deploymentNames := []string{}

	for _, nodeStatus := range node.Nodes {
		deploymentNames = append(deploymentNames, nodeStatus.DeploymentName)
	}

	return deploymentNames
}

func getReplicaSetNames(node elasticsearch.ElasticsearchStatus) []string {

	replicasetNames := []string{}

	for _, nodeStatus := range node.Nodes {
		replicasetNames = append(replicasetNames, nodeStatus.ReplicaSetName)
	}

	return replicasetNames
}

func getStatefulSetNames(node elasticsearch.ElasticsearchStatus) []string {

	statefulsetNames := []string{}

	for _, nodeStatus := range node.Nodes {
		statefulsetNames = append(statefulsetNames, nodeStatus.StatefulSetName)
	}

	return statefulsetNames
}
