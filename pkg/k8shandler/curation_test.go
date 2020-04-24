package k8shandler

import (
	"reflect"
	"testing"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewCuratorCronJobWhenFieldsAreUndefined(t *testing.T) {

	cluster := &logging.ClusterLogging{}
	cronJob := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")
	podSpec := cronJob.Spec.JobTemplate.Spec.Template.Spec

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 container")
	}

	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != defaultCuratorMemory {
		t.Errorf("Exp. the default memory limit to be %v", defaultCuratorMemory)
	}
	if resources.Requests[v1.ResourceMemory] != defaultCuratorMemory {
		t.Errorf("Exp. the default memory request to be %v", defaultCuratorMemory)
	}
	if resources.Requests[v1.ResourceCPU] != defaultFluentdCpuRequest {
		t.Errorf("Exp. the default CPU request to be %v", defaultCuratorCpuRequest)
	}
	// check node selecor
	if podSpec.NodeSelector == nil {
		t.Errorf("Exp. the nodeSelector to contains the linux allocation selector but was %T", podSpec.NodeSelector)
	}
}

func TestNewCuratorCronJobWhenResourcesAreDefined(t *testing.T) {
	limitMemory := resource.MustParse("100Gi")
	requestMemory := resource.MustParse("120Gi")
	requestCPU := resource.MustParse("500m")
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Curation: &logging.CurationSpec{
				Type: "curator",
				CuratorSpec: logging.CuratorSpec{
					Resources: newResourceRequirements("100Gi", "", "120Gi", "500m"),
				},
			},
		},
	}
	cronJob := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")
	podSpec := cronJob.Spec.JobTemplate.Spec.Template.Spec

	if len(podSpec.Containers) != 1 {
		t.Error("Exp. there to be 1 curator container")
	}

	resources := podSpec.Containers[0].Resources
	if resources.Limits[v1.ResourceMemory] != limitMemory {
		t.Errorf("Exp. the spec memory limit to be %v", limitMemory)
	}
	if resources.Requests[v1.ResourceMemory] != requestMemory {
		t.Errorf("Exp. the spec memory request to be %v", requestMemory)
	}
	if resources.Requests[v1.ResourceCPU] != requestCPU {
		t.Errorf("Exp. the spec CPU request to be %v", requestCPU)
	}
}

func TestNewCuratorCronJobWhenNoScheduleDefined(t *testing.T) {

	defaultSchedule := "30 3,9,15,21 * * *"

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Curation: &logging.CurationSpec{
				Type:        "curator",
				CuratorSpec: logging.CuratorSpec{},
			},
		},
	}

	cronJob := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")

	schedule := cronJob.Spec.Schedule

	if schedule != defaultSchedule {
		t.Errorf("Exp. the cronjob schedule to be: %v, act. is: %v", defaultSchedule, schedule)
	}
}

func TestNewCuratorCronJobWhenScheduleDefined(t *testing.T) {

	desiredSchedule := "30 */4 * * *"

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Curation: &logging.CurationSpec{
				Type: "curator",
				CuratorSpec: logging.CuratorSpec{
					Schedule: desiredSchedule,
				},
			},
		},
	}

	cronJob := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")

	schedule := cronJob.Spec.Schedule

	if schedule != desiredSchedule {
		t.Errorf("Exp. the cronjob schedule to be: %v, act. is: %v", desiredSchedule, schedule)
	}
}
func TestNewCuratorCronJobWhenNodeSelectorDefined(t *testing.T) {
	expSelector := map[string]string{
		"foo":             "bar",
		utils.OsNodeLabel: utils.LinuxValue,
	}
	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Curation: &logging.CurationSpec{
				Type: "curator",
				CuratorSpec: logging.CuratorSpec{
					NodeSelector: expSelector,
				},
			},
		},
	}

	job := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")
	selector := job.Spec.JobTemplate.Spec.Template.Spec.NodeSelector

	if !reflect.DeepEqual(selector, expSelector) {
		t.Errorf("Exp. the nodeSelector to be %q but was %q", expSelector, selector)
	}
}

func TestNewCuratorNoTolerations(t *testing.T) {
	expTolerations := []v1.Toleration{}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Curation: &logging.CurationSpec{
				Type:        "curator",
				CuratorSpec: logging.CuratorSpec{},
			},
		},
	}

	job := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")
	tolerations := job.Spec.JobTemplate.Spec.Template.Spec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}

func TestNewCuratorWithTolerations(t *testing.T) {
	expTolerations := []v1.Toleration{
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}

	cluster := &logging.ClusterLogging{
		Spec: logging.ClusterLoggingSpec{
			Curation: &logging.CurationSpec{
				Type: "curator",
				CuratorSpec: logging.CuratorSpec{
					Tolerations: expTolerations,
				},
			},
		},
	}

	job := newCuratorCronJob(cluster, "test-app-name", "elasticsearch")
	tolerations := job.Spec.JobTemplate.Spec.Template.Spec.Tolerations

	if !utils.AreTolerationsSame(tolerations, expTolerations) {
		t.Errorf("Exp. the tolerations to be %v but was %v", expTolerations, tolerations)
	}
}
