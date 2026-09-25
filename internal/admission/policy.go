package admission

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

var AdmissionReconcileBackoff = wait.Backoff{
	Steps:    5,
	Duration: 2 * time.Second,
	Factor:   2.0,
	Cap:      30 * time.Second,
}
