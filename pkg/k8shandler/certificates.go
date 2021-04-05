package k8shandler

import (
	"fmt"
	"io/ioutil"
	"path"
	"sync"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/certificates"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

var mutex sync.Mutex

//Syncronize blocks single threads access using the certificate mutex
func Syncronize(action func() error) error {
	mutex.Lock()
	defer mutex.Unlock()
	return action()
}

func (clusterRequest *ClusterLoggingRequest) extractMasterCerts() (extracted bool, err error) {
	secret, err := clusterRequest.GetSecret(constants.MasterCASecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("Unable to get secret %s: %v", constants.MasterCASecretName, err)
	}

	log.V(3).Info("Master cert", "size", len(secret.Data))
	for name, value := range secret.Data {
		log.V(3).Info("Master cert", "file", name)
		if err = utils.WriteToWorkingDirFile(name, value); err != nil {
			log.Error(err, "Error extracting cert from master-cert secret", "file", name)
			return false, err
		}
	}

	return true, nil
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
			log.V(0).Info("The content is nil for certificate file", "file", f.Name())
		}
	}
	return results, nil
}

//CreateOrUpdateCertificates for a cluster logging instance
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCertificates() (err error) {
	return Syncronize(func() error {
		var extracted bool
		if extracted, err = clusterRequest.extractMasterCerts(); err != nil {
			log.Error(err, "Error extracting master-cert")
			return err
		}

		scriptsDir := utils.GetScriptsDir()
		updated := false
		var result []interface{}
		if err, updated, result = certificates.GenerateCertificates(clusterRequest.Cluster.Namespace, scriptsDir, "elasticsearch", utils.GetWorkingDir()); err != nil {
			return fmt.Errorf("Error running script: %v", err)
		}
		log.V(3).Info("Writing secret", "updated", updated)
		if !extracted || updated {
			if err = clusterRequest.writeSecret(); err != nil {
				return err
			}
		}
		for _, r := range result {
			strs := r.([]interface{})
			for _, s := range strs {
				if s == "reason" {
					clusterRequest.EventRecorder.Event(clusterRequest.Cluster, "Normal", "Updated", fmt.Sprint(strs))
				}
			}
		}

		return nil
	})
}
