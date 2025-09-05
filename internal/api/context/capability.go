package context

import (
	"errors"
	"os"

	log "github.com/ViaQ/logerr/v2/log/static"
	utilsjson "github.com/openshift/cluster-logging-operator/internal/utils/json"
)

type Capability struct {
	Enabled bool `json:"enabled,omitempty"`
}

type Capabilities map[string]Capability

func (o Capabilities) IsEnabled(key string) bool {
	if f, found := o[key]; found {
		return f.Enabled
	}
	return false
}

func ReadCapabilities(path string) Capabilities {
	if content, err := os.ReadFile(path); err != nil {
		features := &Capabilities{}
		if err = utilsjson.Unmarshal(string(content), features); err != nil {
			log.V(0).Error(err, "Unable to unmarshall enabled features")
		}
		return *features
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			log.V(0).Error(err, "Unable to read enabled features")
		}
	}
	return Capabilities{}
}
