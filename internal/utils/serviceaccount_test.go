package utils

import "testing"

func TestServiceAccountUsername(t *testing.T) {
	tests := []struct {
		namespace string
		name      string
		want      string
	}{
		{
			namespace: "default",
			name:      "my-sa",
			want:      "system:serviceaccount:default:my-sa",
		},
		{
			namespace: "openshift-logging",
			name:      "cluster-logging-operator",
			want:      "system:serviceaccount:openshift-logging:cluster-logging-operator",
		},
		{
			namespace: "kube-system",
			name:      "daemon-set-controller",
			want:      "system:serviceaccount:kube-system:daemon-set-controller",
		},
	}
	for _, tt := range tests {
		t.Run(tt.namespace+"/"+tt.name, func(t *testing.T) {
			if got := ServiceAccountUsername(tt.namespace, tt.name); got != tt.want {
				t.Errorf("ServiceAccountUsername(%q, %q) = %q, want %q", tt.namespace, tt.name, got, tt.want)
			}
		})
	}
}
