package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetrics_Increment(t *testing.T) {
	vmContext := InitPlugin(t)

	t.Run("metric incremented on every request handled", func(t *testing.T) {
		// Initialize new plugin
		host, contextID, reset := NewContext(t, vmContext)
		defer reset()

		// Call OnRequestHeaders twice
		_ = host.CallOnRequestHeaders(contextID, [][2]string{}, false)
		_ = host.CallOnRequestHeaders(contextID, [][2]string{}, false)

		value, err := host.GetCounterMetric("envoy_wasm_auth_plugin_requests_intercepted_destination_namespace=.=example-service;.;")
		require.NoError(t, err)
		// Validate the value of the metric
		require.Equal(t, uint64(2), value)
	})
}
