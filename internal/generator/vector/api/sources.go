package api

import (
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	gotoml "github.com/pelletier/go-toml"
)

// Sources is a set of sources by their id.
type Sources map[string]types.Source

func (sourceMap *Sources) Add(id string, source types.Source) {
	(*sourceMap)[id] = source
}

func (sourceMap *Sources) UnmarshalTOML(data interface{}) (err error) {
	entries, castable := data.(map[string]interface{})
	if !castable {
		return fmt.Errorf("sources data can not be converted to a map: %v", data)
	}
	if *sourceMap == nil {
		*sourceMap = make(Sources)
	}
	for id, entry := range entries {
		rawSource, ok := entry.(map[string]interface{})
		if !ok {
			return fmt.Errorf("raw entry for source %q can not be converted to a map %v", entry, rawSource)
		}
		tree, mapErr := gotoml.TreeFromMap(rawSource)
		if mapErr != nil {
			return errors.Join(fmt.Errorf("unable to initialize tree from source %q", id), mapErr)
		}
		var typeExtractor struct {
			Type types.SourceType `yaml:"type"`
		}
		var source types.Source
		if err = tree.Unmarshal(&typeExtractor); err != nil {
			return errors.Join(fmt.Errorf("unable to unmarshal source %q from %v to determine type", id, rawSource), err)
		}
		switch typeExtractor.Type {
		case types.SourceTypeFile:
			var s sources.File
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal file source %s: %w", id, err)
			}
			source = &s
		case types.SourceTypeHttpServer:
			var s sources.HttpServer
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal http_server source %s: %w", id, err)
			}
			source = &s
		case types.SourceTypeInternalMetrics:
			var s sources.InternalMetrics
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal internal_metrics source %s: %w", id, err)
			}
			source = &s
		case types.SourceTypeKubernetesLogs:
			var s sources.KubernetesLogs
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal kubernetes_logs source %s: %w", id, err)
			}
			source = &s
		case types.SourceTypeJournald:
			var s sources.Journald
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal journald source %s: %w", id, err)
			}
			source = &s
		case types.SourceTypeSyslog:
			var s sources.Syslog
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal syslog source %s: %w", id, err)
			}
			source = &s
		default:
			return fmt.Errorf("unknown source type %s for source %s", typeExtractor.Type, id)
		}

		(*sourceMap)[id] = source
	}
	return nil
}
