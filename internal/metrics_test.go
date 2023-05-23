package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestMetrics_Increment(t *testing.T) {
	vmContext := InitPlugin(t)

	t.Run("metric incremented on every request handled", func(t *testing.T) {
		// Initialize new plugin
		host, contextID, reset := NewContext(t, vmContext)
		defer reset()

		// Call OnRequestHeaders
		action := host.CallOnRequestHeaders(contextID, [][2]string{}, false)
		require.Equal(t, types.ActionContinue, action)

		value, err := host.GetCounterMetric("envoy_wasm_example_requests")
		require.NoError(t, err)
		require.Equal(t, uint64(1), value)
	})

	t.Run("re-uses same counter definition", func(t *testing.T) {
		// Initialize new plugin
		host, contextID, reset := NewContext(t, vmContext)
		defer reset()

		// Call OnRequestHeaders twice
		action := host.CallOnRequestHeaders(contextID, [][2]string{}, false)
		require.Equal(t, types.ActionContinue, action)
		action = host.CallOnRequestHeaders(contextID, [][2]string{}, false)
		require.Equal(t, types.ActionContinue, action)

		value, err := host.GetCounterMetric("envoy_wasm_example_requests")
		require.NoError(t, err)
		require.Equal(t, uint64(2), value)
	})
}
