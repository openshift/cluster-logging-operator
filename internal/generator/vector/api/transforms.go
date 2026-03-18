package api

import (
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	gotoml "github.com/pelletier/go-toml"
)

type Transforms map[string]types.Transform

func (tMap *Transforms) Add(id string, o types.Transform) {
	(*tMap)[id] = o
}
func (tMap *Transforms) Merge(o Transforms) {
	for id, oTransform := range o {
		(*tMap)[id] = oTransform
	}
}

func (tMap *Transforms) UnmarshalTOML(data interface{}) (err error) {
	entries, castable := data.(map[string]interface{})
	if !castable {
		return fmt.Errorf("data can not be converted to a map: %v", data)
	}
	if *tMap == nil {
		*tMap = make(Transforms)
	}
	for id, entry := range entries {
		raw, ok := entry.(map[string]interface{})
		if !ok {
			return fmt.Errorf("raw entry for transform %q can not be converted to a map %v", entry, raw)
		}
		tree, mapErr := gotoml.TreeFromMap(raw)
		if mapErr != nil {
			return errors.Join(fmt.Errorf("unable to initialize tree from transform %q", id), mapErr)
		}
		var typeExtractor struct {
			Type types.TransformType `yaml:"type" toml:"type"`
		}
		var transform types.Transform
		if err = tree.Unmarshal(&typeExtractor); err != nil {
			return errors.Join(fmt.Errorf("unable to unmarshal transform %q from %v to determine type", id, raw), err)
		}
		switch typeExtractor.Type {
		case types.TransformTypeDetectExceptions:
			var t transforms.DetectExceptions
			if err = tree.Unmarshal(&t); err != nil {
				return fmt.Errorf("failed to unmarshal transform %q: %w", id, err)
			}
			transform = &t
		case types.TransformTypeFilter:
			var s transforms.Filter
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal transform %q: %w", id, err)
			}
			transform = &s
		case types.TransformTypeLogToMetric:
			var s transforms.LogToMetric
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal transform %q: %w", id, err)
			}
			transform = &s
		case types.TransformTypeReduce:
			var s transforms.Reduce
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal transform %q: %w", id, err)
			}
			transform = &s
		case types.TransformTypeRemap:
			var s transforms.Remap
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal transform %q: %w", id, err)
			}
			transform = &s
		case types.TransformTypeRoute:
			var s transforms.Route
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal transform %q: %w", id, err)
			}
			transform = &s
		case types.TransformTypeThrottle:
			var s transforms.Throttle
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal transform %q: %w", id, err)
			}
			transform = &s
		default:
			return fmt.Errorf("unknown transform type %q for transform %q", typeExtractor.Type, id)
		}

		(*tMap)[id] = transform
	}
	return nil
}
