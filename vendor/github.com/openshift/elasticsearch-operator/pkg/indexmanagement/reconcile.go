package indexmanagement

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/openshift/elasticsearch-operator/pkg/logger"
	k8s "github.com/openshift/elasticsearch-operator/pkg/types/k8s"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/openshift/elasticsearch-operator/pkg/utils/comparators"
)

const (
	indexManagmentNamePrefix = "ocp-index-mgm"
	indexManagementConfigmap = "indexmanagement-scripts"
	defaultShardSize         = int32(40)
)

var (
	defaultCPURequest      = resource.MustParse("100m")
	defaultMemoryRequest   = resource.MustParse("32Mi")
	jobHistoryLimitFailed  = utils.GetInt32(2)
	jobHistoryLimitSuccess = utils.GetInt32(1)

	millisPerSecond = uint64(1000)
	millisPerMinute = uint64(60 * millisPerSecond)
	millisPerHour   = uint64(millisPerMinute * 60)
	millisPerDay    = uint64(millisPerHour * 24)
	millisPerWeek   = uint64(millisPerDay * 7)

	//fullExecMode 0777
	fullExecMode = utils.GetInt32(int32(511))

	imLabels = map[string]string{
		"provider":      "openshift",
		"component":     "indexManagement",
		"logging-infra": "indexManagement",
	}
)

type rolloverConditions struct {
	MaxAge  string `json:"max_age,omitempty"`
	MaxDocs int32  `json:"max_docs,omitempty"`
	MaxSize string `json:"max_size,omitempty"`
}

func RemoveCronJobsForMappings(apiclient client.Client, cluster *apis.Elasticsearch, mappings []apis.IndexManagementPolicyMappingSpec, policies apis.PolicyMap) error {
	expected := sets.NewString()
	for _, mapping := range mappings {
		policy := policies[mapping.PolicyRef]
		if policy.Phases.Hot != nil {
			expected.Insert(fmt.Sprintf("%s-rollover-%s", indexManagmentNamePrefix, mapping.Name))
		}
		if policy.Phases.Delete != nil {
			expected.Insert(fmt.Sprintf("%s-delete-%s", indexManagmentNamePrefix, mapping.Name))
		}
	}
	logger.Debugf("Expecting to have cronjobs in %s: %v", cluster.Namespace, expected.List())
	selector := labels.NewSelector()
	for k, v := range imLabels {
		req, _ := labels.NewRequirement(k, selection.Equals, []string{v})
		selector.Add(*req)
	}
	cronList := &batch.CronJobList{}
	if err := apiclient.List(context.TODO(), &client.ListOptions{Namespace: cluster.Namespace, LabelSelector: selector}, cronList); err != nil {
		return err
	}
	existing := sets.NewString()
	for _, cron := range cronList.Items {
		existing.Insert(cron.Name)
	}
	difference := existing.Difference(expected)
	logger.Debugf("Removing cronjobs in %s: %v", cluster.Namespace, difference.List())
	for _, name := range difference.List() {
		cronjob := &batch.CronJob{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CronJob",
				APIVersion: batch.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: cluster.Namespace,
			},
		}
		err := apiclient.Delete(context.TODO(), cronjob)
		if err != nil && !errors.IsNotFound(err) {
			logger.Errorf("Failure culling %s/%s cronjob %v", cluster.Namespace, name, err)
		}
	}
	return nil
}

func ReconcileCurationConfigmap(apiclient client.Client, cluster *apis.Elasticsearch) error {
	data := scriptMap
	desired := k8s.NewConfigMap(indexManagementConfigmap, cluster.Namespace, imLabels, data)
	cluster.AddOwnerRefTo(desired)
	err := apiclient.Create(context.TODO(), desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Error creating configmap for cluster %s: %v", cluster.Name, err)
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			current := &v1.ConfigMap{}
			retryError := apiclient.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
			if retryError != nil {
				return fmt.Errorf("Unable to get configmap %s/%s during reconciliation: %v", desired.Namespace, desired.Name, retryError)
			}
			if !reflect.DeepEqual(desired.Data, current.Data) {
				logger.Debugf("Updating configmap %s/%s", current.Namespace, current.Name)
				current.Data = desired.Data
				return apiclient.Update(context.TODO(), current)
			}
			return nil
		})
	}
	return err
}

