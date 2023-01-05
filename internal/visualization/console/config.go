package console

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Name = "logging-view-plugin"

// Config is the configuration struct for the logging console plugin Reconciler.
// Construct with NewConfig to set default values.
type Config struct {
	Owner       client.Object // Owning object - the ClusterLogging instance
	Name        string        // Name for the consoleplugin and related objects
	Image       string        // Image for the logging view service.
	LokiService string        // Name of the LokiStack gateway service.
	LokiPort    int32         // Port of the LokiStack gateway service.
	Features    []string      // The features enabled for the plugin
}

func (cf *Config) Namespace() string        { return cf.Owner.GetNamespace() }
func (cf *Config) CreatedBy() string        { return fmt.Sprintf("%v_%v", cf.Namespace(), cf.Owner.GetName()) }
func (cf *Config) defaultMode() *int32      { return utils.GetInt32(420) }
func (cf *Config) pluginBackendPort() int32 { return 9443 }

// NewConfig returns a config with default settings.
func NewConfig(owner client.Object, lokiService string, features []string) Config {
	return Config{
		Owner:       owner,
		Name:        Name,
		Image:       utils.GetComponentImage(constants.ConsolePluginName),
		LokiService: lokiService,
		LokiPort:    8080,
		Features:    features,
	}
}
