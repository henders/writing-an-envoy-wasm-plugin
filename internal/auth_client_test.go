package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestAuthClient_RequestJWT(t *testing.T) {
	vmContext := InitPlugin(t)
	t.Run("call auth with auth header", func(t *testing.T) {
		// Initialize new plugin
		host, contextID, reset := NewContext(t, vmContext)
		defer reset()

		// Call OnRequestHeaders to initialize context
		_ = host.CallOnRequestHeaders(contextID, [][2]string{
			{AuthHeader, "MAC ts....."},
		}, true)

		// Verify Auth Service is called
		require.Equal(t, 1, len(host.GetCalloutAttributesFromContext(contextID)))
		// At this point, none of dispatched callouts received response therefore the current status must be paused.
		require.Equal(t, types.ActionPause, host.GetCurrentHttpStreamAction(contextID))

		// Now have mocked Auth Service respond to handler
		callout := host.GetCalloutAttributesFromContext(contextID)[0]
		host.CallOnHttpCallResponse(callout.CalloutID, [][2]string{{":status", "200"}}, nil, []byte("MY_JWT"))

		// Verify JWT was added to original request headers
		require.Equal(t, [][2]string{
			{AuthHeader, "MAC ts....."},
			{"x-auth-jwt", "MY_JWT"},
		}, host.GetCurrentRequestHeaders(contextID))
		// The request should have been marked as continued now after processing Auth response
		require.Equal(t, types.ActionContinue, host.GetCurrentHttpStreamAction(contextID))
	})
}
