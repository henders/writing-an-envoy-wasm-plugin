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

		// Verify Auth is called
		require.Equal(t, 1, len(host.GetCalloutAttributesFromContext(contextID)))
		// At this point, none of dispatched callouts received response therefore the current status must be paused.
		require.Equal(t, types.ActionPause, host.GetCurrentHttpStreamAction(contextID))

		// Now have Auth respond to handler
		callout := host.GetCalloutAttributesFromContext(contextID)[0]
		host.CallOnHttpCallResponse(callout.CalloutID, [][2]string{}, nil, nil)

		// Since we didn't return any auth headers in response above, we should only get the headers we originally sent
		require.Equal(t, [][2]string{{AuthHeader, "MAC ts....."}}, host.GetCurrentRequestHeaders(contextID))
		// The request should have been marked as continued now after processing Auth response
		require.Equal(t, types.ActionContinue, host.GetCurrentHttpStreamAction(contextID))
	})

	t.Run("sets auth request headers correctly", func(t *testing.T) {
		// Initialize new plugin
		host, contextID, reset := NewContext(t, vmContext)
		defer reset()

		// Call OnRequestHeaders to initialize context
		_ = host.CallOnRequestHeaders(contextID, [][2]string{
			{AuthHeader, "MAC ts....."},
			{"x-request-id", "abc"},
			{":method", "POST"},
			{":path", "/api/v2/foobar"},
		}, true)

		require.Equal(t, 1, len(host.GetCalloutAttributesFromContext(contextID)))
		require.Equal(t, types.ActionPause, host.GetCurrentHttpStreamAction(contextID))

		callout := host.GetCalloutAttributesFromContext(contextID)[0]
		require.Equal(t, [][2]string{
			{AuthHeader, "MAC ts....."},
			{"x-request-id", "abc"},
			{":method", "POST"},
			{":path", "/api/v2/foobar"},
		}, callout.Headers)
	})

	t.Run("verify auth response headers are appended to original request", func(t *testing.T) {
		// Initialize new plugin
		host, contextID, reset := NewContext(t, vmContext)
		defer reset()

		// Call OnRequestHeaders to initialize context
		_ = host.CallOnRequestHeaders(contextID, [][2]string{
			{AuthHeader, "MAC ts....."},
			{":method", "GET"},
			{":path", "/api/v2/foobar"},
		}, true)

		// Verify Auth is called
		require.Equal(t, 1, len(host.GetCalloutAttributesFromContext(contextID)))
		// At this point, none of dispatched callouts received response therefore the current status must be paused.
		require.Equal(t, types.ActionPause, host.GetCurrentHttpStreamAction(contextID))

		// Now have Auth respond to handler
		callout := host.GetCalloutAttributesFromContext(contextID)[0]
		host.CallOnHttpCallResponse(callout.CalloutID, [][2]string{
			{XAuthServiceAuthResponseJWTHeader, "ewoJCSJhbGciOiAiSFMyNTYiLAoJCSJ0eXAiOiAiSldUIgoJfQ.eyJleHAiOjE2NDA5OTUyMjAsImlhdCI6MTY0MDk5NTIwMCwic3lzdGVtX3VzZXIiOnsibmFtZSI6Ii9ucy9jbGFzc2ljL3NhL2RlZmF1bHQifSwidmlhIjoic3lzdGVtX3VzZXIifQ.H3wlqQRWnFjfADtpRLbE1BqK3PLY7HNPtOwJgpTuofM"},
		}, nil, nil)

		// Verify headers were copied onto original request
		require.Equal(t, [][2]string{
			{AuthHeader, "MAC ts....."},
			{":method", "GET"},
			{":path", "/api/v2/foobar"},
			{XAuthServiceAuthResponseJWTHeader, "ewoJCSJhbGciOiAiSFMyNTYiLAoJCSJ0eXAiOiAiSldUIgoJfQ.eyJleHAiOjE2NDA5OTUyMjAsImlhdCI6MTY0MDk5NTIwMCwic3lzdGVtX3VzZXIiOnsibmFtZSI6Ii9ucy9jbGFzc2ljL3NhL2RlZmF1bHQifSwidmlhIjoic3lzdGVtX3VzZXIifQ.H3wlqQRWnFjfADtpRLbE1BqK3PLY7HNPtOwJgpTuofM"},
		}, host.GetCurrentRequestHeaders(contextID))
		// The request should have been marked as continued now after processing Auth response
		require.Equal(t, types.ActionContinue, host.GetCurrentHttpStreamAction(contextID))
	})
}
