package functional

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"k8s.io/apimachinery/pkg/util/wait"
)

// CollectMetricLines polls the collector's Prometheus endpoint until a line
// matching both metricName and waitFor is found, then returns all lines
// matching metricName.
func (f *CollectorFunctionalFramework) CollectMetricLines(metricName, waitFor string, timeout time.Duration) ([]string, error) {
	var matched []string
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		raw, err := f.RunCommand(constants.CollectorName, "curl", "-ks",
			fmt.Sprintf("https://%s.%s:24231/metrics", f.Name, f.Namespace))
		if err != nil {
			return false, nil
		}
		matched = nil
		for _, line := range strings.Split(raw, "\n") {
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.Contains(line, metricName) {
				matched = append(matched, line)
			}
		}
		for _, line := range matched {
			if strings.Contains(line, waitFor) {
				return true, nil
			}
		}
		return false, nil
	})
	return matched, err
}
