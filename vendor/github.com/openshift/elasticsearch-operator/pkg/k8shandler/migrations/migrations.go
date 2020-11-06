package migrations

import (
	"fmt"

	"github.com/openshift/elasticsearch-operator/pkg/elasticsearch"
	estypes "github.com/openshift/elasticsearch-operator/pkg/types/elasticsearch"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MigrationRequest interface {
	RunKibanaMigrations() error
	RunElasticsearchMigrations() error
}

func NewMigrationRequest(client client.Client, esClient elasticsearch.Client) MigrationRequest {
	return &migrationRequest{
		client:   client,
		esClient: esClient,
	}
}

type migrationRequest struct {
	client   client.Client
	esClient elasticsearch.Client
}

func (mr *migrationRequest) RunKibanaMigrations() error {
	logrus.Debugf("running Kibana migrations")
	if index, _ := mr.esClient.GetIndex(kibanaIndex); index == nil {
		logrus.Debugf("skipping kibana migrations: no index %q available", kibanaIndex)
		return nil
	}

	indices, err := mr.esClient.GetAllIndices(kibanaIndex)
	if err != nil {
		return fmt.Errorf("failed to get `.kibana` index health before running migrations: %s", err)
	}

	health, err := getIndexHealth(indices, kibanaIndex)
	if err != nil {
		return fmt.Errorf("failed to get `.kibana` index health before running migrations: %s", err)
	}

	if health != "green" && health != "yellow" {
		return fmt.Errorf("waiting for `.kibana` index recovery before running migrations: %s / (green,yellow)", health)
	}

	logrus.Debugf("attempting re-indexing `.kibana` into `.kibana-6`")
	if err := mr.reIndexKibana5to6(); err != nil {
		return fmt.Errorf("failed re-indexing `.kibana` into `.kibana-6`: %s", err)
	}
	return nil
}

func (mr *migrationRequest) RunElasticsearchMigrations() error {
	logrus.Debugf("running elasticsearch migrations")
	return nil
}

func (mr *migrationRequest) matchRequiredMajorVersion(version string) (bool, error) {
	versions, err := mr.esClient.GetClusterNodeVersions()
	if err != nil {
		return false, err
	}

	if versions == nil {
		return false, nil
	}

	if len(versions) > 1 {
		return false, nil
	}

	return (utils.GetMajorVersion(versions[0]) == version), nil
}

func getIndexHealth(indices estypes.CatIndicesResponses, name string) (string, error) {
	if len(indices) == 0 {
		return "unknown", fmt.Errorf("failed to get index health for %q ", name)
	}

	return indices[0].Health, nil
}
