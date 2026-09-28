package utils

import "fmt"

// ServiceAccountUsername returns the authenticated user identity string for a
// ServiceAccount, in the format used by Kubernetes authentication and RBAC.
// This is the canonical "system:serviceaccount:<namespace>:<name>" format.
func ServiceAccountUsername(namespace, name string) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", namespace, name)
}
