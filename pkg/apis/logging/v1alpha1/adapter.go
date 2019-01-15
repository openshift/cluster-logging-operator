package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func (cluster *ClusterLogging) Type() metav1.TypeMeta {
	return cluster.TypeMeta
}
func (cluster *ClusterLogging) Meta() metav1.ObjectMeta {
	return cluster.ObjectMeta
}

func (cluster *ClusterLogging) SchemeGroupVersion() string {
	return SchemeGroupVersion.String()
}
