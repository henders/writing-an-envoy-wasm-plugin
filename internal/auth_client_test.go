package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestAuthClient_RequestJWT(t *testing.T) {
	// Initialize new plugin with http context
	host, contextID, reset := NewContext(t, InitPlugin(t))
	defer reset()

	// Call OnRequestHeaders to initialize context
	_ = host.CallOnRequestHeaders(contextID, [][2]string{{AuthHeader, "MAC ts....."}}, true)

	// Verify Auth Service is called with DispatchHttpCall by checking the length of the callout array
	require.Equal(t, 1, len(host.GetCalloutAttributesFromContext(contextID)))

	// At this point, none of dispatched callouts received response therefore the current status must be paused.
	require.Equal(t, types.ActionPause, host.GetCurrentHttpStreamAction(contextID))

	// Get the handle to the DispatchHttpCall request
	callout := host.GetCalloutAttributesFromContext(contextID)[0]

	// Validate we sent the Auth Service the right headers
	require.Equal(t, [][2]string{
		{"accept", "*/*"},
		{":authority", "auth"},
		{":method", "GET"},
		{":path", "/base64/RkFLRV9KV1QK"},
		{AuthHeader, "MAC ts....."},
	}, callout.Headers)

	// Now have a pretend Auth Service respond to our callback handler. Note we can specify any response headers or bodies.
	host.CallOnHttpCallResponse(callout.CalloutID, [][2]string{{":status", "200"}}, nil, []byte("test JWT"))

	// Verify the JWT from the above request was added to the original request's headers
	require.Equal(t, [][2]string{
		{AuthHeader, "MAC ts....."},
		{"x-auth-jwt", "test JWT"},
	}, host.GetCurrentRequestHeaders(contextID))

	// The request should now have been marked as continued after processing Auth response
	require.Equal(t, types.ActionContinue, host.GetCurrentHttpStreamAction(contextID))
}
