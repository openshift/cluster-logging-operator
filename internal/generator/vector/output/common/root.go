package common

import "github.com/openshift/cluster-logging-operator/internal/generator/helpers"

type RootMixin struct {
	Compression helpers.OptionalPair
}

func NewRootMixin(compression interface{}) RootMixin {
	return RootMixin{
		Compression: helpers.NewOptionalPair("compression", compression),
	}
}
