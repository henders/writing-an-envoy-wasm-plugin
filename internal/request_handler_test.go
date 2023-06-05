package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestRequestContext_OnHttpRequestHeaders(t *testing.T) {
	tests := []struct {
		name         string
		headers      [][2]string
		wantResponse types.Action
		wantAuthCall bool
	}{
		{
			name:         "Empty headers",
			headers:      [][2]string{},
			wantAuthCall: false,
			wantResponse: types.ActionContinue,
		},
		{
			name:         "No auth headers",
			headers:      [][2]string{{XRequestIdHeader, "abc"}},
			wantAuthCall: false,
			wantResponse: types.ActionContinue,
		},
		{
			name:         "authorization header present",
			headers:      [][2]string{{XRequestIdHeader, "abc"}, {AuthHeader, "MAC <some mac digest>"}},
			wantAuthCall: true,
			wantResponse: types.ActionPause,
		},
	}

	// Load the WASM binary and initialize a bare state for all the proxywasm APIs to work
	vmContext := InitPlugin(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize new plugin context like Envoy would do before intercepting a request
			host, contextID, reset := NewContext(t, vmContext)

			// We want to reset the state of our 'Envoy Host' after every test
			defer reset()

			// Instruct the 'Envoy Host' to call our WASM OnHttpRequestHeaders callback
			action := host.CallOnRequestHeaders(contextID, tt.headers, true)

			// Now we just validate that all the side-effects match expectations
			require.Equal(t, tt.wantResponse, action)
			require.Equal(t, tt.headers, host.GetCurrentRequestHeaders(contextID))
			if tt.wantAuthCall {
				// Verify auth service is called.
				require.Len(t, host.GetCalloutAttributesFromContext(contextID), 1)
			} else {
				require.Empty(t, host.GetCalloutAttributesFromContext(contextID))
			}
		})
	}
}
