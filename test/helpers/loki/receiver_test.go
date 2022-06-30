package loki_test

import (
	"testing"

	_ "github.com/onsi/ginkgo" // Accept ginkgo command line options
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLokiReceiverCanPushAndQueryLogs(t *testing.T) {
	c := client.ForTest(t)
	l := loki.NewReceiver(c.NS.Name, "loki")
	require.NoError(t, l.Create(c.Client))
	sv := loki.StreamValues{
		Stream: map[string]string{"test": "loki"},
		Values: loki.MakeValues([]string{"hello", "there", "mr. frog"}),
	}
	require.NoError(t, l.Push(sv))

	labels, err := l.Labels()
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"__name__", "test"}, labels)

	result, err := l.QueryUntil(`{test="loki"}`, "", 3)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, sv.Lines(), result[0].Lines())
}
