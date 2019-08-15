package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	consolev1 "github.com/openshift/api/console/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewConsoleExternalLogLink stubs an instance of a ConsoleExternalLogLink
func NewConsoleExternalLogLink(resourceName, namespace, consoleText, hrefTemplate string) *consolev1.ConsoleExternalLogLink {
	return &consolev1.ConsoleExternalLogLink{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConsoleExternalLogLink",
			APIVersion: consolev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: namespace,
			Labels: map[string]string{
				"component":     "support",
				"logging-infra": "support",
				"provider":      "openshift",
			},
		},
		Spec: consolev1.ConsoleExternalLogLinkSpec{
			Text:         consoleText,
			HrefTemplate: hrefTemplate,
		},
	}
}

//RemoveConsoleExternalLogLink with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveConsoleExternalLogLink(resourceName string) (err error) {
	consoleExternalLogLink := NewConsoleExternalLogLink(
		resourceName,
		clusterRequest.cluster.Namespace,
		"",
		"",
	)

	err = clusterRequest.Delete(consoleExternalLogLink)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v ConsoleExternalLogLink %v", resourceName, err)
	}
	return nil
}
