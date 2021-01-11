package k8shandler

func (clusterRequest *ClusterLoggingRequest) removeCurator() (err error) {
	if clusterRequest.isManaged() {
		if err = clusterRequest.RemoveSecret("curator"); err != nil {
			return
		}

		if err = clusterRequest.RemoveCronJob("curator"); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap("curator"); err != nil {
			return
		}

		if err = clusterRequest.RemoveServiceAccount("curator"); err != nil {
			return
		}
	}

	return nil
}
