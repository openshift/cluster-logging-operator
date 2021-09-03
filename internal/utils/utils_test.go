package utils

import (
	"os"
	"testing"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAreMapsSameWhenBothAreEmpty(t *testing.T) {
	one := map[string]string{}
	two := map[string]string{}
	if !AreMapsSame(one, two) {
		t.Error("Exp empty maps to evaluate to be equivalent")
	}
}
func TestAreMapsSameWhenOneIsEmptyAndTheOtherIsNot(t *testing.T) {
	one := map[string]string{}
	two := map[string]string{
		"foo": "bar",
	}
	if AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be different - left: %v, right: %v", one, two)
	}
}
func TestAreMapsSameWhenEquivalent(t *testing.T) {
	one := map[string]string{
		"foo": "bar",
		"xyz": "123",
	}
	two := map[string]string{
		"xyz": "123",
		"foo": "bar",
	}
	if !AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be equivalent - left: %v, right: %v", one, two)
	}
}
func TestAreMapsSameWhenDifferent(t *testing.T) {
	one := map[string]string{
		"foo": "456",
		"xyz": "123",
	}
	two := map[string]string{
		"xyz": "123",
		"foo": "bar",
	}
	if AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be different - left: %v, right: %v", one, two)
	}
}

func TestEnvVarEqualEqual(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
		{Name: "MERGE_JSON_LOG", Value: "false"},
	}

	if !EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned false for the equal inputs")
	}
}

func TestEnvVarEqualCheckValueFrom(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
	}

	if !EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned false for the equal inputs")
	}
}

func TestEnvVarEqualNotEqual(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "true"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}

	if EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned true for the not equal inputs")
	}
}

func TestEnvVarEqualShorter(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
	}

	if EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned true when the desired is shorter than the current")
	}
}

func TestEnvVarEqualNotEqual2(t *testing.T) {
	currentenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
	}
	desiredenv := []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: "ES_PORT", Value: "9200"},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
	}

	if EnvValueEqual(currentenv, desiredenv) {
		t.Errorf("EnvVarEqual returned true when the desired is longer than the current")
	}
}

var _ = Describe("GetProxyEnvVars", func() {
	var (
		envvars = map[string]string{}
	)
	BeforeEach(func() {
		for _, envvar := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "NO_PROXY", "no_proxy"} {
			envvars[envvar] = os.Getenv(envvar)
			Expect(os.Setenv(envvar, envvar)).To(Succeed())
		}
	})
	AfterEach(func() {
		for k, v := range envvars {
			Expect(os.Setenv(k, v)).To(Succeed())
		}
	})
	It("should retrieve the proxy settings from the operators ENV variables", func() {
		envvars := GetProxyEnvVars()
		Expect(envvars).To(HaveLen(6)) //proxy,noproxy vars
		for _, envvar := range envvars {
			Expect(envvar.Name).To(Equal(envvar.Value), "Exp. the value to be set to the name for the test")
		}
	})
})
