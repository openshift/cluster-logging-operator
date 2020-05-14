package helpers

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
)

type EnvVarsExpectation struct {
	envVars []v1.EnvVar
}

type EnvVarExpectation struct {
	envVar v1.EnvVar
}

func ExpectEnvVars(envVars []v1.EnvVar) *EnvVarsExpectation {
	return &EnvVarsExpectation{envVars: envVars}
}

func (exp *EnvVarsExpectation) ToIncludeName(name string) *EnvVarExpectation {
	for _, env := range exp.envVars {
		if env.Name == name {
			return &EnvVarExpectation{env}
		}
	}
	Fail(fmt.Sprintf("Exp to find an environment variable %q in the list of env vars: %v", name, exp.envVars))
	return nil
}

func (exp *EnvVarExpectation) WithFieldRefPath(path string) *EnvVarExpectation {
	Expect(exp.envVar.ValueFrom).ToNot(BeNil(), "The valueFrom field for %v is nil", exp.envVar)
	Expect(exp.envVar.ValueFrom.FieldRef).ToNot(BeNil(), "The valueFrom.fieldRef for %v is nil", exp.envVar)
	Expect(exp.envVar.ValueFrom.FieldRef.FieldPath).To(Equal(path))
	return exp
}

func (exp *EnvVarExpectation) WithValue(value string) *EnvVarExpectation {
	Expect(exp.envVar.Value).To(Equal(value))
	return exp
}
