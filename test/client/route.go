package client

import (
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func ingressReady(ingress routev1.RouteIngress) bool {
	for _, cond := range ingress.Conditions {
		if cond.Type == routev1.RouteAdmitted && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func RouteReady(e watch.Event) (bool, error) {
	route := e.Object.(*routev1.Route)
	if len(route.Status.Ingress) == 0 {
		return false, nil
	}
	for _, ingress := range route.Status.Ingress {
		if !ingressReady(ingress) {
			return false, nil
		}
	}
	return true, nil
}
