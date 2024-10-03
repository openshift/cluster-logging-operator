package apiaudit

import (
	"bytes"
	"encoding/json"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"os"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/apiaudit"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/cmd"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	. "k8s.io/apiserver/pkg/apis/audit/v1"
	"sigs.k8s.io/yaml"
)

func TestVRLGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[functional][filters][apiaudit][vrl]")
}

var (
	// Use a single test client and pod to launch all the vector commands, big speedup vs pod-per-command.
	c   *client.Test
	pod *corev1.Pod // Initialized in TestVectorConfGenerator
)

var _ = BeforeSuite(func() {
	c = client.NewTest()
	name := "apiaudit-test"
	image := utils.GetComponentImage(constants.VectorName)
	pod = runtime.NewPodBuilder(runtime.NewPod(c.NS.Name, name)).
		AddContainer(name, image).
		AddEnvVar("VECTOR_LOG", "warn").
		WithCmd([]string{"sleep", "1h"}).End().Pod
	Expect(c.Create(pod)).To(Succeed())
	Expect(c.WaitFor(pod, client.PodRunning)).To(Succeed())
})

var _ = AfterSuite(func() {
	if c != nil {
		c.Close()
	}
})

// Helpers for the tests

func Filtered(policy *obs.KubeAPIAudit, event Event) *Event {
	fillEvent(&event)
	event.GetObjectKind().SetGroupVersionKind(runtime.GroupVersionKind(&event))
	b, err := json.Marshal(event)
	Expect(err).To(Succeed())
	out := FilteredBytes(policy, b)
	if strings.TrimSpace(string(out)) == "" {
		return nil
	}
	e2 := &Event{}
	e2.GetObjectKind().SetGroupVersionKind(runtime.GroupVersionKind(e2))
	test.Must(json.Unmarshal(out, e2))
	return e2
}

func FilteredBytes(policy *obs.KubeAPIAudit, b []byte) []byte {
	cmd := vectorCmd(policy)
	cmd.Stdin = bytes.NewReader(b)
	out, err := cmd.Output()
	test.Must(err)
	return out
}

func vectorCmd(p *obs.KubeAPIAudit) *exec.Cmd {
	vrl, err := apiaudit.NewFilter(p).VRL()
	Expect(err).NotTo(HaveOccurred(), "%#v", *p)
	conf := fmt.Sprintf(`
# Vector config for tests that read from stdin and print filtered events to stdout
[sources.in]
type = "stdin"
decoding.codec = "json"

[transforms.policy]
type = "remap"
inputs = ["in"]
source = '''
. = {"_internal": {"structured": .}}
%v
. = ._internal.structured
'''

[sinks.console]
type = "console"
inputs = ["policy"]
encoding.codec = "json"
`, vrl)
	Expect(cmd.PodWrite(pod, "", "/tmp/vector.toml", []byte(conf))).To(Succeed())
	cmd := testruntime.Exec(pod, "vector", "-c", "/tmp/vector.toml")
	cmd.Stderr = test.Writer()
	return cmd

}

func HaveLevel(level Level) types.GomegaMatcher {
	checkLevel := func(out *Event) (Level, error) {
		if out == nil {
			return LevelNone, nil
		}
		ok := (out.Level == LevelRequestResponse && out.RequestObject != nil && out.ResponseObject != nil) ||
			(out.Level == LevelRequest && out.RequestObject != nil && out.ResponseObject == nil) ||
			(out.Level == LevelMetadata && out.RequestObject == nil && out.ResponseObject == nil)
		if !ok {
			return out.Level, fmt.Errorf("request/response mismatch for level %v", out.Level)
		}
		return out.Level, nil
	}
	return WithTransform(checkLevel, Equal(level))
}

func readPolicy(path string) *obs.KubeAPIAudit {
	b, err := os.ReadFile(path)
	test.Must(err)
	policy := &obs.KubeAPIAudit{}
	test.Must(yaml.Unmarshal(b, policy))
	return policy
}

// fillEvent fills event defaults for filter tests.
func fillEvent(in *Event) {
	if in.Level == "" {
		in.Level = LevelRequestResponse
	}
	obj := &apiruntime.Unknown{Raw: []byte("{}")}
	if in.RequestObject == nil {
		in.RequestObject = obj
	}
	if in.ResponseObject == nil {
		in.ResponseObject = obj
	}
	if in.AuditID == "" {
		in.AuditID = "0000"
	}
	if in.Verb == "" {
		in.Verb = "create"
	}
}