func ReconcileRolloverCronjob(apiclient client.Client, cluster *apis.Elasticsearch, policy apis.IndexManagementPolicySpec, mapping apis.IndexManagementPolicyMappingSpec, primaryShards int32) error {
	if policy.Phases.Hot == nil {
		logger.Infof("Skipping rollover cronjob for policymapping %q; hot phase not defined", mapping.Name)
		return nil
	}
	schedule, err := crontabScheduleFor(policy.PollInterval)
	if err != nil {
		return err
	}
	conditions := calculateConditions(policy, primaryShards)
	name := fmt.Sprintf("%s-rollover-%s", indexManagmentNamePrefix, mapping.Name)
	payload, err := utils.ToJson(map[string]rolloverConditions{"conditions": conditions})
	if err != nil {
		return fmt.Errorf("There was an error serializing the rollover conditions to JSON: %v", err)
	}
	envvars := []core.EnvVar{
		{Name: "PAYLOAD", Value: base64.StdEncoding.EncodeToString([]byte(payload))},
		{Name: "POLICY_MAPPING", Value: mapping.Name},
	}
	fnContainerHandler := func(container *core.Container) {
		container.Command = []string{"bash"}
		container.Args = []string{
			"-c",
			"/tmp/scripts/rollover",
		}
	}
	desired := newCronJob(cluster.Name, cluster.Spec.Spec.Image, cluster.Namespace, name, schedule, cluster.Spec.Spec.NodeSelector, cluster.Spec.Spec.Tolerations, envvars, fnContainerHandler)

	cluster.AddOwnerRefTo(desired)
	return reconcileCronJob(apiclient, cluster, desired, areCronJobsSame)
}

func ReconcileCurationCronjob(apiclient client.Client, cluster *apis.Elasticsearch, policy apis.IndexManagementPolicySpec, mapping apis.IndexManagementPolicyMappingSpec, primaryShards int32) error {
	if policy.Phases.Delete == nil {
		logger.Infof("Skipping curation cronjob for policymapping %q; delete phase not defined", mapping.Name)
		return nil
	}
	schedule, err := crontabScheduleFor(policy.PollInterval)
	if err != nil {
		return err
	}
	minAgeMillis, err := calculateMillisForTimeUnit(policy.Phases.Delete.MinAge)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s-delete-%s", indexManagmentNamePrefix, mapping.Name)
	envvars := []core.EnvVar{
		{Name: "ALIAS", Value: mapping.Name},
		{Name: "MIN_AGE", Value: strconv.FormatUint(minAgeMillis, 10)},
	}
	fnContainerHandler := func(container *core.Container) {
		container.Command = []string{"bash"}
		container.Args = []string{
			"-c",
			"/tmp/scripts/delete",
		}
	}
	desired := newCronJob(cluster.Name, cluster.Spec.Spec.Image, cluster.Namespace, name, schedule, cluster.Spec.Spec.NodeSelector, cluster.Spec.Spec.Tolerations, envvars, fnContainerHandler)

	cluster.AddOwnerRefTo(desired)
	return reconcileCronJob(apiclient, cluster, desired, areCronJobsSame)
}

func reconcileCronJob(apiclient client.Client, cluster *apis.Elasticsearch, desired *batch.CronJob, fnAreCronJobsSame func(lhs, rhs *batch.CronJob) bool) error {
	err := apiclient.Create(context.TODO(), desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Error creating cronjob for cluster %s: %v", cluster.Name, err)
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			current := &batch.CronJob{}
			retryError := apiclient.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
			if retryError != nil {
				return fmt.Errorf("Unable to get cronjob %s/%s during reconciliation: %v", desired.Namespace, desired.Name, retryError)
			}
			if !fnAreCronJobsSame(current, desired) {
				current.Spec = desired.Spec
				return apiclient.Update(context.TODO(), current)
			}
			return nil
		})
	}
	return err
}

