package k8shandler

import (
	"fmt"
	"io/ioutil"
	"path"
	"sync"

	"github.com/openshift/cluster-logging-operator/pkg/certificates"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

var mutex sync.Mutex

//Syncronize blocks single threads access using the certificate mutex
func Synchronize(action func() error) error {
	mutex.Lock()
	defer mutex.Unlock()
	return action()
}

func (clusterRequest *ClusterLoggingRequest) extractMasterCerts() (err error) {
	secret, err := clusterRequest.GetSecret(constants.MasterCASecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("Unable to get secret %s: %v", constants.MasterCASecretName, err)
	}
	workDir := utils.GetWorkingDir()
	for name, value := range secret.Data {
		if err != utils.WriteToWorkingDirFile(path.Join(workDir, name), value) {
			return err
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) writeSecret() (err error) {

	secrets, err := loadFilesFromWorkingDir()
	if err != nil {
		return err
	}
	secret := NewSecret(
		constants.MasterCASecretName,
		clusterRequest.Cluster.Namespace,
		secrets,
	)
	utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.Cluster))

	return clusterRequest.CreateOrUpdateSecret(secret)
}

func loadFilesFromWorkingDir() (map[string][]byte, error) {
	workDir := utils.GetWorkingDir()
	files, err := ioutil.ReadDir(workDir)
	if err != nil {
		return nil, err
	}
	results := map[string][]byte{}
	for _, f := range files {
		content := utils.GetFileContents(path.Join(workDir, f.Name()))
		if content != nil {
			results[f.Name()] = content
		} else {
			logger.Infof("The content is nil for certificate file: %s", f.Name())
		}
	}
	return results, nil
}

//CreateOrUpdateCertificates for a Cluster logging instance
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCertificates() (err error) {
	return Synchronize(func() error {
		if err = clusterRequest.extractMasterCerts(); err != nil {
			fmt.Printf("Error %v", err)
			return err
		}

		scriptsDir := utils.GetScriptsDir()
		updated := false
		if err, updated = certificates.GenerateCertificates(clusterRequest.Cluster.Namespace, scriptsDir, "elasticsearch", utils.GetWorkingDir()); err != nil {
			return fmt.Errorf("Error running script: %v", err)
		}
		logger.Tracef("Writing secret updated: %v", updated)
		if updated {
			if err = clusterRequest.writeSecret(); err != nil {
				return err
			}
		}

		return nil
	})
}
