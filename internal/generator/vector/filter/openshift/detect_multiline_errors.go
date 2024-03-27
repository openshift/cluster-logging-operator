package openshift

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
)

const (
	DetectMultilineException = "detectMultilineExceptions"
)

func NewDetectException(id string, inputs ...string) framework.Element {
	return normalize.DetectExceptions{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
	}
}