func areCronJobsSame(lhs, rhs *batch.CronJob) bool {
	logger.Debugf("Evaluating cronjob '%s/%s' ...", lhs.Namespace, lhs.Name)
	if len(lhs.Spec.JobTemplate.Spec.Template.Spec.Containers) != len(lhs.Spec.JobTemplate.Spec.Template.Spec.Containers) {
		logger.Debugf("Container lengths are different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
		return false
	}
	if !comparators.AreStringMapsSame(lhs.Spec.JobTemplate.Spec.Template.Spec.NodeSelector, rhs.Spec.JobTemplate.Spec.Template.Spec.NodeSelector) {
		logger.Debugf("NodeSelector is different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
		return false
	}

	if !comparators.AreTolerationsSame(lhs.Spec.JobTemplate.Spec.Template.Spec.Tolerations, rhs.Spec.JobTemplate.Spec.Template.Spec.Tolerations) {
		logger.Debugf("Tolerations are different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
		return false
	}
	if lhs.Spec.Schedule != rhs.Spec.Schedule {
		logger.Debugf("Schedule is different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
		lhs.Spec.Schedule = rhs.Spec.Schedule
		return false
	}
	if lhs.Spec.Suspend != nil && rhs.Spec.Suspend != nil && *lhs.Spec.Suspend != *rhs.Spec.Suspend {
		logger.Debugf("Suspend is different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
		return false
	}

	for i, container := range lhs.Spec.JobTemplate.Spec.Template.Spec.Containers {
		logger.Debugf("Evaluating cronjob container %q ...", container.Name)
		other := rhs.Spec.JobTemplate.Spec.Template.Spec.Containers[i]
		if container.Name != other.Name {
			logger.Debugf("Container name is different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
			return false
		}
		if container.Image != other.Image {
			logger.Debugf("Container image is different between current and desired for %s/%s: %q != %q", lhs.Namespace, lhs.Name, container.Image, other.Image)
			return false
		}

		if !reflect.DeepEqual(container.Command, other.Command) {
			logger.Debugf("Container command is different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
			return false
		}
		if !reflect.DeepEqual(container.Args, other.Args) {
			logger.Debugf("Container command args is different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
			return false
		}

		if !comparators.AreResourceRequementsSame(container.Resources, other.Resources) {
			logger.Debugf("Container resources are different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
			return false
		}

		if !comparators.EnvValueEqual(container.Env, other.Env) {
			logger.Debugf("Container EnvVars are different between current and desired for %s/%s", lhs.Namespace, lhs.Name)
			return false
		}

	}
	logger.Debug("The current and desired cronjobs are the same")
	return true
}

func newCronJob(clusterName, image, namespace, name, schedule string, nodeSelector map[string]string, tolerations []core.Toleration, envvars []core.EnvVar, fnContainerHander func(*core.Container)) *batch.CronJob {
	container := core.Container{
		Name:            "indexmanagement",
		Image:           image,
		ImagePullPolicy: core.PullIfNotPresent,
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultMemoryRequest,
				v1.ResourceCPU:    defaultCPURequest,
			},
		},
		Env: []core.EnvVar{
			{Name: "ES_SERVICE", Value: fmt.Sprintf("https://%s:9200", clusterName)},
		},
	}
	container.Env = append(container.Env, envvars...)
	fnContainerHander(&container)

	container.VolumeMounts = []v1.VolumeMount{
		{Name: "certs", ReadOnly: true, MountPath: "/etc/indexmanagement/keys"},
		{Name: "scripts", ReadOnly: false, MountPath: "/tmp/scripts"},
	}
	podSpec := core.PodSpec{
		ServiceAccountName: clusterName,
		Containers:         []v1.Container{container},
		Volumes: []v1.Volume{
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: clusterName}}},
			{Name: "scripts", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: indexManagementConfigmap}, DefaultMode: fullExecMode}}},
		},
		NodeSelector:                  utils.EnsureLinuxNodeSelector(nodeSelector),
		Tolerations:                   tolerations,
		RestartPolicy:                 v1.RestartPolicyNever,
		TerminationGracePeriodSeconds: utils.GetInt64(300),
	}

	cronJob := &batch.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: batch.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    imLabels,
		},
		Spec: batch.CronJobSpec{
			SuccessfulJobsHistoryLimit: jobHistoryLimitSuccess,
			FailedJobsHistoryLimit:     jobHistoryLimitFailed,
			Schedule:                   schedule,
			JobTemplate: batch.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					BackoffLimit: utils.GetInt32(0),
					Parallelism:  utils.GetInt32(1),
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: namespace,
							Labels:    imLabels,
						},
						Spec: podSpec,
					},
				},
			},
		},
	}

	return cronJob
}
