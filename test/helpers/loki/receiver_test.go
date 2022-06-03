package loki_test

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLokiReceiver(t *testing.T) {
	c := client.ForTest(t)
	l := loki.NewReceiver(c.NS.Name, "loki").EnableMultiTenant()
	require.NoError(t, l.Create(c.Client))
	sv := loki.StreamValues{
		Stream: map[string]string{"test": "loki"},
		Values: loki.MakeValues([]string{"hello", "there", "mr. frog"}),
	}
	require.NoError(t, l.Push("tenant", sv))

	t.Run("canPushAndQuery", func(t *testing.T) {
		labels, err := l.Labels("tenant")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"__name__", "test"}, labels)

		result, err := l.QueryUntil(`{test="loki"}`, "tenant", 3)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, sv.Lines(), result[0].Lines())
	})

	t.Run("respectsTenancy", func(t *testing.T) {
		result, err := l.QueryUntil(`{test="loki"}`, "tenant", 1)
		require.NoError(t, err)
		assert.Equal(t, []string{"hello"}, result[0].Lines())

		// Different tenat should not get any logs.
		result, err = l.Query(`{test="loki"}`, "nottenant", 1)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
